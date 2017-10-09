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

	"github.com/nats-io/go-nats"
	"github.com/oysterpack/oysterpack.go/pkg/messaging"
	opnats "github.com/oysterpack/oysterpack.go/pkg/messaging/nats"
	"github.com/oysterpack/oysterpack.go/pkg/messaging/nats/server"
	"github.com/rs/zerolog/log"
	"time"
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


WHEN explicitly configuring the full mesh in the routes
ACTUAL RESULT: PASS : 9 messages are received - the full mesh is formed

SUMMARY: This is not working as expected according to (http://nats.io/documentation/server/gnatsd-cluster)

NEXT STEPS IS TO SEE IF THIS WORKS WITH SERVER
*/


func skipTestCluster_UsingSeedNodeInRoutes_External(t *testing.T) {
	configs := []*server.NATSServerConfig{}
	const CONN_COUNT = 3
	for i := 0; i < CONN_COUNT; i++ {
		configs = append(configs, &server.NATSServerConfig{
			Cluster:     messaging.ClusterName("osyterpack-test"),
			ServerPort:  server.DEFAULT_SERVER_PORT + i,
			MonitorPort: server.DEFAULT_MONITOR_PORT + i,
			ClusterPort: server.DEFAULT_CLUSTER_PORT + i,

			Routes: defaultRoutesWithSeed(),
			// full mesh needs to be defined, but this contradicts what's documented
			//Routes:defaultRoutesWithSeed(server.DEFAULT_CLUSTER_PORT + 1, server.DEFAULT_CLUSTER_PORT + 2),

			TLSConfig: serverTLSConfig(),
			LogLevel:  server.DEBUG,
		})
	}


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

		qsub, err := managedConn.TopicQueueSubscribe(TOPIC,QUEUE,nil)
		if err != nil {
			t.Errorf("Failed to create queue subscription on : %v", config.ServerPort)
		}
		qsubscriptions = append(qsubscriptions,qsub)
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
		t.Errorf("The number of msgs received did not match what was expected : %d ! %d", msgReceivedCount,EXPECTED_MSG_COUNT)
	}

}

func TestCluster_UsingSingleSeedNode_DistributedTopicAndQueue(t *testing.T) {
	configs := []*server.NATSServerConfig{}
	const CONN_COUNT = 3
	for i := 0; i < CONN_COUNT; i++ {
		config := &server.NATSServerConfig{
			Cluster:     messaging.ClusterName("osyterpack-test"),
			ServerPort:  server.DEFAULT_SERVER_PORT + i,
			MonitorPort: server.DEFAULT_MONITOR_PORT + i,
			ClusterPort: server.DEFAULT_CLUSTER_PORT + i,

			Routes: defaultRoutesWithSeed(),
			// full mesh needs to be defined, but this contradicts what's documented
			//Routes:defaultRoutesWithSeed(server.DEFAULT_CLUSTER_PORT + 1, server.DEFAULT_CLUSTER_PORT + 2),

			TLSConfig: serverTLSConfig(),
			ClusterTLSConfig: clusterTLSConfig(),
			LogLevel:  server.DEBUG,
		}
		t.Logf("%v",*config)
		configs = append(configs, config)
	}

	var servers []server.NATSServer
	for _, config := range configs {
		server, err := server.NewNATSServer(config)
		if err != nil {
			t.Fatalf("server.NewNATSServer failed : %v", err)
		}
		servers = append(servers, server)
	}

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

		qsub, err := managedConn.TopicQueueSubscribe(TOPIC,QUEUE,nil)
		if err != nil {
			t.Errorf("Failed to create queue subscription on : %v", config.ServerPort)
		}
		qsubscriptions = append(qsubscriptions,qsub)

		logNATServerInfo(servers, fmt.Sprintf("created subscription on %d", config.ServerPort))
	}

	logNATServerInfo(servers, "Connected to each server and created a subsription on eac server")

	i := 0
	for _, conn := range conns {
		//conn.Flush()
		i++
		conn.Publish(TOPIC, []byte(fmt.Sprintf("MSG #%d", i)))
	}

	logNATServerInfo(servers, "After publishing messages")

	var EXPECTED_MSG_COUNT = (CONN_COUNT * len(subscriptions)) + len(qsubscriptions)
	msgReceivedCount := 0
	for i := 1; i <= EXPECTED_MSG_COUNT*2; i++ {
		select {
		case msg := <-subscriptions[0].Channel():
			msgReceivedCount++
			log.Logger.Info().Msgf("subscriptions[0] #%d : %v", i, string(msg.Data))
		case msg := <-subscriptions[1].Channel():
			msgReceivedCount++
			log.Logger.Info().Msgf("subscriptions[1] #%d : %v", i, string(msg.Data))
		case msg := <-subscriptions[2].Channel():
			msgReceivedCount++
			log.Logger.Info().Msgf("subscriptions[2] #%d : %v", i, string(msg.Data))
		case msg := <-qsubscriptions[0].Channel():
			msgReceivedCount++
			log.Logger.Info().Msgf("qsubscriptions[0] #%d : %v", i, string(msg.Data))
		case msg := <-qsubscriptions[1].Channel():
			msgReceivedCount++
			log.Logger.Info().Msgf("qsubscriptions[1] #%d : %v", i, string(msg.Data))
		case msg := <-qsubscriptions[2].Channel():
			msgReceivedCount++
			log.Logger.Info().Msgf("qsubscriptions[2] #%d : %v", i, string(msg.Data))
		default:
			if msgReceivedCount == EXPECTED_MSG_COUNT {
				break
			}
			log.Logger.Info().Msg("*** NO MESSAGE ***")
			time.Sleep(time.Millisecond)
		}
	}

	if msgReceivedCount != EXPECTED_MSG_COUNT {
		t.Errorf("The number of msgs received did not match what was expected : %d ! %d", msgReceivedCount,EXPECTED_MSG_COUNT)
	}

	defer shutdownServers(servers)
}
