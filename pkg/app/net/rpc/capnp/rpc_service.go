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

package capnp

import (
	"context"
	"net"

	"fmt"

	"github.com/oysterpack/oysterpack.go/pkg/app"
	opnet "github.com/oysterpack/oysterpack.go/pkg/app/net"
	opsync "github.com/oysterpack/oysterpack.go/pkg/app/sync"
	"github.com/rs/zerolog"
	"gopkg.in/tomb.v2"
	"zombiezen.com/go/capnproto2"
	"zombiezen.com/go/capnproto2/rpc"
)

// StartRPCService creates and starts a new RPCService asynchronously.
//
// errors:
//	- validation errors
//		- ErrServiceNil
//		- ErrServiceNotAlive
//		- ErrListenerFactoryNil
//		- ErrRPCMainInterfaceNil
//		- ErrRPCServiceMaxConnsZero
func StartRPCService(service *Service, listenerFactory opnet.ListenerFactory, tlsConfigProvider opnet.TLSConfigProvider, server RPCMainInterface, maxConns uint) (*RPCService, error) {
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

	serviceCommandChannel, err := NewServiceCommandChannel(service, 1)
	if err != nil {
		return nil, err
	}

	rpcService := &RPCService{
		ServiceCommandChannel: serviceCommandChannel,
		server:                server,
		startedChan:           make(chan struct{}),
		connSemaphore:         opsync.NewCountingSemaphore(maxConns),
		conns:                 make(map[uint64]*rpc.Conn),
		logger:                NewConnLogger(service.logger),
		listener:              &listener{factory: listenerFactory, tlsConfigProvider: tlsConfigProvider},
	}
	rpcService.start()

	return rpcService, nil
}

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
//   - once the max connection capacity limit has been reached, the listener will automatically close itself. Clients will
//     fail fast with connection refused errors because the port will be down. Load balancers should automatically route
//     clients to other servers with capacity. Once connection capacity is freed up, then the listener will automatically restart.
// - RPC conn handler goroutine
//   - handles RPC requests
//	 - registers itself
//	 - once the connection is closed, then the connection token is released and unregisters itself
//
// Log Events:
// - WARN
//	 - RPC_SERVICE_LISTENER_RESTART - will log the listener error
//	 - RPC_CONN_CLOSE_ERR - if an error occurred when closing the RPC conn
// - INFO
// 	 - SERVICE_STARTING - when the main service goroutine is launched
// 	 - SERVICE_STARTED - once the listener goroutine is launched
// 	 - RPC_SERVICE_LISTENER_STARTED - once the listener goroutine has initialized
// - DEBUG
//   - RPC_SERVICE_NEW_CONN - when the connection is registered
//   - RPC_SERVICE_CONN_REMOVED - when the connection is unregistered
//	 - RPC_SERVICE_CONN_CLOSED - when the conn is closed and the conn goroutine is exiting
//
// Errors :
// - ListenerFactoryError
// - NetListenError
// - RPCServerFactoryError
// - ErrServiceNotAlive
// - ErrRPCListenerNotStarted - on ListenerAddress()
type RPCService struct {
	*app.ServiceCommandChannel

	listener    *listener
	server      RPCMainInterface
	startedChan chan struct{}

	connSemaphore opsync.CountingSemaphore
	connSeq       opsync.Sequence
	conns         map[uint64]*rpc.Conn

	// wraps the service logger
	logger rpc.Logger
}

// RPCMainInterface provides the RPC server main interface
type RPCMainInterface func() (capnp.Client, error)

func (a *RPCService) start() {
	a.Go(func() error {
		app.SERVICE_STARTING.Log(a.Logger().Info()).Msg("starting")
		a.startListener()
		app.SERVICE_STARTED.Log(a.Logger().Info()).Msg("started")

		a.registerRPCService()

		for {
			select {
			case <-a.Dying():
				a.stop()
				return nil
			case <-a.listener.Dead():
				select {
				case <-a.Dying():
				default:
					if _, unrecoverable := a.listener.Err().(UnrecoverableError); unrecoverable {
						a.Kill(a.listener.Err())
					} else {
						a.restartListener()
					}
				}
			case f := <-a.CommandChan():
				f()
			}
		}
	})
}

func (a *RPCService) registerRPCService() {
	submitCommand(func() {
		rpcServices[a.Service.ID()] = a
	})
}

func (a *RPCService) unregisterRPCService() {
	submitCommand(func() {
		delete(rpcServices, a.Service.ID())
	})
}

