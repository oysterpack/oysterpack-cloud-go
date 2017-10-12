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

package server_test

import (
	"fmt"
	"testing"

	"io/ioutil"
	"net/http"

	"github.com/nats-io/go-nats"
	"github.com/oysterpack/oysterpack.go/pkg/messaging"
	opnats "github.com/oysterpack/oysterpack.go/pkg/messaging/nats"
	"github.com/oysterpack/oysterpack.go/pkg/messaging/nats/server"
	"github.com/rs/zerolog/log"
)

/*
Test Desciption:

- create a 3 node cluster
  - the first node is used to seed the cluster
  - TLS and client certs are required between server and client and between servers for routing
    - in order for the cluster TLS to work properly with verification, the TLS certs need to be created with IP SANS properly

		easypki create --ca-name oysterpack --dns "*" --ip 127.0.0.1 cluster.nats.dev.oysterpack.com

- create a connection to each cluster
- create a subscriber on each of connections
- create a queue subscriber on each connection
- publish a message from each of the connections

EXPECTED RESULT :
- each subscriber receives 3 messages for a total of 9 messages
- the queue subscriber pool should receive 3 messages

*/

func TestNATSServer_Cluster_TLS(t *testing.T) {
	const CONN_COUNT = 3
	configs := createNATSServerConfigs(CONN_COUNT)
	servers := createNATSServers(t, configs)
	startServers(servers)
	defer shutdownServers(servers)
	logNATServerInfo(servers, "Servers have been started")
	if err := waitforClusterMeshToForm(servers); err != nil {
		t.Fatalf("Full mesh did not form : %v", err)
	}

	// connect to each server
	// create a subscription on each server
	// create a queue subscription on each server
	const TOPIC = "TestNewNATSServer"
	const QUEUE = "TestNewNATSServer"
	var subscriptions []messaging.Subscription
	var qsubscriptions []messaging.QueueSubscription
	var conns []*opnats.ManagedConn
	for _, config := range configs {
		conn, err := nats.Connect(fmt.Sprintf("tls://localhost:%d", config.ServerPort), nats.Secure(clientTLSConfig()))
		if err != nil {
			t.Fatalf("Failed to connect : %v", err)
		}

		managedConn := opnats.NewManagedConn(config.Cluster, fmt.Sprintf("%s:%d", config.Cluster, config.ServerPort), conn, nil)
		conns = append(conns, managedConn)
		sub, err := managedConn.TopicSubscribe(TOPIC, nil)
		if err != nil {
			t.Errorf("Failed to create subscription on : %v", config.ServerPort)
		}
		subscriptions = append(subscriptions, sub)

		qsub, err := managedConn.TopicQueueSubscribe(TOPIC, QUEUE, nil)
		if err != nil {
			t.Errorf("Failed to create queue subscription on : %v", config.ServerPort)
		}
		qsubscriptions = append(qsubscriptions, qsub)

		logNATServerInfo(servers, fmt.Sprintf("created subscription on %d", config.ServerPort))
	}
	logNATServerInfo(servers, "Connected to each server and created a subsription on each server")
	if err := checkClientConnectionCounts(servers, 1); err != nil {
		t.Fatalf("%v", err)
	}
	waitForClusterToBecomeAwareOfAllSubscriptions(servers, len(qsubscriptions)+len(qsubscriptions))

	// publish a message on each connection
	i := 0
	for _, conn := range conns {
		i++
		conn.Publish(TOPIC, []byte(fmt.Sprintf("MSG #%d", i)))
	}
	log.Logger.Info().Msg("Published messages")

	var (
		SUBSCRIBER_EXPECTED_MSG_COUNT  = (CONN_COUNT * len(subscriptions))
		QSUBSCRIBER_EXPECTED_MSG_COUNT = len(conns)
	)
	subscriberMsgCount := receiveMessagesOnSubscriptions(subscriptions, SUBSCRIBER_EXPECTED_MSG_COUNT)
	qsubscriberMsgCount := receiveMessagesOnQueueSubscriptions(qsubscriptions, QSUBSCRIBER_EXPECTED_MSG_COUNT)
	if subscriberMsgCount != SUBSCRIBER_EXPECTED_MSG_COUNT {
		t.Errorf("subscriberMsgCount != SUBSCRIBER_EXPECTED_MSG_COUNT : %d ! %d", subscriberMsgCount, SUBSCRIBER_EXPECTED_MSG_COUNT)
	}
	if qsubscriberMsgCount != QSUBSCRIBER_EXPECTED_MSG_COUNT {
		t.Errorf("qsubscriberMsgCount != QSUBSCRIBER_EXPECTED_MSG_COUNT : %d ! %d", qsubscriberMsgCount, QSUBSCRIBER_EXPECTED_MSG_COUNT)
	}
}

