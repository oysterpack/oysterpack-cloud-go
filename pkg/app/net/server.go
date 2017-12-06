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

package net

import (
	"crypto/tls"
	"net"

	"fmt"

	"sync"

	"time"

	"context"

	"github.com/oysterpack/oysterpack.go/pkg/app"
	"github.com/oysterpack/oysterpack.go/pkg/app/net/config"
	opsync "github.com/oysterpack/oysterpack.go/pkg/app/sync"
	"github.com/prometheus/client_golang/prometheus"
)

// TODO: Server registry

// NewTLSServer requires a tls.Config provider.
//
// errors :
// 	- ServerSettings validation errors
//	- ListenerProviderError
//	- TLSConfigError
func StartServer(settings ServerSettings) (*Server, error) {
	if err := settings.Validate(); err != nil {
		return nil, err
	}

	connCountGauge := app.MetricRegistry.Gauge(settings.Service.ID(), SERVER_CONN_COUNT_METRIC_ID)
	if connCountGauge == nil {
		err := fmt.Errorf("Server conn count gauge metric missing : ServiceID(0x%x) : MetricID(0x%x)", settings.Service.ID(), SERVER_CONN_COUNT_METRIC_ID)
		return nil, app.NewConfigError(err)
	}
	totalConnCreatedCount := app.MetricRegistry.Counter(settings.Service.ID(), SERVER_CONN_TOTAL_CREATED_METRIC_ID)
	if totalConnCreatedCount == nil {
		err := fmt.Errorf("Server total conn created counter metric missing : ServiceID(0x%x) : MetricID(0x%x)", settings.Service.ID(), SERVER_CONN_TOTAL_CREATED_METRIC_ID)
		return nil, app.NewConfigError(err)
	}

	l, err := settings.newListener()
	if err != nil {
		return nil, err
	}

	server := &Server{
		Service:               settings.Service,
		settings:              settings,
		connSemaphore:         opsync.NewCountingSemaphore(uint(settings.maxConns)),
		listener:              l,
		running:               make(chan struct{}),
		connCount:             connCountGauge,
		totalConnCreatedCount: totalConnCreatedCount,
	}

	server.Service.Go(server.run)
	server.Service.Go(func() error {
		select {
		case <-server.Service.Dying():
			server.closeListener()
			return nil
		case <-app.Dying():
			server.closeListener()
			return nil
		}
	})

	return server, nil
}

// NewServerSettings constructs a new ServerSettings
//
// errors:
//	- app.ErrServiceNil
//	- app.ErrServiceNotAlive
//	- ErrConnHandlerNil
//	- errors from NewServerSpec()
func NewServerSettings(service *app.Service, spec config.ServerSpec, handler ConnHandler) (ServerSettings, error) {
	if service == nil {
		return ServerSettings{}, app.ErrServiceNil
	}
	if !service.Alive() {
		return ServerSettings{}, app.ErrServiceNotAlive
	}
	if handler == nil {
		return ServerSettings{}, ErrConnHandlerNil
	}

	serverSpec, err := NewServerSpec(spec)
	if err != nil {
		return ServerSettings{}, err
	}
	return ServerSettings{service, serverSpec, handler}, nil
}

// ServerSettings is used to create a new Server
type ServerSettings struct {
	*app.Service

	*ServerSpec

	ConnHandler ConnHandler
}

// Validate validates the settings
//
// errors:
//	- app.ErrServiceNil
//	- app.ErrServiceNotAlive
//	- ErrListenerProviderNil
//	- ErrTLSConfigProviderNil
//	- ErrServerNameBlank
func (a *ServerSettings) Validate() error {
	if a.Service == nil {
		return app.ErrServiceNil
	}
	if !a.Service.Alive() {
		return app.ErrServiceNotAlive
	}
	if a.ServerSpec == nil {
		return ErrServerSpecNil
	}
	if a.ConnHandler == nil {
		return ErrConnHandlerNil
	}

	return nil
}

func (a *ServerSettings) newListener() (net.Listener, error) {
	// starting for the first time
	l, err := a.ServerSpec.ListenerProvider()()
	if err != nil {
		return nil, NewListenerProviderError(err)
	}

	tlsConfig, err := a.TLSConfigProvider()()
	if err != nil {
		return nil, NewTLSConfigError(err)
	}
	return tls.NewListener(l, tlsConfig), nil
}

// Listener abstracts away how the net.Listener is provided.
//
// Design:
//	- every server maps to a Service, i.e., the server lifecycle aligns with the service lifecycle
//	-
type Server struct {
	settings ServerSettings

	*app.Service
	connSemaphore *opsync.CountingSemaphore

	listenerMutex sync.Mutex
	listener      net.Listener
	tlsConfig     *tls.Config

	connSeq opsync.Sequence

	// signal
	running chan struct{}

	// metrics
	connCount             prometheus.Gauge
	totalConnCreatedCount prometheus.Counter
}

// Running is used to signal when the server is running.
// It does not mean that clients can connect - that depends on available server connections
func (a *Server) Running() <-chan struct{} {
	return a.running
}

