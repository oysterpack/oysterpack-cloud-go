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

	"fmt"

	"github.com/nats-io/go-nats"
	"github.com/nats-io/nuid"
	"github.com/oysterpack/oysterpack.go/pkg/commons"
	"github.com/oysterpack/oysterpack.go/pkg/commons/collections/sets"
	"github.com/oysterpack/oysterpack.go/pkg/logging"
	"github.com/oysterpack/oysterpack.go/pkg/messaging"
	"github.com/oysterpack/oysterpack.go/pkg/metrics"
	"github.com/oysterpack/oysterpack.go/pkg/service"
	"github.com/prometheus/client_golang/prometheus"
)

// ConnManager tracks connections that are created through it.
// The following connection lifecycle events are tracked:
//
// 1. Connection disconnects
// 2. Connection reconnects
// 3. Connection errors
// 4. Connection closed - when closed, it is no longer tracked and removed async from its managed list
type ConnManager interface {
	// ClusterName returns the name of the NATS cluster we are connecting to/
	// The name is a logical name from the application's perspective.
	Cluster() messaging.ClusterName

	Connect(tags ...string) (conn *ManagedConn, err error)

	// ManagedConn looks up a conn by its id.
	// The connection will not be present it is closed.
	ManagedConn(id string) *ManagedConn

	// ManagedConns returns all conns that have matching tags
	// If no tag filters are provided, then all are returned.
	ManagedConns(tags ...string) []*ManagedConn

	ConnCount() int

	ConnInfo(id string) *ConnInfo

	ConnInfos() []*ConnInfo

	CloseAll()

	ConnectedCount() int

	DisconnectedCount() int

	// TotalMsgsIn returns the total number of messages that have been received on all current connections
	TotalMsgsIn() uint64

	// TotalMsgsOut returns the total number of messages that have been sent on all current connections
	TotalMsgsOut() uint64

	// TotalBytesIn returns the total number of bytes that have been received on all current connections
	TotalBytesIn() uint64

	// TotalBytesOut returns the total number of bytes that have been sent on all current connections
	TotalBytesOut() uint64

	HealthChecks() []metrics.HealthCheck

	prometheus.Collector

	// GaugeMetricDescs lists all of the gauge metrics that are collected on demand
	GaugeMetricDescs() []*metrics.GaugeVecDesc

	// PublisherCountsPerTopic returns the number of publishers per topic
	PublisherCountsPerTopic() map[messaging.Topic]int

	// TopicSubscriberMetrics returns the number of subscribers per topic
	TopicSubscriberMetrics() map[messaging.Topic]*SubscriptionMetrics

	// QueueSubscriberMetrics returns the number of subscribers per topic queue
	QueueSubscriberMetrics() map[TopicQueueKey]*QueueSubscriptionMetrics
}

// DefaultOptions that are applied when creating a new ConnManager
func DefaultOptions() nats.Options {
	options := nats.GetDefaultOptions()
	DefaultConnectTimeout(&options)
	DefaultReConnectTimeout(&options)
	AlwaysReconnect(&options)
	return options
}

// TODO : define explicit config which builds a ConnManagerSettings
// see https://godoc.org/github.com/nats-io/go-nats#Options

// ConnManagerSettings are used to create new ConnManager instances
type ConnManagerSettings struct {
	messaging.ClusterName
	Options []nats.Option
}

// NewConnManager factory method.
// Default connection options are : DefaultConnectTimeout, DefaultReConnectTimeout, AlwaysReconnect
func NewConnManager(settings *ConnManagerSettings) ConnManager {
	return newConnManager(settings)
}