func TestNATSServer_Monitoring(t *testing.T) {
	const CONN_COUNT = 1
	configs := createNATSServerConfigs(CONN_COUNT)
	servers := createNATSServers(t, configs)
	startServers(servers)
	defer shutdownServers(servers)
	logNATServerInfo(servers, "Servers have been started")

	// connect to each server
	// create a subscription on each server
	// create a queue subscription on each server
	const TOPIC = "TestNewNATSServer"
	const QUEUE = "TestNewNATSServerQ"
	var subscriptions []messaging.Subscription
	var qsubscriptions []messaging.QueueSubscription
	var conns []*opnats.ManagedConn
	for _, config := range configs {
		conn, err := nats.Connect(fmt.Sprintf("tls://localhost:%d", config.ServerPort), nats.Secure(clientTLSConfig()))
		if err != nil {
			t.Fatalf("Failed to connect : %v", err)
		}

		managedConn := opnats.NewManagedConn(config.Cluster, fmt.Sprintf("%s:%d", config.Cluster, config.ServerPort), conn, nil)
		conns = append(conns, managedConn)
		sub, err := managedConn.TopicSubscribe(TOPIC, nil)
		if err != nil {
			t.Errorf("Failed to create subscription on : %v", config.ServerPort)
		}
		subscriptions = append(subscriptions, sub)

		qsub, err := managedConn.TopicQueueSubscribe(TOPIC, QUEUE, nil)
		if err != nil {
			t.Errorf("Failed to create queue subscription on : %v", config.ServerPort)
		}
		qsubscriptions = append(qsubscriptions, qsub)

		logNATServerInfo(servers, fmt.Sprintf("created subscription on %d", config.ServerPort))
	}

	for _, server := range servers {
		baseMonitoringURL := fmt.Sprintf("http://%v", server.MonitorAddr())

		endpoints := []string{"/varz", "/connz", "/routez", "/subsz"}
		for _, endpoint := range endpoints {
			resp, err := http.Get(fmt.Sprintf("%s/%s", baseMonitoringURL, endpoint))
			if err != nil {
				t.Errorf("HTTP GET failed for endpoint: %v", endpoint, err)
			}
			defer resp.Body.Close()
			body, err := ioutil.ReadAll(resp.Body)
			t.Logf("%s response\n%s", endpoint, string(body))
		}

		resp, err := http.Get(fmt.Sprintf("http://localhost:%d/metrics", server.PrometheusHTTPExportPort()))
		if err != nil {
			t.Errorf("HTTP GET failed for prometheus metric export endpoint : %v", err)
		}
		defer resp.Body.Close()
		body, err := ioutil.ReadAll(resp.Body)
		t.Logf("prometheus metrics\n%s", string(body))
	}
}

func createNATSServerConfigs(count int) []*server.NATSServerConfig {
	configs := []*server.NATSServerConfig{}
	for i := 0; i < count; i++ {
		config := &server.NATSServerConfig{
			Cluster:             messaging.ClusterName("osyterpack-test"),
			ServerPort:          server.DEFAULT_SERVER_PORT + i,
			MonitorPort:         server.DEFAULT_MONITOR_PORT + i,
			ClusterPort:         server.DEFAULT_CLUSTER_PORT + i,
			MetricsExporterPort: server.DEFAULT_PROMETHEUS_EXPORTER_HTTP_PORT + i,

			Routes: defaultRoutesWithSeed(),
			// full mesh needs to be defined, but this contradicts what's documented
			//Routes:defaultRoutesWithSeed(server.DEFAULT_CLUSTER_PORT + 1, server.DEFAULT_CLUSTER_PORT + 2),

			TLSConfig:        serverTLSConfig(),
			ClusterTLSConfig: clusterTLSConfig(),
			LogLevel:         server.DEBUG,
		}
		configs = append(configs, config)
	}
	return configs
}
