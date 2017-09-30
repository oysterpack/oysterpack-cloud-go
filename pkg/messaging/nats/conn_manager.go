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

package nats

import (
	"sync"

	"github.com/nats-io/go-nats"
	"github.com/nats-io/nuid"
	"github.com/oysterpack/oysterpack.go/pkg/logging"
	"github.com/oysterpack/oysterpack.go/pkg/service"
)

// TODO: metrics

// ConnManager tracks connections that are created through it.
// The following connection lifecycle events are tracked:
//
// 1. Connection disconnects
// 2. Connection reconnects
// 3. Connection errors
// 4. Connection closed - when closed, it is no longer tracked and removed async from its managed list
type ConnManager interface {
	Connect(tags ...string) (conn *ManagedConn, err error)

	ManagedConn(id string) *ManagedConn

	ConnCount() int

	ConnInfo(id string) *ConnInfo

	ConnInfos() []*ConnInfo

	CloseAll()

	ConnectedCount() (count int, total int)

	DisconnectedCount() (count int, total int)
}

func DefaultOptions() nats.Options {
	options := nats.GetDefaultOptions()
	options.Url = nats.DefaultURL
	DefaultConnectTimeout(&options)
	DefaultReConnectTimeout(&options)
	AlwaysReconnect(&options)
	return options
}

// NewConnManager factory method.
// Default connection options are : DefaultConnectTimeout, DefaultReConnectTimeout, AlwaysReconnect
func NewConnManager(options ...nats.Option) ConnManager {
	connOptions := DefaultOptions()
	for _, option := range options {
		if err := option(&connOptions); err != nil {
			logger.Panic().Err(err).Msg("Failed to apply option")
		}
	}

	connMgr := &connManager{options: connOptions}
	connMgr.init()
	return connMgr
}

// ManagedConn is a managed NATS connection.

type connManager struct {
	options nats.Options

	*service.RestartableService

	mutex sync.RWMutex

	n     nuid.NUID
	conns map[string]*ManagedConn
}

// Connect creates a new managed NATS connection that connects to a NATS server running on the localhost.
// Tags are used to help identify how the connection is being used by the app. Some common Tags are provided, but
// additional application specific Tags can also be added, e.g., the name of the service using the connection.
// Connections are configured to always reconnect.
//
// Use Case : A gnatsd process runs on each server node, forming cluster mesh. Applications connect to the local gnatsd
// server to communicate remotely, thus forming a service mesh. The apps are not network aware, i.e., they always connect
// to localhost.
//
func (a *connManager) Connect(tags ...string) (*ManagedConn, error) {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	nc, err := a.options.Connect()
	if err != nil {
		return nil, err
	}
	connId := nuid.Next()
	managedConn := NewManagedConn(connId, nc, tags)
	a.conns[connId] = managedConn

	nc.SetClosedHandler(func(conn *nats.Conn) {
		logger.Info().Str(logging.EVENT, EVENT_CONN_CLOSED).Str(CONN_ID, connId).Msg("")
		a.mutex.Lock()
		defer a.mutex.Unlock()
		delete(a.conns, connId)
		logger.Info().Str(logging.EVENT, EVENT_CONN_CLOSED).Str(CONN_ID, connId).Msg("deleted")
	})

	nc.SetDisconnectHandler(func(conn *nats.Conn) { managedConn.disconnected() })
	nc.SetReconnectHandler(func(conn *nats.Conn) { managedConn.reconnected() })
	nc.SetDiscoveredServersHandler(func(conn *nats.Conn) { managedConn.discoveredServers() })
	nc.SetErrorHandler(func(conn *nats.Conn, subscription *nats.Subscription, err error) {
		managedConn.subscriptionError(subscription, err)
	})

	return managedConn, nil
}

func (a *connManager) ConnCount() int {
	a.mutex.RLock()
	defer a.mutex.RUnlock()
	return len(a.conns)
}

func (a *connManager) init() {
	a.mutex.Lock()
	defer a.mutex.Unlock()
	a.conns = make(map[string]*ManagedConn)
}

func (a *connManager) CloseAll() {
	a.mutex.Lock()
	defer a.mutex.Unlock()
	for _, nc := range a.conns {
		nc.Conn.SetClosedHandler(nil)
		nc.Conn.Close()
	}
	a.conns = make(map[string]*ManagedConn)
}

func (a *connManager) ConnInfo(id string) *ConnInfo {
	a.mutex.RLock()
	defer a.mutex.RUnlock()
	if c := a.conns[id]; c != nil {
		return c.ConnInfo()
	}
	return nil
}

func (a *connManager) ConnInfos() []*ConnInfo {
	a.mutex.RLock()
	defer a.mutex.RUnlock()
	infos := make([]*ConnInfo, len(a.conns))
	i := 0
	for _, v := range a.conns {
		infos[i] = v.ConnInfo()
		i++
	}
	return infos
}

func (a *connManager) ManagedConn(id string) *ManagedConn {
	a.mutex.RLock()
	defer a.mutex.RUnlock()
	return a.conns[id]
}

func (a *connManager) ConnectedCount() (count int, total int) {
	a.mutex.RLock()
	defer a.mutex.RUnlock()
	total = len(a.conns)
	for _, c := range a.conns {
		if c.Conn.IsConnected() {
			count++
		}
	}
	return
}

func (a *connManager) DisconnectedCount() (count int, total int) {
	a.mutex.RLock()
	defer a.mutex.RUnlock()
	total = len(a.conns)
	for _, c := range a.conns {
		if !c.Conn.IsConnected() {
			count++
		}
	}
	return
}