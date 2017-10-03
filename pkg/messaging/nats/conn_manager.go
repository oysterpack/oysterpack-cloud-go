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

	ManagedConn(id string) *ManagedConn

	ConnCount() int

	ConnInfo(id string) *ConnInfo

	ConnInfos() []*ConnInfo

	CloseAll()

	ConnectedCount() (count int, total int)

	DisconnectedCount() (count int, total int)

	// TotalMsgsIn returns the total number of messages that have been received on all current connections
	TotalMsgsIn() uint64

	// TotalMsgsOut returns the total number of messages that have been sent on all current connections
	TotalMsgsOut() uint64

	// TotalBytesIn returns the total number of bytes that have been received on all current connections
	TotalBytesIn() uint64

	// TotalBytesOut returns the total number of bytes that have been sent on all current connections
	TotalBytesOut() uint64

	HealthChecks() []metrics.HealthCheck
}

// DefaultOptions that are applied when creating a new ConnManager
func DefaultOptions() nats.Options {
	options := nats.GetDefaultOptions()
	DefaultConnectTimeout(&options)
	DefaultReConnectTimeout(&options)
	AlwaysReconnect(&options)
	return options
}

// ConnManagerSettings are used to create new ConnManager instances
type ConnManagerSettings struct {
	messaging.ClusterName
	Options []nats.Option
}

// NewConnManager factory method.
// Default connection options are : DefaultConnectTimeout, DefaultReConnectTimeout, AlwaysReconnect
func NewConnManager(settings ConnManagerSettings) ConnManager {
	return newConnManager(settings)
}

func newConnManager(settings ConnManagerSettings) *connManager {
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
		connCount:      connCount.WithLabelValues(settings.ClusterName.String()),
		closedCounter:  closedCounter.WithLabelValues(settings.ClusterName.String()),

		msgsInDesc:   prometheus.NewDesc(metrics.GaugeFQName(MsgsInGauge.GaugeOpts), MsgsInGauge.GaugeOpts.Help, MetricLabels, MsgsInGauge.GaugeOpts.ConstLabels),
		msgsOutDesc:  prometheus.NewDesc(metrics.GaugeFQName(MsgsOutGauge.GaugeOpts), MsgsOutGauge.GaugeOpts.Help, MetricLabels, MsgsOutGauge.GaugeOpts.ConstLabels),
		bytesInDesc:  prometheus.NewDesc(metrics.GaugeFQName(BytesInGauge.GaugeOpts), BytesInGauge.GaugeOpts.Help, MetricLabels, BytesInGauge.GaugeOpts.ConstLabels),
		bytesOutDesc: prometheus.NewDesc(metrics.GaugeFQName(BytesOutGauge.GaugeOpts), BytesOutGauge.GaugeOpts.Help, MetricLabels, BytesOutGauge.GaugeOpts.ConstLabels),
	}
	connMgr.init()
	metrics.Registry.MustRegister(connMgr)
	return connMgr
}

// ManagedConn is a managed NATS connection.

type connManager struct {
	cluster messaging.ClusterName
	options nats.Options

	*service.RestartableService

	mutex sync.RWMutex

	n     nuid.NUID
	conns map[string]*ManagedConn

	healthChecks []metrics.HealthCheck

	createdCounter prometheus.Counter
	connCount      prometheus.Gauge
	closedCounter  prometheus.Counter

	msgsInDesc   *prometheus.Desc
	msgsOutDesc  *prometheus.Desc
	bytesInDesc  *prometheus.Desc
	bytesOutDesc *prometheus.Desc
}

// Describe implements prometheus.Collector
func (a *connManager) Describe(ch chan<- *prometheus.Desc) {
	ch <- a.msgsInDesc
	ch <- a.msgsOutDesc
	ch <- a.bytesInDesc
	ch <- a.bytesOutDesc
}

// Collect implements prometheus.Collector
func (a *connManager) Collect(ch chan<- prometheus.Metric) {
	a.mutex.RLock()
	defer a.mutex.RUnlock()

	var msgsIn, msgsOut, bytesIn, bytesOut uint64

	for _, conn := range a.conns {
		msgsIn += conn.InMsgs
		msgsOut += conn.OutMsgs
		bytesIn += conn.InBytes
		bytesOut += conn.OutBytes
	}

	ch <- prometheus.MustNewConstMetric(
		a.msgsInDesc,
		prometheus.GaugeValue,
		float64(msgsIn),
		a.cluster.String(),
	)

	ch <- prometheus.MustNewConstMetric(
		a.msgsOutDesc,
		prometheus.GaugeValue,
		float64(msgsOut),
		a.cluster.String(),
	)

	ch <- prometheus.MustNewConstMetric(
		a.bytesInDesc,
		prometheus.GaugeValue,
		float64(bytesIn),
		a.cluster.String(),
	)

	ch <- prometheus.MustNewConstMetric(
		a.bytesOutDesc,
		prometheus.GaugeValue,
		float64(bytesOut),
		a.cluster.String(),
	)
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
//
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
		a.connCount.Dec()
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
	a.connCount.Inc()

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

	if len(a.healthChecks) == 0 {
		a.healthChecks = []metrics.HealthCheck{
			metrics.NewHealthCheckVector(connectivityHealthCheck, runinterval,
				service.SkipHealthCheckDuringAppShutdown(func() error {
					connected, total := a.ConnectedCount()
					if connected != total {
						return fmt.Errorf("%d / %d connections are disconnected", total-connected, total)
					}
					return nil
				}), []string{a.cluster.String()}),
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
		nc.Conn.SetClosedHandler(func(conn *nats.Conn) {
			a.connCount.Dec()
			a.closedCounter.Inc()
			logger.Info().Str(logging.EVENT, EVENT_CONN_CLOSED).Str(CONN_ID, nc.ID()).Msg("CloseAll")
		})
		func() {
			defer commons.IgnorePanic()
			nc.Conn.Close()
		}()
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
