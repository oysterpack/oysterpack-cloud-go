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

	"time"

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

func skipTestCluster_UsingSeedNodeInRoutes_External(t *testing.T) {
	const CONN_COUNT = 3
	configs := createNATSServerConfigs(CONN_COUNT)

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
	}

	i := 0
	for _, conn := range conns {
		conn.Flush()
		i++
		conn.Publish(TOPIC, []byte(fmt.Sprintf("MSG #%d", i)))
	}

	var EXPECTED_MSG_COUNT = (CONN_COUNT * len(subscriptions)) + len(qsubscriptions)
	msgReceivedCount := 0
	for i := 1; i <= EXPECTED_MSG_COUNT*2; i++ {
		select {
		case msg := <-subscriptions[0].Channel():
			msgReceivedCount++
			log.Logger.Info().Msgf("#%d : %v", i, string(msg.Data))
		case msg := <-subscriptions[1].Channel():
			msgReceivedCount++
			log.Logger.Info().Msgf("#%d : %v", i, string(msg.Data))
		case msg := <-subscriptions[2].Channel():
			msgReceivedCount++
			log.Logger.Info().Msgf("#%d : %v", i, string(msg.Data))
		case msg := <-qsubscriptions[0].Channel():
			msgReceivedCount++
			log.Logger.Info().Msgf("#%d : %v", i, string(msg.Data))
		case msg := <-qsubscriptions[1].Channel():
			msgReceivedCount++
			log.Logger.Info().Msgf("#%d : %v", i, string(msg.Data))
		case msg := <-qsubscriptions[2].Channel():
			msgReceivedCount++
			log.Logger.Info().Msgf("#%d : %v", i, string(msg.Data))
		default:
			if msgReceivedCount == EXPECTED_MSG_COUNT {
				break
			}
			log.Logger.Info().Msg("*** NO MESSAGE ***")
			time.Sleep(time.Millisecond)
		}
	}

	if msgReceivedCount != EXPECTED_MSG_COUNT {
		t.Errorf("The number of msgs received did not match what was expected : %d ! %d", msgReceivedCount, EXPECTED_MSG_COUNT)
	}

}

func TestCluster_UsingSingleSeedNode_DistributedTopicAndQueue(t *testing.T) {
	const CONN_COUNT = 3
	configs := createNATSServerConfigs(CONN_COUNT)
	servers := createNATSServers(t, configs)
	startServers(servers)
	logNATServerInfo(servers, "Servers have been started")

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

	defer shutdownServers(servers)
}

func receiveMessagesOnSubscriptions(subscriptions []messaging.Subscription, messageCount int) int {
	msgReceivedCount := 0
	for i, subscription := range subscriptions {
		msg := <-subscription.Channel()
		msgReceivedCount++
		log.Logger.Info().Msgf("subscriptions[%d] #%d : %v", i, msgReceivedCount, string(msg.Data))
	}

	for i := 0; i < 10; i++ {
		for i, subscription := range subscriptions {
			select {
			case msg := <-subscription.Channel():
				msgReceivedCount++
				log.Logger.Info().Msgf("subscriptions[%d] #%d : %v", i, msgReceivedCount, string(msg.Data))
			default:
			}
		}

		if msgReceivedCount >= messageCount {
			return msgReceivedCount
		}
		log.Logger.Info().Msgf("receiveMessagesOnQueueSubscriptions: %d / %d", msgReceivedCount, messageCount)
		time.Sleep(time.Millisecond)
	}
	return msgReceivedCount
}

func receiveMessagesOnQueueSubscriptions(qsubscriptions []messaging.QueueSubscription, messageCount int) int {
	msgReceivedCount := 0
	for i := 0; i < 10; i++ {
		for i, subscription := range qsubscriptions {
			select {
			case msg := <-subscription.Channel():
				msgReceivedCount++
				log.Logger.Info().Msgf("qsubscriptions[%d] #%d : %v", i, msgReceivedCount, string(msg.Data))
			default:
			}
		}

		if msgReceivedCount >= messageCount {
			return msgReceivedCount
		}
		log.Logger.Info().Msgf("receiveMessagesOnQueueSubscriptions: %d / %d", msgReceivedCount, messageCount)
		time.Sleep(time.Millisecond)
	}
	return msgReceivedCount
}

func waitForClusterToBecomeAwareOfAllSubscriptions(servers []server.NATSServer, subscriptionCount int) {
	for {
		for _, server := range servers {
			if int(server.NumSubscriptions()) != subscriptionCount {
				log.Logger.Info().Msgf("Subscription count = %d", server.NumSubscriptions())
				time.Sleep(time.Millisecond)
				continue
			}
		}
		log.Logger.Info().Msg("Entire cluster is aware of all subscriptions")
		break
	}
}

func createNATSServerConfigs(count int) []*server.NATSServerConfig {
	configs := []*server.NATSServerConfig{}
	for i := 0; i < count; i++ {
		config := &server.NATSServerConfig{
			Cluster:     messaging.ClusterName("osyterpack-test"),
			ServerPort:  server.DEFAULT_SERVER_PORT + i,
			MonitorPort: server.DEFAULT_MONITOR_PORT + i,
			ClusterPort: server.DEFAULT_CLUSTER_PORT + i,

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

func createNATSServers(t *testing.T, configs []*server.NATSServerConfig) []server.NATSServer {
	var servers []server.NATSServer
	for _, config := range configs {
		server, err := server.NewNATSServer(config)
		if err != nil {
			t.Fatalf("server.NewNATSServer failed : %v", err)
		}
		servers = append(servers, server)
	}
	return servers
}
