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

package app

import (
	"context"
	"net"

	"github.com/rs/zerolog"
	"gopkg.in/tomb.v2"
	"zombiezen.com/go/capnproto2/rpc"
	"zombiezen.com/go/capnproto2"
)

// StartRPCService creates and starts a new RPCService.
//
//
func StartRPCService(service *Service, listenerFactory ListenerFactory, serverFactory RPCServerFactory, maxConns uint) (*RPCService, error) {
	if service == nil {
		return nil, ErrServiceNil
	}
	if !service.Alive() {
		return nil, ErrServiceNotAlive
	}
	if listenerFactory == nil {
		return nil, ErrListenerFactoryNil
	}
	if serverFactory == nil {
		return nil, ErrRPCServerFactoryNil
	}
	if maxConns == 0 {
		return nil, ErrRPCServiceMaxConnsZero
	}

	rpcService := &RPCService{
		Service:         service,
		listenerFactory: listenerFactory,
		serverFactory:   serverFactory,
		connSemaphore:   NewCountingSemaphore(maxConns),
		conns:           make(map[*rpc.Conn]struct{}),
		logger:          NewConnLogger(service.logger),
		connDied:        make(chan *rpc.Conn),
	}

	return rpcService, nil
}

type ListenerFactory func() (net.Listener, error)

type RPCServerFactory func() (CapnpRpcServer, error)

type RPCService struct {
	*Service

	listenerFactory ListenerFactory
	serverFactory   RPCServerFactory

	listenerTomb tomb.Tomb
	listener     net.Listener
	server       CapnpRpcServer

	connSemaphore CountingSemaphore
	conns         map[*rpc.Conn]struct{}
	connDied      chan *rpc.Conn

	// wraps the service logger
	logger rpc.Logger
}

type CapnpRpcServer interface {
	MainInterface() capnp.Client
}

func (a *RPCService) start() {
	a.Go(func() error {
		SERVICE_STARTED.Log(a.Logger().Info()).Msg("started")
		a.startListener()

		for {
			select {
			case <-a.Dying():
				a.stop()
				return nil
			case <-a.listenerTomb.Dead():
				select {
				case <-a.Dying():
				default:
					// restart the listener
					a.logListenerRestart()
					a.listenerTomb = tomb.Tomb{}
					a.startListener()
				}
			case rpcConn := <-a.connDied:
				delete(a.conns, rpcConn)
				RPC_SERVICE_CONN_REMOVED.Log(a.Logger().Debug()).Int("conns",len(a.conns)).Msg("RPCService conn removed")
			}
		}
	})
}

func (a *RPCService) logListenerRestart() {
	err := a.listenerTomb.Err()
	restartEvent := RPC_SERVICE_LISTENER_RESTART.Log(a.Logger().Warn())
	if err != nil {
		restartEvent.Err(err)
	}
	restartEvent.Msg("restarting RPCService listener")
}

func (a *RPCService) startListener() {
	a.listenerTomb.Go(func() (err error) {
		a.listener, err = a.listenerFactory()
		if err != nil {
			return NewListenerFactoryError(err)
		}
		defer a.listener.Close()
		a.server, err = a.serverFactory()
		if err != nil {
			return NewRPCServerFactoryError(err)
		}

		for {
			select {
			case <-a.Dying():
				return nil
			case <-a.connSemaphore:
				conn, err := a.listener.Accept()
				if err != nil {
					return err
				}
				rpcConn := rpc.NewConn(rpc.StreamTransport(conn), rpc.MainInterface(a.server.MainInterface()), rpc.ConnLog(a.logger))
				a.conns[rpcConn] = struct{}{}
				RPC_SERVICE_NEW_CONN.Log(a.Logger().Debug()).Int("conns",len(a.conns)).Msg("new RPCService conn")
				go func() {
					defer func() {
						RPC_SERVICE_CONN_CLOSED.Log(a.Logger().Debug()).Int("conns",len(a.conns)).Msg("RPCService conn closed")
						select {
						case <-a.Dying():
						case a.connDied <- rpcConn:
						}
					}()
					err := rpcConn.Wait()
					if err != nil {
						a.Service.Logger()
					}
				}()
			}
		}
	})
}

func (a *RPCService) stop() {
	// TODO
}

func (a *RPCService) MaxConns() int {
	return cap(a.connSemaphore)
}

func (a *RPCService) ConnCount() int {
	return len(a.connSemaphore)
}

func (a *RPCService) ListenerAddress() (net.Addr, error) {
	if !a.Alive() {
		return nil, ErrServiceNotAlive
	}
	if a.listener == nil {
		return nil, ErrServiceNotAvailable
	}
	return a.listener.Addr(), nil
}

type RpcConn struct {
	tomb.Tomb
	conn *rpc.Conn
}

func NewConnLogger(logger zerolog.Logger) rpc.Logger { return &CapnpRpcConnLogger{logger} }

// CapnpRpcConnLogger implements zombiezen.com/go/capnproto2/rpc/Logger.
type CapnpRpcConnLogger struct {
	zerolog.Logger
}

// Infof implements zombiezen.com/go/capnproto2/rpc/Logger
func (a *CapnpRpcConnLogger) Infof(ctx context.Context, format string, args ...interface{}) {
	a.Info().Msgf(format, args...)
}

// Errorf implements zombiezen.com/go/capnproto2/rpc/Logger
func (a *CapnpRpcConnLogger) Errorf(ctx context.Context, format string, args ...interface{}) {
	a.Error().Msgf(format, args...)
}
