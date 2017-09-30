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
	"sync"
	"time"

	"github.com/nats-io/go-nats"
	"github.com/oysterpack/oysterpack.go/pkg/logging"
)

type ManagedConn struct {
	mutex sync.RWMutex

	*nats.Conn

	id      string
	created time.Time
	tags    []string

	lastReconnectTime time.Time

	disconnects        int
	lastDisconnectTime time.Time

	errors        int
	lastErrorTime time.Time
}

func (a *ManagedConn) ID() string {
	return a.id
}

func (a *ManagedConn) Created() time.Time {
	a.mutex.RLock()
	defer a.mutex.RUnlock()
	return a.created
}

func (a *ManagedConn) Tags() []string {
	a.mutex.RLock()
	defer a.mutex.RUnlock()
	return a.tags
}

func (a *ManagedConn) Disconnects() int {
	a.mutex.RLock()
	defer a.mutex.RUnlock()
	return a.disconnects
}

func (a *ManagedConn) LastDisconnectTime() time.Time {
	a.mutex.RLock()
	defer a.mutex.RUnlock()
	return a.lastDisconnectTime
}

func (a *ManagedConn) LastReconnectTime() time.Time {
	a.mutex.RLock()
	defer a.mutex.RUnlock()
	return a.lastReconnectTime
}

func (a *ManagedConn) LastErrorTime() time.Time {
	a.mutex.RLock()
	defer a.mutex.RUnlock()
	return a.lastErrorTime
}

// Errors returns the number of errors that have occurred on this connection
func (a *ManagedConn) Errors() int {
	a.mutex.RLock()
	defer a.mutex.RUnlock()
	return a.errors
}

// updates are protected by a mutex to make changes concurrency safe
func (a *ManagedConn) disconnected() {
	disconnectedCounter.Inc()
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
	reconnectedCounter.Inc()
	a.lastReconnectTime = time.Now()
	event := logger.Info().Str(logging.EVENT, EVENT_CONN_RECONNECT).Str(CONN_ID, a.id).Uint64(RECONNECTS, a.Reconnects)
	if len(a.tags) > 0 {
		event.Strs(CONN_TAGS, a.tags)
	}
	event.Msg("")
}

func (a *ManagedConn) subscriptionError(subscription *nats.Subscription, err error) {
	errorCounter.Inc()
	a.errors++
	a.lastErrorTime = time.Now()

	event := logger.Error().Str(logging.EVENT, EVENT_CONN_ERR).Str(CONN_ID, a.id).Err(err).Bool(SUBSCRIPTION_VALID, subscription.IsValid())
	if len(a.tags) > 0 {
		event.Strs(CONN_TAGS, a.tags)
	}

	subscriptionErrors := []error{}
	// pending does not apply to channel based subscriptions
	// if a message is not ready to be received on the channel, then it is dropped.
	if subscription.Type() != nats.ChanSubscription {
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

		Statistics:        a.Stats(),
		LastReconnectTime: a.lastReconnectTime,

		Disconnects:        a.disconnects,
		LastDisconnectTime: a.lastDisconnectTime,

		Errors:        a.errors,
		LastErrorTime: a.lastErrorTime,

		Status: a.Conn.Status(),
	}
}

// NewManagedConn factory method
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
