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
	"zombiezen.com/go/capnproto2"
	"zombiezen.com/go/capnproto2/rpc"
)

// StartRPCService creates and starts a new RPCService asynchronously.
//
//
func StartRPCService(service *Service, listenerFactory ListenerFactory, server RPCMainInterface, maxConns uint) (*RPCService, error) {
	if service == nil {
		return nil, ErrServiceNil
	}
	if !service.Alive() {
		return nil, ErrServiceNotAlive
	}
	if listenerFactory == nil {
		return nil, ErrListenerFactoryNil
	}
	if server == nil {
		return nil, ErrRPCMainInterfaceNil
	}
	if maxConns == 0 {
		return nil, ErrRPCServiceMaxConnsZero
	}

	rpcService := &RPCService{
		Service:                service,
		listenerFactory:        listenerFactory,
		server:                 server,
		connSemaphore:          NewCountingSemaphore(maxConns),
		conns:                  make(map[uint64]*rpc.Conn),
		logger:                 NewConnLogger(service.logger),
		workQueue:              make(chan func()),
		getListenerAddressChan: make(chan chan listenerAddress),
	}
	rpcService.start()

	return rpcService, nil
}

type ListenerFactory func() (net.Listener, error)

type RPCService struct {
	*Service

	listenerFactory ListenerFactory

	listenerTomb tomb.Tomb
	listener     net.Listener
	server       RPCMainInterface

	connSemaphore CountingSemaphore

	connSeq   uint64
	conns     map[uint64]*rpc.Conn
	workQueue chan func()

	getListenerAddressChan chan chan listenerAddress

	// wraps the service logger
	logger rpc.Logger
}

type RPCMainInterface func() (capnp.Client, error)

func (a *RPCService) start() {
	a.Go(func() error {
		SERVICE_STARTING.Log(a.Logger().Info()).Msg("starting")
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
					a.restartListener()
				}
			case f := <-a.workQueue:
				f()
			case c := <-a.getListenerAddressChan:
				a.getListenerAddress(c)
			}
		}
	})
}

func (a *RPCService) restartListener() {
	a.logListenerRestart()
	a.listenerTomb = tomb.Tomb{}
	a.startListener()
}

func (a *RPCService) registerConn(key uint64, conn *rpc.Conn) func() {
	return func() {
		a.conns[key] = conn
		RPC_SERVICE_NEW_CONN.Log(a.Logger().Debug()).Int("conns", len(a.conns)).Msg("new RPCService conn")
	}
}

func (a *RPCService) unregisterConn(key uint64) func() {
	return func() {
		delete(a.conns, key)
		RPC_SERVICE_CONN_REMOVED.Log(a.Logger().Debug()).Int("conns", len(a.conns)).Msg("RPCService conn removed")
	}
}

func (a *RPCService) execute(command func()) {
	select {
	case <-a.Dying():
	case a.workQueue <- command:
	}
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
		mainInterface, err := a.server()
		if err != nil {
			return NewRPCServerFactoryError(err)
		}

		SERVICE_STARTED.Log(a.Logger().Info()).Msg("started")

		for {
			select {
			case <-a.Dying():
				return nil
			case <-a.connSemaphore:
				conn, err := a.listener.Accept()
				if err != nil {
					return err
				}
				rpcConn := rpc.NewConn(rpc.StreamTransport(conn), rpc.MainInterface(mainInterface), rpc.ConnLog(a.logger))
				a.connSeq++
				connKey := a.connSeq
				a.execute(a.registerConn(connKey, rpcConn))
				go func() {
					defer func() {
						a.returnConnToken()
						RPC_SERVICE_CONN_CLOSED.Log(a.Logger().Debug()).Msg("RPCService conn closed")
						a.execute(a.unregisterConn(connKey))
					}()
					err := rpcConn.Wait()
					if err != nil {
						a.Service.Logger().Info().Err(err).Msg("")
					}
				}()
			}
		}
	})
}

func (a *RPCService) returnConnToken() {
	select {
	case <-a.Dying():
	case a.connSemaphore <- struct{}{}:
	}
}

func (a *RPCService) stop() {
	a.listener.Close()

	for _, conn := range a.conns {
		if err := conn.Close(); err != nil {
			a.Logger().Warn().Err(err).Msg("Error on rpc.Conn.Close()")
		}
	}
}

func (a *RPCService) MaxConns() int {
	return cap(a.connSemaphore)
}

func (a *RPCService) ConnCount() int {
	count := cap(a.connSemaphore) - len(a.connSemaphore)
	if count == cap(a.connSemaphore) {
		return count
	}
	if count > 0 {
		// the listener will acquire the next token immediately, and wait for a connection
		// thus, we need to account for it
		return count - 1
	}
	return 0
}

func (a *RPCService) AvailableConnCapacity() int {
	connCapacity := len(a.connSemaphore) + 1
	if connCapacity > cap(a.connSemaphore) {
		return cap(a.connSemaphore)
	}
	return connCapacity
}

func (a *RPCService) TotalConnsCreated() uint64 {
	return a.connSeq
}

func (a *RPCService) ListenerAddress() (net.Addr, error) {
	c := make(chan listenerAddress)

	select {
	case <-a.Dying():
		return nil, ErrServiceNotAlive
	case a.getListenerAddressChan <- c:
		select {
		case <-a.Dying():
			return nil, ErrServiceNotAlive
		case result := <-c:
			return result.addr, result.err
		}
	}
}

type listenerAddress struct {
	addr net.Addr
	err  error
}

func (a *RPCService) getListenerAddress(c chan listenerAddress) {
	response := listenerAddress{}
	if a.listener == nil {
		response.err = ErrServiceNotAvailable
	} else {
		response.addr = a.listener.Addr()
	}

	select {
	case <-a.Dying():
	case c <- response:
	}
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
