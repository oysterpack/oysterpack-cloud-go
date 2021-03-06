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
	"time"

	"github.com/nats-io/go-nats"
	"github.com/nats-io/nuid"
	"github.com/oysterpack/oysterpack.go/pkg/logging"
	"github.com/oysterpack/oysterpack.go/pkg/messaging"
	"github.com/prometheus/client_golang/prometheus"
)

// NewManagedConn factory method
func NewManagedConn(cluster messaging.ClusterName, connId string, conn *nats.Conn, tags []string) *ManagedConn {
	return NewManagedConnWithConfig(cluster, connId, conn, tags, ManagedConnConfig{
		FlushTimeout: 5 * time.Second,
	})
}

// NewManagedConnWithConfig factory method
func NewManagedConnWithConfig(cluster messaging.ClusterName, connId string, conn *nats.Conn, tags []string, config ManagedConnConfig) *ManagedConn {
	return &ManagedConn{
		Conn:               conn,
		id:                 connId,
		created:            time.Now(),
		tags:               tags,
		cluster:            cluster,
		reconnectedCounter: reconnectedCounter.WithLabelValues(cluster.String()),
		errorCounter:       errorCounter.WithLabelValues(cluster.String()),
		topicSubscriptions: newTopicSubscriptions(),
		queueSubscriptions: newQueueSubscriptions(),
		publishers:         &publishers{topicPublishers: map[messaging.Topic]messaging.Publisher{}},
		config:             config,
	}
}

// ManagedConn represents a managed NATS connection
// It tracks connection lifecycle events and collects metrics.
type ManagedConn struct {
	mutex sync.RWMutex

	*nats.Conn

	id      string
	created time.Time
	tags    []string
	cluster messaging.ClusterName

	lastReconnectTime time.Time

	disconnects        int
	lastDisconnectTime time.Time

	errors        int
	lastErrorTime time.Time

	reconnectedCounter prometheus.Counter
	errorCounter       prometheus.Counter

	*topicSubscriptions
	*queueSubscriptions
	*publishers

	config ManagedConnConfig
}

// ManagedConnConfig ManagedConn config
type ManagedConnConfig struct {
	FlushTimeout time.Duration
}

type publishers struct {
	sync.RWMutex
	topicPublishers map[messaging.Topic]messaging.Publisher
}

func (a *publishers) publisher(topic messaging.Topic) messaging.Publisher {
	a.RLock()
	defer a.RUnlock()
	return a.topicPublishers[topic]
}

// ID is the unique id assigned to the connection for tracking purposes
func (a *ManagedConn) ID() string {
	return a.id
}

// Cluster returns the name of the cluster that the connection belongs to
func (a *ManagedConn) Cluster() messaging.ClusterName {
	return a.cluster
}

// Tags are used for tracking connections based on application needs
func (a *ManagedConn) Tags() []string {
	a.mutex.RLock()
	defer a.mutex.RUnlock()
	return a.tags
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
	a.reconnectedCounter.Inc()
	a.lastReconnectTime = time.Now()
	event := logger.Info().Str(logging.EVENT, EVENT_CONN_RECONNECT).Str(CONN_ID, a.id).Uint64(RECONNECTS, a.Reconnects)
	if len(a.tags) > 0 {
		event.Strs(CONN_TAGS, a.tags)
	}
	event.Msg("")
	// flushing will ensure the cluster is immediately made aware of any subscriptions that are active on this connection
	a.Flush()
}

func (a *ManagedConn) subscriptionError(subscription *nats.Subscription, err error) {
	a.errorCounter.Inc()
	a.errors++
	a.lastErrorTime = time.Now()

	event := logger.Error().Str(logging.EVENT, EVENT_CONN_SUBSCRIBER_ERR).Str(CONN_ID, a.id).Err(err).Bool(SUBSCRIPTION_VALID, subscription.IsValid())
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

// ConnInfo returns information regarding the conn
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

		SubscriptionCount:      len(a.topicSubscriptions.subscriptions),
		QueueSubscriptionCount: len(a.queueSubscriptions.subscriptions),
	}
}

func (a *ManagedConn) String() string {
	bytes, _ := json.Marshal(a.ConnInfo())
	return string(bytes)
}

