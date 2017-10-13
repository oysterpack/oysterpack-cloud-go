// Copyright (c) 2017 OysterPack, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package server

import (
	"net"

	"crypto/tls"
	"fmt"
	"time"

	"net/url"
	"strings"

	"sync"

	natsserver "github.com/nats-io/gnatsd/server"
	"github.com/nats-io/prometheus-nats-exporter/exporter"
	"github.com/oysterpack/oysterpack.go/pkg/messaging"
	"github.com/rs/zerolog"
)

// Default config settings
const (
	DEFAULT_SERVER_PORT                   = natsserver.DEFAULT_PORT
	DEFAULT_CLUSTER_PORT                  = natsserver.DEFAULT_PORT + 1000
	DEFAULT_MONITOR_PORT                  = natsserver.DEFAULT_HTTP_PORT
	DEFAULT_MAXPAYLOAD                    = 1024 * 100 // 100 KB
	DEFAULT_PROMETHEUS_EXPORTER_HTTP_PORT = 4444
)

// NewNATSServer creates a new NATSServer
func NewNATSServer(config *NATSServerConfig) (NATSServer, error) {
	opts, err := config.ServerOpts()
	if err != nil {
		return nil, err
	}

	s := natsserver.New(opts)
	if config.Logger == nil {
		config.Logger = &logger
	}
	s.SetLogger(NewNATSLogger(config.Logger), config.DebugLogEnabled(), config.TraceLogEnabled())

	if config.MetricsExporterPort <= 0 {
		config.MetricsExporterPort = DEFAULT_PROMETHEUS_EXPORTER_HTTP_PORT
	}

	return &natsServer{Server: s, exporterPort: config.MetricsExporterPort, cluster: config.Cluster}, nil
}

// NATSServer is the Service interface for a NATS server
type NATSServer interface {
	Cluster() messaging.ClusterName

	// Addr will return the net.Addr object for the current listener.
	Addr() net.Addr

	// ClusterAddr returns the net.Addr object for the route listener.
	ClusterAddr() *net.TCPAddr

	// MonitorAddr will return the net.Addr object for the monitoring listener.
	MonitorAddr() *net.TCPAddr

	// PrometheusHTTPExportPort returns the HTTP port that the prometheus metrics are exported on
	PrometheusHTTPExportPort() int

	// NumClients will report the number of registered clients.
	NumClients() int

	// NumRemotes will report number of registered remotes.
	NumRemotes() int

	// NumRoutes will report the number of registered routes.
	NumRoutes() int

	// NumSubscriptions will report how many subscriptions are active.
	NumSubscriptions() uint32

	// Start up the server, this will block. Start via a Go routine if needed.
	Start()

	// StartMonitoring starts the HTTP server if needed.
	StartMonitoring() error

	// Shutdown will shutdown the server instance by kicking out the AcceptLoop and closing all associated clients.
	Shutdown()

	// ReadyForConnections returns `true` if the server is ready to accept client and, if routing is enabled, route connections.
	// If after the duration `dur` the server is still not ready, returns `false`.
	ReadyForConnections(dur time.Duration) bool
}

type natsServer struct {
	sync.Mutex

	cluster messaging.ClusterName

	*natsserver.Server
	logger *zerolog.Logger

	exporterPort int
	*exporter.NATSExporter
}

func (a *natsServer) Cluster() messaging.ClusterName {
	return a.cluster
}

func (a *natsServer) PrometheusHTTPExportPort() int {
	return a.exporterPort
}

// Start starts the server and monitoring.
// Monitoring is started automatically when the server is started because a monitoring port has been configured.
// The NATSExporter is also started, which collects metrics from the embedded HTTP monitor.
func (a *natsServer) StartUp() error {
	a.Lock()
	defer a.Unlock()

	go a.Server.Start()

	for {
		if a.ReadyForConnections(10 * time.Second) {
			break
		}
		logger.Warn().Msg("Waiting for the server to start up ...")
	}

	if err := a.startExporter(); err != nil {
		logger.Error().Err(err).Msg("NATSExporter failed to start up. Server will be shutdown.")
		a.NATSExporter = nil
		a.Server.Shutdown()
		return err
	}

	return nil
}