func newConnManager(settings *ConnManagerSettings) *connManager {
	if err := settings.ClusterName.Validate(); err != nil {
		logger.Panic().Err(err).Msg("Failed to create ConnManager")
	}

	connOptions := DefaultOptions()
	for _, option := range settings.Options {
		if err := option(&connOptions); err != nil {
			logger.Panic().Err(err).Msg("Failed to apply option")
		}
	}

	connMgr := &connManager{
		cluster: settings.ClusterName,
		options: connOptions,

		createdCounter: createdCounter.WithLabelValues(settings.ClusterName.String()),
		closedCounter:  closedCounter.WithLabelValues(settings.ClusterName.String()),
	}

	connMgr.connCountDesc = connMgr.addGaugeDesc(ConnCountOpts)
	connMgr.notConnectedCountDesc = connMgr.addGaugeDesc(NotConnectedCountOpts)

	connMgr.msgsInDesc = connMgr.addGaugeDesc(MsgsInGauge)
	connMgr.msgsOutDesc = connMgr.addGaugeDesc(MsgsOutGauge)
	connMgr.bytesInDesc = connMgr.addGaugeDesc(BytesInGauge)
	connMgr.bytesOutDesc = connMgr.addGaugeDesc(BytesOutGauge)

	connMgr.topicSubscriberCount = connMgr.addGaugeDesc(TopicSubscriberCount)
	connMgr.topicPendingMessages = connMgr.addGaugeDesc(TopicPendingMessages)
	connMgr.topicPendingBytes = connMgr.addGaugeDesc(TopicPendingBytes)
	connMgr.topicMaxPendingMessages = connMgr.addGaugeDesc(TopicMaxPendingMessages)
	connMgr.topicMaxPendingBytes = connMgr.addGaugeDesc(TopicMaxPendingBytes)
	connMgr.topicMessagesDropped = connMgr.addGaugeDesc(TopicMessagesDropped)
	connMgr.topicMessagesDelivered = connMgr.addGaugeDesc(TopicMessagesDelivered)

	connMgr.queueSubscriberCount = connMgr.addGaugeDesc(QueueSubscriberCount)
	connMgr.queuePendingMessages = connMgr.addGaugeDesc(QueuePendingMessages)
	connMgr.queuePendingBytes = connMgr.addGaugeDesc(QueuePendingBytes)
	connMgr.queueMaxPendingMessages = connMgr.addGaugeDesc(QueueMaxPendingMessages)
	connMgr.queueMaxPendingBytes = connMgr.addGaugeDesc(QueueMaxPendingBytes)
	connMgr.queueMessagesDropped = connMgr.addGaugeDesc(QueueMessagesDropped)
	connMgr.queueMessagesDelivered = connMgr.addGaugeDesc(QueueMessagesDelivered)

	connMgr.publisherCount = connMgr.addGaugeDesc(PublisherCount)

	connMgr.init()
	metrics.Registry.MustRegister(connMgr)
	return connMgr
}

// implements prometheus.Collector, i.e., it collects connection related metrics
type connManager struct {
	cluster messaging.ClusterName
	options nats.Options

	*service.RestartableService

	mutex sync.RWMutex

	n     nuid.NUID
	conns map[string]*ManagedConn

	healthChecks []metrics.HealthCheck

	createdCounter prometheus.Counter
	closedCounter  prometheus.Counter

	connCountDesc         *prometheus.Desc
	notConnectedCountDesc *prometheus.Desc
	msgsInDesc            *prometheus.Desc
	msgsOutDesc           *prometheus.Desc
	bytesInDesc           *prometheus.Desc
	bytesOutDesc          *prometheus.Desc

	topicSubscriberCount    *prometheus.Desc
	topicPendingMessages    *prometheus.Desc
	topicPendingBytes       *prometheus.Desc
	topicMaxPendingMessages *prometheus.Desc
	topicMaxPendingBytes    *prometheus.Desc
	topicMessagesDelivered  *prometheus.Desc
	topicMessagesDropped    *prometheus.Desc

	queueSubscriberCount    *prometheus.Desc
	queuePendingMessages    *prometheus.Desc
	queuePendingBytes       *prometheus.Desc
	queueMaxPendingMessages *prometheus.Desc
	queueMaxPendingBytes    *prometheus.Desc
	queueMessagesDelivered  *prometheus.Desc
	queueMessagesDropped    *prometheus.Desc

	publisherCount *prometheus.Desc

	gaugeDescs []*metrics.GaugeVecDesc
}

func (a *connManager) GaugeMetricDescs() []*metrics.GaugeVecDesc {
	return a.gaugeDescs
}

func (a *connManager) addGaugeDesc(opts *metrics.GaugeVecOpts) *prometheus.Desc {
	desc := prometheus.NewDesc(metrics.GaugeFQName(opts.GaugeOpts), opts.GaugeOpts.Help, opts.Labels, opts.GaugeOpts.ConstLabels)
	a.gaugeDescs = append(a.gaugeDescs, &metrics.GaugeVecDesc{Opts: opts, Desc: desc})
	return desc
}

// Describe implements prometheus.Collector
func (a *connManager) Describe(ch chan<- *prometheus.Desc) {
	ch <- a.connCountDesc
	ch <- a.notConnectedCountDesc

	for _, desc := range a.gaugeDescs {
		ch <- desc.Desc
	}
}