// TopicSubscribe creates a Subscription
// The Subscription is registered and tracked for metrics collection purposes
func (a *ManagedConn) TopicSubscribe(topic messaging.Topic, settings *messaging.SubscriptionSettings) (messaging.Subscription, error) {
	if err := checkSubscriptionSettings(settings); err != nil {
		return nil, err
	}
	counter := topicMsgsReceivedCounter.WithLabelValues(a.cluster.String(), string(topic))
	c := make(chan *messaging.Message)
	sub, err := a.Subscribe(string(topic), func(msg *nats.Msg) {
		counter.Inc()
		sendMessage(msg, c)
	})
	if err != nil {
		return nil, err
	}
	// based on testing within a cluster, in order for the cluster to be immediately aware of the subscription, we need to flush the connection to notify the server
	if err := a.Flush(); err != nil {
		return nil, err
	}
	if settings != nil && settings.PendingLimits != nil {
		sub.SetPendingLimits(settings.MsgLimit, settings.BytesLimit)
	}
	topicSubscription := &subscription{id: nuid.Next(), cluster: a.Cluster(), sub: sub, c: c, unsubscribed: func(subscription *subscription) {
		go a.topicSubscriptions.remove(subscription.id)
		closeMessageChannel(c)
	}}
	a.topicSubscriptions.add(topicSubscription)
	return topicSubscription, nil
}

// Flush will flush the connection using the configured timeout : ManagedConnConfig.FlushTimeout
// If it was not configured, then it will default to 5 seconds.
func (a *ManagedConn) Flush() error {
	if a.config.FlushTimeout <= 0 {
		a.config.FlushTimeout = 5 * time.Second
	}
	return a.Conn.FlushTimeout(a.config.FlushTimeout)
}

func closeMessageChannel(c chan *messaging.Message) {
	// we can't close the channel, because only the sending gorouting can close the channel
	// if we close the channel, then that would cause a race condition

	// instead, in case there are goroutines blocked on this channel, we'll send nil msg
	// a nil message signals that the subscription is no longer valid
	sendMessage(nil, c)
}

func sendMessage(msg *nats.Msg, c chan *messaging.Message) {
	defer func() {
		if p := recover(); p != nil {
			logger.Warn().Msg("Message was received after the channel was closed")
		}
	}()
	c <- toMessage(msg)
}

// TopicQueueSubscribe creates a QueueSubscription
// The QueueSubscription is registered and tracked for metrics collection purposes
func (a *ManagedConn) TopicQueueSubscribe(topic messaging.Topic, queue messaging.Queue, settings *messaging.SubscriptionSettings) (messaging.QueueSubscription, error) {
	if err := checkSubscriptionSettings(settings); err != nil {
		return nil, err
	}
	counter := queueMsgsReceivedCounter.WithLabelValues(a.cluster.String(), string(topic), string(queue))
	c := make(chan *messaging.Message)
	sub, err := a.QueueSubscribe(string(topic), string(queue), func(msg *nats.Msg) {
		counter.Inc()
		sendMessage(msg, c)
	})
	if err != nil {
		return nil, err
	}
	// based on testing within a cluster, in order for the cluster to be immediately aware of the subscription, we need to flush the connection to notify the server
	if err := a.Flush(); err != nil {
		return nil, err
	}
	if settings != nil && settings.PendingLimits != nil {
		sub.SetPendingLimits(settings.MsgLimit, settings.BytesLimit)
	}
	id := nuid.Next()
	topicSubscription := &subscription{id: id, cluster: a.Cluster(), sub: sub, c: c, unsubscribed: func(subscription *subscription) {
		go a.queueSubscriptions.remove(subscription.id)
		closeMessageChannel(c)
	}}
	qSubscription := &queueSubscription{Subscription: topicSubscription, queue: queue}
	a.queueSubscriptions.add(qSubscription)
	return qSubscription, nil
}

// Publisher will return a Publisher for the specified topic.
// Publishers are cached per topic per connection.
func (a *ManagedConn) Publisher(topic messaging.Topic) (messaging.Publisher, error) {
	if pub := a.publishers.publisher(topic); pub != nil {
		return pub, nil
	}

	if err := topic.Validate(); err != nil {
		return nil, err
	}

	if a.IsClosed() {
		return nil, messaging.ErrConnectionIsClosed
	}

	a.publishers.Lock()
	defer a.publishers.Unlock()

	// in case the publisher was created since we last checked
	if pub := a.publishers.topicPublishers[topic]; pub != nil {
		return pub, nil
	}

	pub := NewPublisher(a, topic)
	a.publishers.topicPublishers[topic] = pub
	return pub, nil
}
