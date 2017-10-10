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
	"errors"
	"fmt"
	"time"

	"net/url"
	"strings"

	natsserver "github.com/nats-io/gnatsd/server"
	"github.com/oysterpack/oysterpack.go/pkg/messaging"
)

// Default config settings
const (
	DEFAULT_SERVER_PORT  = natsserver.DEFAULT_PORT
	DEFAULT_CLUSTER_PORT = natsserver.DEFAULT_PORT + 1000
	DEFAULT_MONITOR_PORT = natsserver.DEFAULT_HTTP_PORT
	DEFAULT_MAXPAYLOAD   = 1024 * 100 // 100 KB
)

// NewNATSServer creates a new NATSServer
func NewNATSServer(config *NATSServerConfig) (NATSServer, error) {
	opts, err := config.ServerOpts()
	if err != nil {
		return nil, err
	}

	s := natsserver.New(opts)
	if s == nil {
		return nil, errors.New("No NATS Server object was returned.")
	}
	s.SetLogger(natsServerLogger, config.DebugLogEnabled(), config.TraceLogEnabled())
	return &natsServer{s}, nil
}

// TODO: metrics, healthchecks

// NATSServer is the Service interface for a NATS server
type NATSServer interface {
	// Addr will return the net.Addr object for the current listener.
	Addr() net.Addr

	// ClusterAddr returns the net.Addr object for the route listener.
	ClusterAddr() *net.TCPAddr

	// MonitorAddr will return the net.Addr object for the monitoring listener.
	MonitorAddr() *net.TCPAddr

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
	*natsserver.Server
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
	Routes      []string

	// applies to ServerPort, i.e., client-server connections
	TLSConfig *tls.Config

	// applies to ClusterPort, i.e., server-server connections within the NATS cluster
	ClusterTLSConfig *tls.Config

	LogLevel NATSLogLevel

	// OPTIONAL
	MaxPayload int
	// OPTIONAL
	MaxConn int
}

func (a *NATSServerConfig) ServerOpts() (*natsserver.Options, error) {
	if err := a.Cluster.Validate(); err != nil {
		return nil, err
	}

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
		return nil, fmt.Errorf("There are port collisions : ServerPort (%d) MonitorPort(%d) ClusterPort(%d)", a.ServerPort, a.MonitorPort, a.ClusterPort)
	}

	if a.MaxPayload <= 0 {
		a.MaxPayload = DEFAULT_MAXPAYLOAD
	}

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

	switch a.LogLevel {
	case NOLOG:
		opts.NoLog = true
	case DEBUG:
		opts.Debug = true
	case TRACE:
		opts.Trace = true
	default:
		return nil, fmt.Errorf("Invalid LogLevel : %d", a.LogLevel)
	}

	return opts, nil
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