func (a *connManager) PublisherCountsPerTopic() map[messaging.Topic]int {
	a.mutex.RLock()
	defer a.mutex.RUnlock()
	topicPublisherCounts := map[messaging.Topic]int{}
	for _, conn := range a.conns {
		for key := range conn.publishers.topicPublishers {
			topicPublisherCounts[key]++
		}
	}

	return topicPublisherCounts
}

func (a *connManager) TopicSubscriberMetrics() map[messaging.Topic]*SubscriptionMetrics {
	a.mutex.RLock()
	defer a.mutex.RUnlock()

	topicSubscriptionMetrics := map[messaging.Topic]*SubscriptionMetrics{}

	for _, conn := range a.conns {
		for topic, metrics := range conn.topicSubscriptions.collectMetrics() {
			agg, exists := topicSubscriptionMetrics[topic]
			if exists {
				agg.add(metrics)
			} else {
				topicSubscriptionMetrics[topic] = metrics
			}
		}
	}

	return topicSubscriptionMetrics
}

func (a *connManager) QueueSubscriberMetrics() map[TopicQueueKey]*QueueSubscriptionMetrics {
	a.mutex.RLock()
	defer a.mutex.RUnlock()

	queueSubscriptionMetrics := map[TopicQueueKey]*QueueSubscriptionMetrics{}
	for _, conn := range a.conns {
		for key, metrics := range conn.queueSubscriptions.collectMetrics() {
			agg, exists := queueSubscriptionMetrics[key]
			if exists {
				agg.add(metrics.SubscriptionMetrics)
			} else {
				queueSubscriptionMetrics[key] = metrics
			}
		}
	}
	return queueSubscriptionMetrics
}

// Collect implements prometheus.Collector
func (a *connManager) Collect(ch chan<- prometheus.Metric) {
	a.mutex.RLock()
	defer a.mutex.RUnlock()

	ch <- a.createdCounter
	ch <- a.closedCounter

	var msgsIn, msgsOut, bytesIn, bytesOut uint64
	notConnectedCount := 0
	topicSubscriptionMetrics := map[messaging.Topic]*SubscriptionMetrics{}
	queueSubscriptionMetrics := map[TopicQueueKey]*QueueSubscriptionMetrics{}
	topicPublisherCounts := map[messaging.Topic]int{}

	for _, conn := range a.conns {
		msgsIn += conn.InMsgs
		msgsOut += conn.OutMsgs
		bytesIn += conn.InBytes
		bytesOut += conn.OutBytes
		if !conn.IsConnected() {
			notConnectedCount++
		}

		for topic, metrics := range conn.topicSubscriptions.collectMetrics() {
			agg, exists := topicSubscriptionMetrics[topic]
			if exists {
				agg.add(metrics)
			} else {
				topicSubscriptionMetrics[topic] = metrics
			}
		}

		for key, metrics := range conn.queueSubscriptions.collectMetrics() {
			agg, exists := queueSubscriptionMetrics[key]
			if exists {
				agg.add(metrics.SubscriptionMetrics)
			} else {
				queueSubscriptionMetrics[key] = metrics
			}
		}

		for key := range conn.publishers.topicPublishers {
			topicPublisherCounts[key]++
		}
	}

	ch <- prometheus.MustNewConstMetric(a.connCountDesc,
		prometheus.GaugeValue, float64(len(a.conns)), a.cluster.String(),
	)
	ch <- prometheus.MustNewConstMetric(a.notConnectedCountDesc,
		prometheus.GaugeValue, float64(notConnectedCount), a.cluster.String(),
	)
	ch <- prometheus.MustNewConstMetric(a.msgsInDesc,
		prometheus.GaugeValue, float64(msgsIn), a.cluster.String(),
	)
	ch <- prometheus.MustNewConstMetric(a.msgsOutDesc,
		prometheus.GaugeValue, float64(msgsOut), a.cluster.String(),
	)
	ch <- prometheus.MustNewConstMetric(a.bytesInDesc,
		prometheus.GaugeValue, float64(bytesIn), a.cluster.String(),
	)
	ch <- prometheus.MustNewConstMetric(a.bytesOutDesc,
		prometheus.GaugeValue, float64(bytesOut), a.cluster.String(),
	)

	a.reportTopicSubscriptionMetrics(ch, topicSubscriptionMetrics)
	a.reportQueueSubscriptionMetrics(ch, queueSubscriptionMetrics)

	for topic, count := range topicPublisherCounts {
		ch <- prometheus.MustNewConstMetric(a.publisherCount,
			prometheus.GaugeValue, float64(count), a.cluster.String(), string(topic),
		)
	}
}