func (a *natsServer) startExporter() error {
	exporterOpts := exporter.GetDefaultExporterOptions()
	exporterOpts.ListenPort = a.exporterPort
	exporterOpts.GetConnz = true
	exporterOpts.GetRoutez = true
	exporterOpts.GetSubz = true
	exporterOpts.GetVarz = true
	logger.Info().Str("MonitorURL", exporterOpts.NATSServerURL).Msg("")
	a.NATSExporter = exporter.NewExporter(exporterOpts)

	monitorAddr := a.MonitorAddr()
	a.NATSExporter.AddServer(a.ID(), fmt.Sprintf("http://%s", monitorAddr))
	return a.NATSExporter.Start()
}

func (a *natsServer) Start() {
	if err := a.StartUp(); err != nil {
		logger.Panic().Err(err).Msg("Server startup failed.")
	}
}

// Shutdown stops the NATSExporter first, and then stops the NATS server.
func (a *natsServer) Shutdown() {
	a.Lock()
	defer a.Unlock()

	if a.NATSExporter != nil {
		a.NATSExporter.Stop()
		a.NATSExporter = nil
	}
	a.Server.Shutdown()
}

// NATSServerConfig config used to creata a new NATSServer
type NATSServerConfig struct {
	Cluster messaging.ClusterName

	// OPTIONAL
	ServerHost string
	// OPTIONAL - DEFAULT_SERVER_PORT
	ServerPort int

	// OPTIONAL
	MonitorHost string
	// OPTIONAL - DEFAULT_MONITOR_PORT
	MonitorPort int

	// OPTIONAL
	ClusterHost string
	// OPTIONAL - DEFAULT_CLUSTER_PORT
	ClusterPort int
	// OPTIONAL - if not specified, then it adds itself as a route - assumming that this is a cluster seed node.
	// However, this should normally be set to point to seed nodes.
	Routes []string

	// applies to ServerPort, i.e., client-server connections
	TLSConfig *tls.Config

	// applies to ClusterPort, i.e., server-server connections within the NATS cluster
	ClusterTLSConfig *tls.Config

	// Configures NATS server logging
	LogLevel NATSLogLevel

	// OPTIONAL
	MaxPayload int
	// OPTIONAL
	MaxConn int

	// OPTIONAL - Prometheus metrics exporter HTTP port - DEFAULT_PROMETHEUS_EXPORTER_HTTP_PORT
	MetricsExporterPort int

	Logger *zerolog.Logger
}

// ServerOpts converts the config to a ServerOpts that can be used to create a new nats.Server.
// It ensures the server is started up with the HTTP monitor and configured to join a cluster.
// If no routes are specified, then it adds itself as a route - assuming that this is a cluster seed node.
func (a *NATSServerConfig) ServerOpts() (*natsserver.Options, error) {
	if err := a.Cluster.Validate(); err != nil {
		return nil, err
	}

	if err := a.checkPorts(); err != nil {
		return nil, err
	}

	if a.MaxPayload <= 0 {
		a.MaxPayload = DEFAULT_MAXPAYLOAD
	}

	routes, err := a.routes()
	if err != nil {
		return nil, err
	}

	a.checkHosts()

	if a.MaxConn <= 0 {
		a.MaxConn = natsserver.DEFAULT_MAX_CONNECTIONS
	}

	opts := &natsserver.Options{
		NoSigs:    true,
		Host:      a.ServerHost,
		Port:      a.ServerPort,
		TLSConfig: a.TLSConfig,
		Cluster:   natsserver.ClusterOpts{Host: a.ClusterHost, Port: a.ClusterPort, TLSConfig: a.ClusterTLSConfig},
		Routes:    routes,

		HTTPHost: a.MonitorHost,
		HTTPPort: a.MonitorPort,

		MaxPayload: a.MaxPayload,
		MaxConn:    a.MaxConn,

		PingInterval:   natsserver.DEFAULT_PING_INTERVAL,
		MaxPingsOut:    natsserver.DEFAULT_PING_MAX_OUT,
		MaxControlLine: natsserver.MAX_CONTROL_LINE_SIZE,
		WriteDeadline:  natsserver.DEFAULT_FLUSH_DEADLINE,
	}

	if err := a.setLogLevel(opts); err != nil {
		return nil, err
	}

	return opts, nil
}

