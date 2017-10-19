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

	"fmt"
	"time"

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