func (a *connManager) reportTopicSubscriptionMetrics(ch chan<- prometheus.Metric, topicSubscriptionMetrics map[messaging.Topic]*SubscriptionMetrics) {
	for _, metrics := range topicSubscriptionMetrics {
		ch <- prometheus.MustNewConstMetric(a.topicSubscriberCount,
			prometheus.GaugeValue, float64(metrics.SubscriberCount), a.cluster.String(), string(metrics.Topic),
		)
		ch <- prometheus.MustNewConstMetric(a.topicPendingMessages,
			prometheus.GaugeValue, float64(metrics.PendingMsgs), a.cluster.String(), string(metrics.Topic),
		)
		ch <- prometheus.MustNewConstMetric(a.topicPendingBytes,
			prometheus.GaugeValue, float64(metrics.PendingBytes), a.cluster.String(), string(metrics.Topic),
		)
		ch <- prometheus.MustNewConstMetric(a.topicMaxPendingMessages,
			prometheus.GaugeValue, float64(metrics.PendingMsgsMax), a.cluster.String(), string(metrics.Topic),
		)
		ch <- prometheus.MustNewConstMetric(a.topicMaxPendingBytes,
			prometheus.GaugeValue, float64(metrics.PendingBytesMax), a.cluster.String(), string(metrics.Topic),
		)
		ch <- prometheus.MustNewConstMetric(a.topicMessagesDropped,
			prometheus.GaugeValue, float64(metrics.Dropped), a.cluster.String(), string(metrics.Topic),
		)
		ch <- prometheus.MustNewConstMetric(a.topicMessagesDelivered,
			prometheus.GaugeValue, float64(metrics.Delivered), a.cluster.String(), string(metrics.Topic),
		)
	}
}

func (a *connManager) reportQueueSubscriptionMetrics(ch chan<- prometheus.Metric, queueSubscriptionMetrics map[TopicQueueKey]*QueueSubscriptionMetrics) {
	for _, metrics := range queueSubscriptionMetrics {
		ch <- prometheus.MustNewConstMetric(a.queueSubscriberCount,
			prometheus.GaugeValue, float64(metrics.SubscriberCount), a.cluster.String(), string(metrics.Topic), string(metrics.Queue),
		)
		ch <- prometheus.MustNewConstMetric(a.queuePendingMessages,
			prometheus.GaugeValue, float64(metrics.PendingMsgs), a.cluster.String(), string(metrics.Topic), string(metrics.Queue),
		)
		ch <- prometheus.MustNewConstMetric(a.queuePendingBytes,
			prometheus.GaugeValue, float64(metrics.PendingBytes), a.cluster.String(), string(metrics.Topic), string(metrics.Queue),
		)
		ch <- prometheus.MustNewConstMetric(a.queueMaxPendingMessages,
			prometheus.GaugeValue, float64(metrics.PendingMsgsMax), a.cluster.String(), string(metrics.Topic), string(metrics.Queue),
		)
		ch <- prometheus.MustNewConstMetric(a.queueMaxPendingBytes,
			prometheus.GaugeValue, float64(metrics.PendingBytesMax), a.cluster.String(), string(metrics.Topic), string(metrics.Queue),
		)
		ch <- prometheus.MustNewConstMetric(a.queueMessagesDropped,
			prometheus.GaugeValue, float64(metrics.Dropped), a.cluster.String(), string(metrics.Topic), string(metrics.Queue),
		)
		ch <- prometheus.MustNewConstMetric(a.queueMessagesDelivered,
			prometheus.GaugeValue, float64(metrics.Delivered), a.cluster.String(), string(metrics.Topic), string(metrics.Queue),
		)
	}
}

func (a *connManager) Cluster() messaging.ClusterName {
	return a.cluster
}

func (a *connManager) HealthChecks() []metrics.HealthCheck {
	return a.healthChecks
}