func (a *Server) closeListener() {
	a.listenerMutex.Lock()
	defer a.listenerMutex.Unlock()
	if a.listener != nil {
		a.listener.Close()
		a.listener = nil
		SERVER_LISTENER_CLOSED.Log(a.Service.Logger().Info()).Msg("Listener closed")
	}
}

func (a *Server) getListener() (l net.Listener, err error) {
	a.listenerMutex.Lock()
	defer a.listenerMutex.Unlock()
	if a.listener == nil {
		a.listener, err = a.settings.newListener()
		if err != nil {
			return nil, err
		}
		SERVER_LISTENER_RESTART.Log(a.Logger().Warn()).Msg("restarting listener")
	}
	return a.listener, nil
}

// TODO: metrics :
// 	- connection count - gauge
func (a *Server) run() (err error) {
	conns := connMap{conns: make(map[uint64]net.Conn)}

	defer func() {
		a.closeListener()
		conns.closeAll()
		a.connCount.Set(0)
		SERVER_ALL_CONNS_CLOSED.Log(a.Service.Logger().Info()).Msg("All connections are closed")
	}()

	l, err := a.getListener()
	if err != nil {
		return err
	}

	SERVER_LISTENER_STARTED.Log(a.Logger().Info()).
		Str("addr", fmt.Sprintf("%s://%s", l.Addr().Network(), l.Addr().String())).
		Int("max-conns", a.connSemaphore.TotalTokens()).
		Uint8("keep-alive-period-secs", a.settings.keepAlivePeriodSecs).
		Msg("listener started")

	close(a.running)
	for {
		select {
		case <-a.Service.Dying():
			return nil
		case <-a.connSemaphore.C:
			l, err := a.getListener()
			if err != nil {
				if !a.Service.Alive() || !app.Alive() {
					// the error can be ignored because it means the server is being killed
					return nil
				}
				return err
			}

			conn, err := l.Accept()
			if err != nil {
				if !a.Service.Alive() || !app.Alive() {
					// the error can be ignored because it means the server is being killed
					return nil
				}
				return err
			}
			a.connCount.Inc()
			a.totalConnCreatedCount.Inc()
			if err := a.settings.ConfigureConnBuffers(conn); err != nil {
				// should never happen
				a.Logger().Warn().Err(err).Msg("Failed to configure conn buffers")
			}

			if tcpConn, ok := conn.(*net.TCPConn); ok {
				tcpConn.SetKeepAlive(true)
				tcpConn.SetKeepAlivePeriod(time.Second * time.Duration(a.settings.keepAlivePeriodSecs))
			}

			go func() {
				connKey := a.connSeq.Next()
				conns.put(connKey, conn)
				SERVER_NEW_CONN.Log(a.Logger().Debug()).Int("conns", a.ConnectionCount()).Msg("new conn")
				defer func() {
					a.connSemaphore.ReturnToken()
					conns.close(connKey)
					a.connCount.Dec()
					SERVER_CONN_CLOSED.Log(a.Logger().Debug()).Msg("conn closed")
				}()
				ctx := context.WithValue(context.Background(), CTX_SERVER_SPEC, a.settings.ServerSpec)
				ctx = context.WithValue(ctx, CTX_SERVICE, a.Service)
				a.settings.ConnHandler(ctx, conn)
			}()

			if a.connSemaphore.AvailableTokens() == 0 {
				// no longer accept connections - we want clients to fail fast and not hang waiting to be served
				a.closeListener()
				SERVER_MAX_CONNS_REACHED.Log(a.Service.Logger().Warn()).Msg("Listener has been closed until connections free up.")
			}
		}
	}
}

func (a *Server) newContext() context.Context {
	ctx := context.WithValue(context.Background(), CTX_SERVER_SPEC, a.settings.ServerSpec)
	ctx = context.WithValue(ctx, CTX_SERVICE, a.Service)
	return ctx
}

func (a *Server) Address() (net.Addr, error) {
	a.listenerMutex.Lock()
	defer a.listenerMutex.Unlock()
	if a.listener != nil {
		return a.listener.Addr(), nil
	}
	return nil, ErrListenerDown
}

func (a *Server) MaxConnections() uint {
	return uint(a.settings.maxConns)
}

func (a *Server) ConnectionCount() int {
	return a.connSemaphore.TotalTokens() - a.connSemaphore.AvailableTokens()
}

type connMap struct {
	sync.Mutex
	conns map[uint64]net.Conn
}

func (a connMap) put(key uint64, conn net.Conn) {
	a.Lock()
	defer a.Unlock()
	a.conns[key] = conn
}

func (a connMap) delete(key uint64) {
	a.Lock()
	defer a.Unlock()
	delete(a.conns, key)
}

func (a connMap) close(key uint64) {
	a.Lock()
	defer a.Unlock()
	conn := a.conns[key]
	if conn != nil {
		conn.Close()
		delete(a.conns, key)
	}
}

func (a connMap) closeAll() {
	a.Lock()
	defer a.Unlock()
	for _, conn := range a.conns {
		conn.Close()
	}
	a.conns = make(map[uint64]net.Conn)
}