func (a *RPCService) Started() <-chan struct{} {
	return a.startedChan
}

func (a *RPCService) startListener() {
	a.listener.Go(func() (err error) {
		listener, err := a.listener.start()
		if err != nil {
			return err
		}
		defer func() {
			if listener != nil {
				listener.Close()
			}
		}()
		// signal that the server is started, i.e., the listener was started
		close(a.startedChan)
		mainInterface, err := a.server()
		if err != nil {
			return NewRPCServerFactoryError(err)
		}

		addr := listener.Addr()
		RPC_SERVICE_LISTENER_STARTED.Log(a.Logger().Info()).
			Str("addr", fmt.Sprintf("%s://%s", addr.Network(), addr.String())).
			Int("max-conns", a.MaxConns()).
			Bool("tls", a.listener.tlsConfig != nil).
			Msg("rpc listener started")
		for {
			select {
			case <-a.Dying():
				return nil
			case <-a.listener.Dying():
				return nil
			case <-a.connSemaphore:
				if listener == nil {
					listener, err = a.listener.start()
					if err != nil {
						return err
					}
				}
				conn, err := listener.Accept()
				if err != nil {
					return err
				}
				go func(conn net.Conn) {
					rpcConn := rpc.NewConn(rpc.StreamTransport(conn), rpc.MainInterface(mainInterface), rpc.ConnLog(a.logger))
					connKey := a.connSeq.Next()
					a.Submit(a.registerConn(connKey, rpcConn))
					defer func() {
						a.connSemaphore.ReturnToken()
						RPC_SERVICE_CONN_CLOSED.Log(a.Logger().Debug()).Msg("RPCService conn closed")
						a.Submit(a.unregisterConn(connKey))
					}()
					err := rpcConn.Wait()
					if err != nil {
						a.Service.Logger().Info().Err(err).Msg("")
					}
				}(conn)

				if a.RemainingConnectionCapacity() == 0 {
					// no longer accept connections - we want clients to fail fast and not hang waiting to be served
					listener.Close()
					listener = nil
				}
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
		a.closeRPCConn(conn)
	}

	a.unregisterRPCService()
}

func (a *RPCService) closeRPCConn(conn *rpc.Conn) {
	defer func() {
		if p := recover(); p != nil {
			RPC_CONN_CLOSE_ERR.Log(a.Logger().Warn()).Msgf("Panic on rpc.Conn.Close() : %v", p)
		}
	}()
	if err := conn.Close(); err != nil {
		RPC_CONN_CLOSE_ERR.Log(a.Logger().Warn()).Err(err).Msg("Error on rpc.Conn.Close()")
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

// MaxConns returns the max number of concurrent connections supported by this RPC server
func (a *RPCService) MaxConns() int {
	return cap(a.connSemaphore)
}

// ActiveConns returns the current number of active connections
func (a *RPCService) ActiveConns() int {
	count := cap(a.connSemaphore) - len(a.connSemaphore)
	if count == cap(a.connSemaphore) {
		c := make(chan int, 1)
		a.Submit(func() {
			c <- len(a.conns)
		})

		select {
		case <-a.Dying():
			return 0
		case count = <-c:
			return count
		}
	}
	if count > 0 {
		// the listener will acquire the next token immediately, and wait for a connection
		// thus, we need to account for it
		return count - 1
	}
	return 0
}

func (a *RPCService) RemainingConnectionCapacity() int {
	return len(a.connSemaphore)
}

// TotalConnsCreated returns the total number of connections that have been created since the RPC server initially started
func (a *RPCService) TotalConnsCreated() uint64 {
	return a.connSeq.Value()
}

// ListenerAddress returns the address that the RPC server is bound to.
// errors :
// - ErrServiceNotAlive
// - ErrRPCListenerNotStarted
func (a *RPCService) ListenerAddress() (net.Addr, error) {
	select {
	case <-a.Dying():
		return nil, app.ErrServiceNotAlive
	default:
		if l := a.listener.get(); l != nil {
			return l.Addr(), nil
		}
		return nil, ErrRPCListenerNotStarted
	}
}

// ListenerAlive returns true if the main service and listener goroutines are alive
func (a *RPCService) ListenerAlive() bool {
	return a.Alive() && a.listener.Alive()
}

// NewConnLogger creates a new capnp rpc.Logger, which delegates to the specified zerolog.Logger
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
