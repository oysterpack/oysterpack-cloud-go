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

package nats_test

import (
	"testing"

	"github.com/oysterpack/oysterpack.go/pkg/messaging/nats"
	"github.com/oysterpack/oysterpack.go/pkg/messaging/natstest"
	"github.com/oysterpack/oysterpack.go/pkg/metrics"

	"encoding/json"
	"strings"
	"time"

	natsio "github.com/nats-io/go-nats"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_model/go"
	dto "github.com/prometheus/client_model/go"
)

func skipTestConnManager_Metrics_RestartingServer(t *testing.T) {
	metrics.ResetRegistry()
	defer metrics.ResetRegistry()
	nats.RegisterMetrics()

	backup := nats.DefaultReConnectTimeout
	const ReConnectTimeout = 10 * time.Millisecond
	nats.DefaultReConnectTimeout = natsio.ReconnectWait(ReConnectTimeout)
	defer func() { nats.DefaultReConnectTimeout = backup }()

	server := natstest.RunServer()
	defer server.Shutdown()

	connMgr := nats.NewConnManager(TestConnManagerSettings)
	defer connMgr.CloseAll()
	pubConn := mustConnect(t, connMgr)
	subConn := mustConnect(t, connMgr)

	// restart the server
	server.Shutdown()

	for {
		if !pubConn.IsConnected() || !subConn.IsConnected() {
			break
		}
		t.Logf("server has been shutdown : waiting for connections to disconnect : pubConn.IsConnected() = %v : subConn.IsConnected() = %v", pubConn.IsConnected(), subConn.IsConnected())
	}

	for _, healthcheck := range connMgr.HealthChecks() {
		result := healthcheck.Run()
		if result.Success() {
			t.Error("*** ERROR *** healthcheck should have failed because server was shutdown.")
		}
	}

	// subscribe and publish while disconnected
	const TOPIC = "TestConnManager_Metrics"

	ch := make(chan *natsio.Msg)
	subConn.ChanSubscribe(TOPIC, ch)

	pubConn.Publish(TOPIC, []byte("TEST"))

	server = natstest.RunServer()
	defer server.Shutdown()

	// ensure connections are reconnected
	for {
		if pubConn.IsConnected() && subConn.IsConnected() {
			break
		}
		t.Logf("waiting for connections to re-connect : pubConn.IsConnected() = %v : subConn.IsConnected() = %v", pubConn.IsConnected(), subConn.IsConnected())
	}
	pubConn.Flush()

	select {
	case <-ch:
	default:
	}

	gatheredMetrics, err := metrics.Registry.Gather()
	checkMetricsAfterReconnecting(t, gatheredMetrics, pubConn, subConn)

	connMgr.CloseAll()
	// give some time for conn handler callbacks to be invoked
	time.Sleep(10 * time.Millisecond)

	gatheredMetrics, err = metrics.Registry.Gather()
	if err != nil {
		t.Fatalf("Failed to gather gatheredMetrics : %v", err)
	}

	logMetrics(t, gatheredMetrics)

	checkLifecycleRelatedMetricCounts(t, gatheredMetrics, pubConn, subConn)
	checkMsgPublishSubscribeStats(t, gatheredMetrics, pubConn, subConn)
}

func checkMsgPublishSubscribeStats(t *testing.T, gatheredMetrics []*dto.MetricFamily, pubConn *nats.ManagedConn, subConn *nats.ManagedConn) {
	t.Helper()
	t.Logf("pubConn : %v", pubConn)
	t.Logf("subConn : %v", subConn)

	if value := *gauge(gatheredMetrics, nats.MsgsInGauge).GetMetric()[0].GetGauge().Value; uint64(value) != 0 {
		t.Errorf("*** ERROR *** msgs in is wrong : %v", value)
	}
	if value := *gauge(gatheredMetrics, nats.MsgsOutGauge).GetMetric()[0].GetGauge().Value; uint64(value) != 0 {
		t.Errorf("*** ERROR *** msgs out is wrong : %v", value)
	}
	if value := *gauge(gatheredMetrics, nats.BytesInGauge).GetMetric()[0].GetGauge().Value; uint64(value) != 0 {
		t.Errorf("*** ERROR *** bytes in is wrong : %v", value)
	}
	if value := *gauge(gatheredMetrics, nats.BytesOutGauge).GetMetric()[0].GetGauge().Value; uint64(value) != 0 {
		t.Errorf("*** ERROR *** bytes out is wrong : %v", value)
	}
}

