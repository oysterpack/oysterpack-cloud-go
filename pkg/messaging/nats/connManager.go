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
	"fmt"
	"time"

	"sync"

	"github.com/nats-io/go-nats"
	"github.com/nats-io/nuid"
	"github.com/oysterpack/oysterpack.go/pkg/logging"
	"github.com/oysterpack/oysterpack.go/pkg/service"
)

// ConnectionManager manages NATS connections.
// TODO: metrics
type ConnectionManager interface {
	Connect(tags ...string) (conn *ManagedConn, err error)

	ManagedConn(id string) *ManagedConn

	ConnCount() int

	ConnInfo(id string) *ConnInfo

	ConnInfos() []*ConnInfo

	CloseAll()
}

func NewConnectionManager() ConnectionManager {
	connMgr := &connManager{}
	connMgr.init()
	return connMgr
}

// ManagedConn is a managed NATS connection.
type ManagedConn struct {
	mutex sync.RWMutex

	*nats.Conn

	id      string
	created time.Time
	tags    []string

	reconnects        int
	lastReconnectTime time.Time

	disconnects        int
	lastDisconnectTime time.Time

	errors        int
	lastErrorTime time.Time
}

func (a *ManagedConn) Disconnects() int {
	a.mutex.RLock()
	defer a.mutex.RUnlock()
	return a.disconnects
}

// updates are protected by a mutex to make changes concurrency safe
func (a *ManagedConn) disconnected() {
	a.mutex.Lock()
	defer a.mutex.Unlock()
	a.lastDisconnectTime = time.Now()
	a.disconnects++
	event := logger.Info().Str(logging.EVENT, EVENT_CONN_DISCONNECT).Str(CONN_ID, a.id).Int(DISCONNECTS, a.disconnects)
	if len(a.tags) > 0 {
		event.Strs(CONN_TAGS, a.tags)
	}
	event.Msg("")
}

func (a *ManagedConn) reconnected() {
	a.lastReconnectTime = time.Now()
	a.Reconnects++
	event := logger.Info().Str(logging.EVENT, EVENT_CONN_RECONNECT).Str(CONN_ID, a.id)
	if len(a.tags) > 0 {
		event.Strs(CONN_TAGS, a.tags)
	}
	event.Msg("")
}

func (a *ManagedConn) subscriptionError(subscription *nats.Subscription, err error) {
	a.errors++
	a.lastErrorTime = time.Now()

	event := logger.Error().Str(logging.EVENT, EVENT_CONN_ERR).Str(CONN_ID, a.id).Err(err).Bool(SUBSCRIPTION_VALID, subscription.IsValid())
	if len(a.tags) > 0 {
		event.Strs(CONN_TAGS, a.tags)
	}

	subscriptionErrors := []error{}
	if maxQueuedMsgs, maxQueuedBytes, e := subscription.MaxPending(); e != nil {
		subscriptionErrors = append(subscriptionErrors, e)
	} else {
		event.Int(MAX_QUEUED_MSGS, maxQueuedMsgs).Int(MAX_QUEUED_BYTES, maxQueuedBytes)
	}
	if queuedMsgs, queuedBytes, e := subscription.Pending(); e != nil {
		subscriptionErrors = append(subscriptionErrors, e)
	} else {
		event.Int(QUEUED_MSGS, queuedMsgs).Int(QUEUED_BYTES, queuedBytes)
	}
	if queuedMsgLimit, queuedBytesLimit, e := subscription.PendingLimits(); e != nil {
		subscriptionErrors = append(subscriptionErrors, e)
	} else {
		event.Int(QUEUED_MSGS_LIMIT, queuedMsgLimit).Int(QUEUED_BYTES_LIMIT, queuedBytesLimit)
	}
	if delivered, e := subscription.Delivered(); e != nil {
		subscriptionErrors = append(subscriptionErrors, e)
	} else {
		event.Int64(DELIVERED, delivered)
	}
	if dropped, e := subscription.Dropped(); e != nil {
		subscriptionErrors = append(subscriptionErrors, e)
	} else {
		event.Int(DROPPED, dropped)
	}

	if len(subscriptionErrors) > 0 {
		event.Errs("sub", subscriptionErrors)
	}

	event.Msg("")
}

func (a *ManagedConn) discoveredServers() {
	event := logger.Info().Str(logging.EVENT, EVENT_CONN_DISCOVERED_SERVERS).Str(CONN_ID, a.id).Strs(DISCOVERED_SERVERS, a.Conn.DiscoveredServers())
	if len(a.tags) > 0 {
		event.Strs(CONN_TAGS, a.tags)
	}
	event.Msg("")
}

func (a *ManagedConn) ConnInfo() *ConnInfo {
	a.mutex.RLock()
	defer a.mutex.RUnlock()
	return &ConnInfo{
		Id:      a.id,
		Created: a.created,
		Tags:    a.tags,

		Reconnects:        a.reconnects,
		LastReconnectTime: a.lastReconnectTime,

		Disconnects:        a.disconnects,
		LastDisconnectTime: a.lastDisconnectTime,

		Errors:        a.errors,
		LastErrorTime: a.lastErrorTime,
	}
}

type ConnInfo struct {
	Id      string
	Created time.Time
	Tags    []string

	Reconnects        int
	LastReconnectTime time.Time

	Disconnects        int
	LastDisconnectTime time.Time

	Errors        int
	LastErrorTime time.Time
}

func (a *ConnInfo) String() string {
	bytes, err := json.Marshal(a)
	if err != nil {
		// should never happen
		logger.Warn().Err(err).Msg("json.Marshal() failed")
		return fmt.Sprintf("%v", *a)
	}
	return string(bytes)
}

// NewManagedConn is used to create a new ManagedConn
func NewManagedConn(connId string, conn *nats.Conn, tags []string) *ManagedConn {
	return &ManagedConn{
		Conn:    conn,
		id:      connId,
		created: time.Now(),
		tags:    tags,
	}
}

func (a *ManagedConn) String() string {
	bytes, err := json.Marshal(a.ConnInfo())
	if err != nil {
		// should never happen
		logger.Warn().Err(err).Msg("json.Marshal() failed")
		return fmt.Sprintf("%v", *a.ConnInfo())
	}
	return string(bytes)
}

type connManager struct {
	*service.RestartableService

	mutex sync.RWMutex

	n     nuid.NUID
	conns map[string]*ManagedConn
}

// Connect creates a new managed NATS connection.
// Tags are used to help identify how the connection is being used by the app. Some common Tags are provided, but
// additional application specific Tags can also be added, e.g., the name of the service using the connection.
func (a *connManager) Connect(tags ...string) (*ManagedConn, error) {
	a.mutex.Lock()
	defer a.mutex.Unlock()
	nc, err := nats.Connect(nats.DefaultURL, DefaultConnectTimeout, AlwaysReconnect)
	if err != nil {
		return nil, err
	}
	connId := nuid.Next()
	managedConn := NewManagedConn(connId, nc, tags)
	a.conns[connId] = managedConn

	nc.SetClosedHandler(func(conn *nats.Conn) {
		logger.Info().Str(logging.EVENT, EVENT_CONN_CLOSED).Str(CONN_ID, connId).Msg("")
		go func() {
			a.mutex.Lock()
			defer a.mutex.Unlock()
			delete(a.conns, connId)
			logger.Info().Str(logging.EVENT, EVENT_CONN_CLOSED).Str(CONN_ID, connId).Msg("deleted")
		}()
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