// Connect creates a new managed NATS connection.
// Tags are used to help identify how the connection is being used by the app. Tags are currently used to augment logging.
//
// Connections are configured to always reconnect.
func (a *connManager) Connect(tags ...string) (*ManagedConn, error) {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	nc, err := a.options.Connect()
	if err != nil {
		return nil, err
	}
	connId := nuid.Next()
	managedConn := NewManagedConn(a.cluster, connId, nc, tags)
	a.conns[connId] = managedConn

	nc.SetClosedHandler(func(conn *nats.Conn) {
		a.closedCounter.Inc()
		logger.Info().Str(logging.EVENT, EVENT_CONN_CLOSED).Str(CONN_ID, connId).Msg("")
		a.mutex.Lock()
		defer a.mutex.Unlock()
		delete(a.conns, connId)
		logger.Info().Str(logging.EVENT, EVENT_CONN_CLOSED).Str(CONN_ID, connId).Msg("deleted")

		if a.options.ClosedCB != nil {
			a.options.ClosedCB(conn)
		}
	})

	nc.SetDisconnectHandler(func(conn *nats.Conn) {
		managedConn.disconnected()
		if a.options.DisconnectedCB != nil {
			a.options.DisconnectedCB(conn)
		}
	})
	nc.SetReconnectHandler(func(conn *nats.Conn) {
		managedConn.reconnected()
		if a.options.ReconnectedCB != nil {
			a.options.ReconnectedCB(conn)
		}
	})
	nc.SetDiscoveredServersHandler(func(conn *nats.Conn) {
		managedConn.discoveredServers()
		if a.options.DisconnectedCB != nil {
			a.options.DisconnectedCB(conn)
		}
	})
	nc.SetErrorHandler(func(conn *nats.Conn, subscription *nats.Subscription, err error) {
		managedConn.subscriptionError(subscription, err)
		if a.options.AsyncErrorCB != nil {
			a.options.AsyncErrorCB(conn, subscription, err)
		}
	})

	a.createdCounter.Inc()

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

	// register healthchecks
	if len(a.healthChecks) == 0 {
		a.healthChecks = []metrics.HealthCheck{
			metrics.NewHealthCheckVector(connectivityHealthCheck, runinterval,
				func() error {
					total := a.ConnCount()
					connected := a.ConnectedCount()
					if connected != total {
						return fmt.Errorf("%d / %d connections are disconnected", total-connected, total)
					}
					return nil
				}, []string{a.cluster.String()}),
		}
	}
}

// TotalMsgsIn returns the total number of messages that have been received on all current connections.
func (a *connManager) TotalMsgsIn() uint64 {
	a.mutex.RLock()
	defer a.mutex.RUnlock()
	var count uint64
	for _, conn := range a.conns {
		count += conn.InMsgs
	}
	return count
}

// TotalMsgsOut returns the total number of messages that have been sent on all current connections.
func (a *connManager) TotalMsgsOut() uint64 {
	a.mutex.RLock()
	defer a.mutex.RUnlock()
	var count uint64
	for _, conn := range a.conns {
		count += conn.OutMsgs
	}
	return count
}

// TotalMsgsIn returns the total number of bytes that have been received on all current connections.
func (a *connManager) TotalBytesIn() uint64 {
	a.mutex.RLock()
	defer a.mutex.RUnlock()
	var count uint64
	for _, conn := range a.conns {
		count += conn.InBytes
	}
	return count
}

// TotalMsgsOut returns the total number of bytes that have been sent on all current connections.
func (a *connManager) TotalBytesOut() uint64 {
	a.mutex.RLock()
	defer a.mutex.RUnlock()
	var count uint64
	for _, conn := range a.conns {
		count += conn.OutBytes
	}
	return count
}

func (a *connManager) CloseAll() {
	a.mutex.Lock()
	defer a.mutex.Unlock()
	for _, nc := range a.conns {
		func() {
			defer commons.IgnorePanic()
			nc.Conn.Close()
		}()
	}
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

func (a *connManager) ConnectedCount() (count int) {
	a.mutex.RLock()
	defer a.mutex.RUnlock()
	for _, c := range a.conns {
		if c.Conn.IsConnected() {
			count++
		}
	}
	return
}

func (a *connManager) DisconnectedCount() (count int) {
	a.mutex.RLock()
	defer a.mutex.RUnlock()
	for _, c := range a.conns {
		if !c.Conn.IsConnected() {
			count++
		}
	}
	return
}

func (a *connManager) ManagedConns(tags ...string) (conns []*ManagedConn) {
	a.mutex.RLock()
	defer a.mutex.RUnlock()
	hasTags := len(tags) > 0
CONN_LOOP:
	for _, c := range a.conns {
		if hasTags {
			if len(tags) > len(c.tags) {
				continue
			}
			tagSet := sets.NewStrings()
			tagSet.AddAll(c.tags...)
			for _, tag := range tags {
				if !tagSet.Contains(tag) {
					continue CONN_LOOP
				}
			}
			conns = append(conns, c)
		} else {
			conns = append(conns, c)
		}
	}
	return
}