func checkLifecycleRelatedMetricCounts(t *testing.T, gatheredMetrics []*dto.MetricFamily, pubConn *nats.ManagedConn, subConn *nats.ManagedConn) {
	t.Helper()
	if value := *counter(gatheredMetrics, nats.CreatedCounterOpts).GetMetric()[0].GetCounter().Value; value != 2 {
		t.Errorf("*** ERROR *** created count is wrong : %v", value)
	}

	if value := *counter(gatheredMetrics, nats.ClosedCounterOpts).GetMetric()[0].GetCounter().Value; value != 2 {
		t.Errorf("*** ERROR *** closed count is wrong : %v", value)
	}

	if value := *gauge(gatheredMetrics, nats.ConnCountOpts).GetMetric()[0].GetGauge().Value; value != 0 {
		t.Errorf("*** ERROR *** closed count is wrong : %v", value)
	}

	if value := *counter(gatheredMetrics, nats.ReconnectedCounterOpts).GetMetric()[0].GetCounter().Value; value != 2 {
		t.Errorf("*** ERROR *** reconnects count is wrong : %v", value)
	}
}

func checkMetricsExist(t *testing.T, gatheredMetrics []*dto.MetricFamily) {
	t.Helper()
	for _, opts := range nats.ConnManagerMetrics.CounterVecOpts {
		if counter(gatheredMetrics, opts) == nil {
			t.Errorf("*** ERROR *** Counter Metric was not gathered : %v", prometheus.BuildFQName(opts.Namespace, opts.Subsystem, opts.Name))
		}
	}
	for _, opts := range nats.ConnManagerMetrics.GaugeVecOpts {
		if gauge(gatheredMetrics, opts) == nil {
			t.Errorf("*** ERROR *** Gauge Metric was not gathered : %v", prometheus.BuildFQName(opts.Namespace, opts.Subsystem, opts.Name))
		}
	}
}

func checkMetricsAfterReconnecting(t *testing.T, gatheredMetrics []*dto.MetricFamily, pubConn *nats.ManagedConn, subConn *nats.ManagedConn) {
	t.Helper()
	if value := *gauge(gatheredMetrics, nats.ConnCountOpts).GetMetric()[0].GetGauge().Value; value != 2 {
		t.Errorf("*** ERROR *** conn count is wrong : %v", value)
	}
	if value := *gauge(gatheredMetrics, nats.MsgsInGauge).GetMetric()[0].GetGauge().Value; uint64(value) != subConn.InMsgs {
		t.Errorf("*** ERROR *** msgs in is wrong : %v", value)
	}
	if value := *gauge(gatheredMetrics, nats.MsgsOutGauge).GetMetric()[0].GetGauge().Value; uint64(value) != pubConn.OutMsgs {
		t.Errorf("*** ERROR *** msgs out is wrong : %v", value)
	}
	if value := *gauge(gatheredMetrics, nats.BytesInGauge).GetMetric()[0].GetGauge().Value; uint64(value) != subConn.InBytes {
		t.Errorf("*** ERROR *** bytes in is wrong : %v", value)
	}
	if value := *gauge(gatheredMetrics, nats.BytesOutGauge).GetMetric()[0].GetGauge().Value; uint64(value) != pubConn.OutBytes {
		t.Errorf("*** ERROR *** bytes out is wrong : %v", value)
	}
}

func logMetrics(t *testing.T, gatheredMetrics []*dto.MetricFamily) {
	t.Helper()
	for _, metric := range gatheredMetrics {
		if strings.HasPrefix(*metric.Name, metrics.METRIC_NAMESPACE_OYSTERPACK) {
			jsonBytes, _ := json.MarshalIndent(metric, "", "   ")
			t.Logf("%v", string(jsonBytes))
		}
	}
}

func counter(metricFamilies []*io_prometheus_client.MetricFamily, opts *metrics.CounterVecOpts) *io_prometheus_client.MetricFamily {
	name := prometheus.BuildFQName(opts.Namespace, opts.Subsystem, opts.Name)
	for _, metric := range metricFamilies {
		if name == *metric.Name {
			return metric
		}
	}
	return nil
}

func gauge(metricFamilies []*io_prometheus_client.MetricFamily, opts *metrics.GaugeVecOpts) *io_prometheus_client.MetricFamily {
	name := prometheus.BuildFQName(opts.Namespace, opts.Subsystem, opts.Name)
	for _, metric := range metricFamilies {
		if name == *metric.Name {
			return metric
		}
	}
	return nil
}
