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

	"sync"

	"github.com/rs/zerolog"
	"gopkg.in/tomb.v2"
	"zombiezen.com/go/capnproto2"
	"zombiezen.com/go/capnproto2/rpc"
)

// StartRPCService creates and starts a new RPCService asynchronously.
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
		ServiceCommandChannel: NewServiceCommandChannel(service, 1),
		server:                server,
		connSemaphore:         NewCountingSemaphore(maxConns),
		conns:                 make(map[uint64]*rpc.Conn),
		logger:                NewConnLogger(service.logger),
		listener:              &listener{factory: listenerFactory},
	}
	rpcService.start()

	return rpcService, nil
}

// ListenerFactory provides a way to start a new listener.
// It abstracts away how the listener is created and what network address it listens on.
type ListenerFactory func() (net.Listener, error)

// RPCService provides the infrastructure to run a capnp RPC based server.
//
// Design overview:
//
//    main service goroutine				o
//	  listener goroutine					o
//	  RPC conn goroutines				  o o o
//
//    connection semaphore				|X|X|X|X|X|X|X|_|_|_|
//	     (X = token)
//
//							In the above example, there are 10 total tokens available, i.e., a max number of concurrent
// 							connections supported is 10. There are 3 active connections, and there are 7 tokens available,
//							7 more connections can be made.
//
// - main service goroutine
//   - starts the listener goroutine
//	 - monitors the listener goroutine and will automatically restart it if it dies
//   - tracks rpc connections
//   - when killed, it will kill the listener followed by any registered rpc conns
// - listener goroutine
//	 - the total number of concurrent connections is limited by a counting semaphore - in order for the listener to accept
//     a new connection, it must first acquire a token
//	 - each new RPC connection is handled in a new goroutine
// - RPC conn handler goroutine
//   - handles RPC requests
//	 - registers itself
//	 - once the connection is closed, then the connection token is released and unregisters itself
type RPCService struct {
	*ServiceCommandChannel

	listener *listener
	server   RPCMainInterface

	connSemaphore CountingSemaphore
	connSeq       Sequence
	conns         map[uint64]*rpc.Conn

	// wraps the service logger
	logger rpc.Logger
}

type listener struct {
	tomb.Tomb
	factory ListenerFactory

	m sync.Mutex
	l net.Listener
}

func (a *listener) get() net.Listener {
	a.m.Lock()
	l := a.l
	a.m.Unlock()
	return l
}

func (a *listener) start() (listener net.Listener, err error) {
	a.m.Lock()
	if a.l == nil {
		// starting for the first time
		listener, err = a.factory()
		if err != nil {
			err = NewListenerFactoryError(err)
		}
	} else {
		// restarting using same address as before
		addr := a.l.Addr()
		listener, err = net.Listen(addr.Network(), addr.String())
		if err != nil {
			err = NewNetListenError(err)
		}
	}
	a.l = listener
	a.m.Unlock()
	return
}

// RPCMainInterface provides the RPC server main interface
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
			case <-a.listener.Dead():
				select {
				case <-a.Dying():
				default:
					a.restartListener()
				}
			case f := <-a.CommandChan():
				f()
			}
		}
	})
}

func (a *RPCService) startListener() {
	a.listener.Go(func() (err error) {
		listener, err := a.listener.start()
		if err != nil {
			return err
		}
		defer listener.Close()
		mainInterface, err := a.server()
		if err != nil {
			return NewRPCServerFactoryError(err)
		}

		SERVICE_STARTED.Log(a.Logger().Info()).Msg("started")

		for {
			select {
			case <-a.Dying():
				return nil
			case <-a.listener.Dying():
				return nil
			case <-a.connSemaphore:
				conn, err := listener.Accept()
				if err != nil {
					return err
				}
				go func(conn net.Conn) {
					rpcConn := rpc.NewConn(rpc.StreamTransport(conn), rpc.MainInterface(mainInterface), rpc.ConnLog(a.logger))
					connKey := a.connSeq.Next()
					a.Submit(a.registerConn(connKey, rpcConn))
					defer func() {
						a.returnConnToken()
						RPC_SERVICE_CONN_CLOSED.Log(a.Logger().Debug()).Msg("RPCService conn closed")
						a.Submit(a.unregisterConn(connKey))
					}()
					err := rpcConn.Wait()
					if err != nil {
						a.Service.Logger().Info().Err(err).Msg("")
					}
				}(conn)
			}
		}
	})
}

func (a *RPCService) restartListener() {
	a.logListenerRestart()
	a.listener.Tomb = tomb.Tomb{}
	a.startListener()
}

func (a *RPCService) stop() {
	if listener := a.listener.get(); listener != nil {
		listener.Close()
	}

	for _, conn := range a.conns {
		if err := conn.Close(); err != nil {
			a.Logger().Warn().Err(err).Msg("Error on rpc.Conn.Close()")
		}
	}
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

func (a *RPCService) logListenerRestart() {
	err := a.listener.Err()
	restartEvent := RPC_SERVICE_LISTENER_RESTART.Log(a.Logger().Warn())
	if err != nil {
		restartEvent.Err(err)
	}
	restartEvent.Msg("restarting RPCService listener")
}

func (a *RPCService) returnConnToken() {
	select {
	case <-a.Dying():
	case a.connSemaphore <- struct{}{}:
	}
}

// MaxConns returns the max number of concurrent connections supported by this RPC server
func (a *RPCService) MaxConns() int {
	return cap(a.connSemaphore)
}

// ActiveConns returns the current number of active connections
func (a *RPCService) ActiveConns() int {
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

// TotalConnsCreated returns the total number of connections that have been created since the RPC server initially started
func (a *RPCService) TotalConnsCreated() uint64 {
	return a.connSeq.Value()
}

// ListenerAddress returns the address that the RPC server is bound to.
func (a *RPCService) ListenerAddress() (net.Addr, error) {
	select {
	case <-a.Dying():
		return nil, ErrServiceNotAlive
	default:
		if l := a.listener.get(); l != nil {
			return l.Addr(), nil
		}
		return nil, ErrRPCListenerNotStarted
	}
}

func (a *RPCService) ListenerAlive() bool {
	return a.Alive() && a.listener.Alive()
}

type listenerAddress struct {
	addr net.Addr
	err  error
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
