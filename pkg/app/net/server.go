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

	"github.com/oysterpack/oysterpack.go/pkg/app"
	opsync "github.com/oysterpack/oysterpack.go/pkg/app/sync"
	"gopkg.in/tomb.v2"
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

	l, err := settings.newListener()
	if err != nil {
		return nil, err
	}

	server := &Server{
		Service:       settings.Service,
		settings:      settings,
		connSemaphore: opsync.NewCountingSemaphore(settings.MaxConns),
		listener:      l,
	}

	server.Service.Go(server.run)

	return server, nil
}

// ServerSettings is used to create a new Server
type ServerSettings struct {
	*app.Service

	Name string

	ListenerProvider  func() (net.Listener, error)
	TLSConfigProvider func() (*tls.Config, error)

	MaxConns uint

	ConnHandler func(conn net.Conn)
}

// Validate validates the settings
//
// errors:
//	- app.ErrServiceNil
//	- app.ErrServiceNotAlive
//	- ErrListenerProviderNil
//	- ErrTLSConfigProviderNil
func (a *ServerSettings) Validate() error {
	if a.Service == nil {
		return app.ErrServiceNil
	}
	if !a.Service.Alive() {
		return app.ErrServiceNotAlive
	}
	if a.ListenerProvider == nil {
		return ErrListenerProviderNil
	}
	if a.TLSConfigProvider == nil {
		return ErrTLSConfigProviderNil
	}
	if a.ConnHandler == nil {
		return ErrConnHandlerNil
	}
	return nil
}

func (a *ServerSettings) newListener() (net.Listener, error) {
	// starting for the first time
	l, err := a.ListenerProvider()
	if err != nil {
		return nil, NewListenerProviderError(err)
	}

	tlsConfig, err := a.TLSConfigProvider()
	if err != nil {
		return nil, NewTLSConfigError(err)
	}
	return tls.NewListener(l, tlsConfig), nil
}

// Listener abstracts away how the net.Listener is provided.
//
// Design:
//	- every server belongs to a Service
//	-
type Server struct {
	settings ServerSettings

	tomb.Tomb
	*app.Service
	connSemaphore *opsync.CountingSemaphore

	listener  net.Listener
	tlsConfig *tls.Config

	connSeq opsync.Sequence
}

// TODO: metrics :
// 	- connection count
func (a *Server) run() (err error) {
	conns := connMap{conns: make(map[uint64]net.Conn)}

	defer func() {
		a.listener.Close()
		conns.closeAll()
	}()

	SERVER_LISTENER_STARTED.Log(a.Logger().Info()).
		Str("addr", fmt.Sprintf("%s://%s", a.listener.Addr().Network(), a.listener.Addr().String())).
		Int("max-conns", a.connSemaphore.TotalTokens()).
		Msg("listener started")

	for {
		select {
		case <-a.Dying():
			return nil
		case <-a.Service.Dying():
			return nil
		case <-a.connSemaphore.C:
			if a.listener == nil {
				a.listener, err = a.settings.newListener()
				if err != nil {
					return err
				}
				a.logListenerRestart()
			}

			conn, err := a.listener.Accept()
			if err != nil {
				return err
			}

			go func() {
				connKey := a.connSeq.Next()
				conns.put(connKey, conn)
				SERVER_NEW_CONN.Log(a.Logger().Debug()).Int("conns", a.ConnectionCount()).Msg("new conn")
				defer func() {
					a.connSemaphore.ReturnToken()
					conns.close(connKey)
					SERVER_CONN_CLOSED.Log(a.Logger().Debug()).Msg("conn closed")
				}()
				a.settings.ConnHandler(conn)
			}()

			if a.connSemaphore.AvailableTokens() == 0 {
				// no longer accept connections - we want clients to fail fast and not hang waiting to be served
				a.listener.Close()
				a.listener = nil
				SERVER_MAX_CONNS_REACHED.Log(a.Service.Logger().Warn()).Msg("Listener has been closed until connections free up.")
			}
		}
	}
}

func (a *Server) Address() (net.Addr, error) {
	if a.listener != nil {
		return a.listener.Addr(), nil
	}
	return nil, ErrListenerDown
}

func (a *Server) MaxConnections() uint {
	return a.settings.MaxConns
}

func (a *Server) ConnectionCount() int {
	return a.connSemaphore.TotalTokens() - a.connSemaphore.AvailableTokens()
}

func (a *Server) logListenerRestart() {
	err := a.Err()
	restartEvent := SERVER_LISTENER_RESTART.Log(a.Logger().Warn())
	if err != nil {
		restartEvent.Err(err)
	}
	restartEvent.Msg("restarting listener")
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
