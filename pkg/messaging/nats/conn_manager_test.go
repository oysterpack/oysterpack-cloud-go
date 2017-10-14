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

	"time"

	"github.com/nats-io/nuid"
	"github.com/oysterpack/oysterpack.go/pkg/messaging"
	"github.com/oysterpack/oysterpack.go/pkg/messaging/nats"
	"github.com/oysterpack/oysterpack.go/pkg/messaging/natstest"
	"github.com/oysterpack/oysterpack.go/pkg/metrics"
	"github.com/prometheus/client_golang/prometheus"
)

func TestConnManager_Metrics(t *testing.T) {
	metrics.ResetRegistry()
	defer metrics.ResetRegistry()

	serverConfigs := natstest.CreateNATSServerConfigsNoTLS(1)
	servers := natstest.CreateNATSServers(t, serverConfigs)
	natstest.StartServers(servers)
	defer natstest.ShutdownServers(servers)

	connManager := nats.NewConnManager(natstest.ConnManagerSettings(serverConfigs[0]))

	metricDescs := make(chan *prometheus.Desc, 100)
	connManager.Describe(metricDescs)
	close(metricDescs)
	metricCount := 0
	for desc := range metricDescs {
		metricCount++
		t.Logf("(%d) : %s", metricCount, desc)
	}
	if metricCount != len(nats.ConnManagerMetrics.GaugeVecOpts) {
		t.Error("The metric descriptors that were returned by the ConnManager do not align with nats.ConnManagerMetrics.GaugeVecOpts")
	}

	collectMetricsWithNoActiveConnections(t, connManager)
	sendReceiveMessages(t, connManager)
	if connManager.DisconnectedCount() > 0 {
		t.Errorf("There should be no connections disconnected : %d", connManager.DisconnectedCount())
	}
	collectMetricsWithAfterSendingReceivingMessages(t, connManager)
}

func collectMetricsWithAfterSendingReceivingMessages(t *testing.T, connManager nats.ConnManager) {
	metrics := make(chan prometheus.Metric, 100)
	connManager.Collect(metrics)
	close(metrics)
	metricCount := 0
	for metric := range metrics {
		metricCount++
		t.Logf("(%d) : %s", metricCount, metric)
	}
	// we are only checking if the number of metrics returned is what's expected
	if metricCount != len(nats.ConnManagerMetrics.GaugeVecOpts) {
		t.Error("The expected number of metrics did not match : %d != %d", metricCount, len(nats.ConnManagerMetrics.GaugeVecOpts))
	}
}

func collectMetricsWithNoActiveConnections(t *testing.T, connManager nats.ConnManager) {
	metrics := make(chan prometheus.Metric, 100)
	connManager.Collect(metrics)
	close(metrics)
	metricCount := 0
	for metric := range metrics {
		metricCount++
		t.Logf("(%d) : %s", metricCount, metric)
	}
	if metricCount != len(nats.ConnectionMetrics) {
		t.Error("The expected number of metrics did not match : %d != %d", metricCount, len(nats.ConnectionMetrics))
	}
}