func (a *NATSServerConfig) setLogLevel(opts *natsserver.Options) error {
	switch a.LogLevel {
	case NOLOG:
		opts.NoLog = true
	case DEBUG:
		opts.Debug = true
	case TRACE:
		opts.Trace = true
	default:
		return fmt.Errorf("Invalid LogLevel : %d", a.LogLevel)
	}
	return nil
}

func (a *NATSServerConfig) checkHosts() {
	a.ServerHost = strings.TrimSpace(a.ServerHost)
	if a.ServerHost == "" {
		a.ServerHost = natsserver.DEFAULT_HOST
	}
	a.ClusterHost = strings.TrimSpace(a.ClusterHost)
	if a.ClusterHost == "" {
		a.ClusterHost = natsserver.DEFAULT_HOST
	}
	a.MonitorHost = strings.TrimSpace(a.MonitorHost)
	if a.MonitorHost == "" {
		a.MonitorHost = natsserver.DEFAULT_HOST
	}
}

func (a *NATSServerConfig) routes() ([]*url.URL, error) {
	routes := make([]*url.URL, len(a.Routes))
	for i := 0; i < len(routes); i++ {
		route, err := url.Parse(a.Routes[i])
		if err != nil {
			return nil, fmt.Errorf("Invalid route URL : %s : %v", a.Routes[i], err)
		}
		routes[i] = route
	}

	if len(routes) == 0 {
		// add a self route
		protocol := "tls"
		if a.ClusterTLSConfig == nil {
			protocol = "nats"
		}
		route, err := url.Parse(fmt.Sprintf("%s://localhost:%d", protocol, a.ClusterPort))
		if err != nil {
			// should never happen
			return nil, fmt.Errorf("Invalid self route URL : %v", err)
		}
		routes = append(routes, route)
	}
	return routes, nil
}

func (a *NATSServerConfig) checkPorts() error {
	if a.ServerPort <= 0 {
		a.ServerPort = DEFAULT_SERVER_PORT
	}
	if a.MonitorPort <= 0 {
		a.MonitorPort = DEFAULT_MONITOR_PORT
	}
	if a.ClusterPort <= 0 {
		a.ClusterPort = DEFAULT_CLUSTER_PORT
	}
	ports := map[int]int{a.ServerPort: a.ServerPort, a.ClusterPort: a.ClusterPort, a.MonitorPort: a.MonitorPort}
	if len(ports) != 3 {
		return fmt.Errorf("There are port collisions : ServerPort (%d) MonitorPort(%d) ClusterPort(%d)", a.ServerPort, a.MonitorPort, a.ClusterPort)
	}
	return nil
}

// DebugLogEnabled true if debug logging is enabled
func (a *NATSServerConfig) DebugLogEnabled() bool {
	return a.LogLevel == DEBUG
}

// TraceLogEnabled true if trace logging is enabled
func (a *NATSServerConfig) TraceLogEnabled() bool {
	return a.LogLevel == TRACE
}

// NATSLogLevel enum for log level setting
type NATSLogLevel int

// NATSLogLevel enum values
const (
	NOLOG NATSLogLevel = iota
	DEBUG
	TRACE
)