func sendReceiveMessages(t *testing.T, connManager nats.ConnManager) {
	conns := createConns(t, connManager, 5)
	const TOPIC = messaging.Topic("TestConnManager_Metrics")
	subscriber1, _ := conns[0].TopicSubscribe(TOPIC, nil)
	subscriber2, _ := conns[1].TopicSubscribe(TOPIC, nil)
	qsubscriber1, _ := conns[0].TopicQueueSubscribe(TOPIC, TOPIC.AsQueue(), nil)
	qsubscriber2, _ := conns[1].TopicQueueSubscribe(TOPIC, TOPIC.AsQueue(), nil)
	qsubscriber3, _ := conns[2].TopicQueueSubscribe(TOPIC, TOPIC.AsQueue(), nil)
	qsubscriber4, _ := conns[3].TopicQueueSubscribe(TOPIC, TOPIC.AsQueue(), nil)
	qsubscriber5, _ := conns[4].TopicQueueSubscribe(TOPIC, TOPIC.AsQueue(), nil)

	const MESSAGE_COUNT = 10
	publisher1, _ := conns[0].Publisher(TOPIC)
	publisher2, _ := conns[4].Publisher(TOPIC)
	for i := 0; i < MESSAGE_COUNT; i++ {
		publisher1.Publish([]byte(nuid.Next()))
		publisher2.Publish([]byte(nuid.Next()))
	}

	messageReceivedCounts := map[string]int{}
	const TOTAL_EXPECTED_MSG_RECEIVED_COUNT = /* topic subscribers */ (MESSAGE_COUNT * 2 * 2) + /* queue subscribers */ (MESSAGE_COUNT * 2)
	for i := 0; i < TOTAL_EXPECTED_MSG_RECEIVED_COUNT*5; i++ {
		select {
		case <-subscriber1.Channel():
			messageReceivedCounts["subscriber1"]++
		case <-subscriber2.Channel():
			messageReceivedCounts["subscriber2"]++
		case <-qsubscriber1.Channel():
			messageReceivedCounts["qsubscriber1"]++
		case <-qsubscriber2.Channel():
			messageReceivedCounts["qsubscriber2"]++
		case <-qsubscriber3.Channel():
			messageReceivedCounts["qsubscriber3"]++
		case <-qsubscriber4.Channel():
			messageReceivedCounts["qsubscriber4"]++
		case <-qsubscriber5.Channel():
			messageReceivedCounts["qsubscriber5"]++
		default:
			totalMsgReceivedCount := 0
			for _, count := range messageReceivedCounts {
				totalMsgReceivedCount += count
			}
			if totalMsgReceivedCount >= TOTAL_EXPECTED_MSG_RECEIVED_COUNT {
				break
			}
			time.Sleep(time.Millisecond)
		}
	}

	totalMsgReceivedCount := 0
	for _, count := range messageReceivedCounts {
		totalMsgReceivedCount += count
	}
	if totalMsgReceivedCount != TOTAL_EXPECTED_MSG_RECEIVED_COUNT {
		t.Errorf("totalMsgReceivedCount != TOTAL_EXPECTED_MSG_RECEIVED_COUNT : %d != %d", totalMsgReceivedCount, TOTAL_EXPECTED_MSG_RECEIVED_COUNT)
	}

	if connManager.TotalMsgsIn() != TOTAL_EXPECTED_MSG_RECEIVED_COUNT {
		t.Errorf("connManager.TotalMsgsIn() did not match : %d != %d", connManager.TotalMsgsIn(), TOTAL_EXPECTED_MSG_RECEIVED_COUNT)
	}
	if connManager.TotalMsgsOut() != MESSAGE_COUNT*2 {
		t.Errorf("connManager.TotalMsgsOut() did not match %d != %d", connManager.TotalMsgsOut(), MESSAGE_COUNT*2)
	}

	t.Logf("total bytes in : %d, total bytes out : %d", connManager.TotalBytesIn(), connManager.TotalBytesOut())
	if connManager.TotalBytesIn() != connManager.TotalBytesOut()*3 {
		t.Errorf("TotalBytesIn did not match what's expected : %d != %d", connManager.TotalBytesIn(), connManager.TotalBytesOut()*3)
	}

}

func TestNewConnManager_InvalidClusterName(t *testing.T) {
	f := func(cluster messaging.ClusterName) {
		serverConfig := natstest.CreateNATSServerConfigsNoTLS(1)[0]
		serverConfig.Cluster = cluster
		settings := natstest.ConnManagerSettings(serverConfig)
		defer func() {
			if p := recover(); p == nil {
				t.Errorf("NewConnManager should have panicked because ClusterName is invalid : [%s]", settings.ClusterName)
			} else {
				t.Log(p)
			}
		}()
		nats.NewConnManager(settings)
	}

	f(messaging.ClusterName(""))
	f(messaging.ClusterName("a"))
	f(messaging.ClusterName("A"))
	f(messaging.ClusterName("Asdfsdf"))
	f(messaging.ClusterName("1sdfsdf"))
}

func createConns(t *testing.T, connManager nats.ConnManager, count int) []*nats.ManagedConn {
	t.Helper()
	conns := []*nats.ManagedConn{}
	for i := 0; i < count; i++ {
		conn, err := connManager.Connect()
		if err != nil {
			t.Fatal(err)
		}
		conns = append(conns, conn)
	}
	return conns
}
