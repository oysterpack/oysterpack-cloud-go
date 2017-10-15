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

	"sync"

	"context"

	"fmt"

	"github.com/nats-io/nuid"
	"github.com/oysterpack/oysterpack.go/pkg/commons/collections/sets"
	"github.com/oysterpack/oysterpack.go/pkg/messaging"
	"github.com/oysterpack/oysterpack.go/pkg/messaging/nats"
	"github.com/oysterpack/oysterpack.go/pkg/messaging/nats/server"
	"github.com/oysterpack/oysterpack.go/pkg/messaging/natstest"
	"github.com/oysterpack/oysterpack.go/pkg/metrics"
	"github.com/rs/zerolog/log"
)

func TestNewConn(t *testing.T) {
	runTestSingleServer(t, func(server server.NATSServer, serverConfig *server.NATSServerConfig, client messaging.Client, connManager nats.ConnManager) {

		if client.Cluster() != server.Cluster() {
			t.Errorf("client.Cluster() does not match : %v != %v", client.Cluster(), server.Cluster())
		}

		conn, err := client.Connect()
		if err != nil {
			t.Fatal(err)
		}
		if conn.Status() != messaging.CONNECTED || !conn.Connected() || conn.Closed() || conn.Reconnecting() {
			t.Errorf("connection is reporting incorrect status")
		}
		if conn.Cluster() != connManager.Cluster() {
			t.Errorf("cluster does not match : %v != %v", conn.Cluster(), connManager.Cluster())
		}
		if int(conn.MaxPayload()) != serverConfig.MaxPayload {
			t.Errorf("MaxPayload did not match the server config : %d != %d", conn.MaxPayload(), serverConfig.MaxPayload)
		}
		if conn.LastError() != nil {
			t.Error(conn.LastError())
		}

		checkTags(t, client)
		checkUniqueConnIds(t, client)

		testPublishSubscribe(t, conn)
		testPublishQueueSubscribe(t, conn)
		testPublishSubscribe_WithEmptySubscriptionSettings(t, conn)
		testPublishSubscribe_WithSubscriptionSettings_WithValidLimits(t, conn)
		testPublishSubscribe_WithSubscriptionSettings_WithInvalidLimits(t, conn)
		testPublishSubscribe_WithSubscriptionSettings_Unlimited(t, conn)

		testAutoUnsubscribe(t, conn)

		testRequest(t, conn)
		testAsyncRequest(t, conn)
		testRequestChannel(t, conn)
		testRequestWithContext(t, conn)
		testRequestChannelWithContext(t, conn)
		testAsyncRequestChannelWithContext(t, conn)

		checkHealthChecks(t, client)
		checkMetrics(t, client)

		natstest.LogConnInfo(t, connManager)
		testConnInfoLookup(t, connManager)
		testManagedConnLookup(t, connManager)

		conn.Close()
		if conn.Status() != messaging.CLOSED || !conn.Closed() {
			t.Errorf("conn should be closed : %v", conn.Status())
		}
	})
}

func TestConn_Reconnect(t *testing.T) {
	metrics.ResetRegistry()
	defer metrics.ResetRegistry()

	serverConfigs := natstest.CreateNATSServerConfigsNoTLS(1)
	servers := natstest.CreateNATSServers(t, serverConfigs)
	natstest.StartServers(servers)
	defer natstest.ShutdownServers(servers)

	connManager := nats.NewConnManager(natstest.ConnManagerSettings(serverConfigs[0]))
	client := nats.NewClient(connManager)
	defer client.CloseAllConns()

	conn, err := client.Connect("reconnect")
	if err != nil {
		t.Fatal(err)
	}

	natstest.ShutdownServers(servers)
	for i := 0; i < 10; i++ {
		if conn.Connected() {
			time.Sleep(5 * time.Millisecond)
		}
	}
	if conn.Connected() {
		t.Error("connection should be disconnected")
	}

	servers = natstest.CreateNATSServers(t, serverConfigs)
	natstest.StartServers(servers)
	defer natstest.ShutdownServers(servers)
	time.Sleep(natstest.ReConnectTimeout + (5 * time.Millisecond))
	if !conn.Connected() {
		t.Error("connection should be disconnected")
	}
}

func TestConn_AutoReconnectInCluster(t *testing.T) {
	runTestWithMultipleServer(t, 2, func(servers []server.NATSServer, serverConfigs []*server.NATSServerConfig) {
		connManager := nats.NewConnManager(natstest.ConnManagerSettings(serverConfigs[0]))
		client := nats.NewClient(connManager)
		defer client.CloseAllConns()

		conn, err := client.Connect("auto-reconnect")
		if err != nil {
			t.Fatal(err)
		}

		servers[0].Shutdown()
		time.Sleep(natstest.ReConnectTimeout + (5 * time.Millisecond))
		if !conn.Connected() {
			t.Error("connection should be disconnected")
		}
	})
}

func TestPublisherCaching(t *testing.T) {
	runTestSingleServer(t, func(server server.NATSServer, config *server.NATSServerConfig, client messaging.Client, connManager nats.ConnManager) {
		conn, err := client.Connect()
		if err != nil {
			t.Fatal(err)
		}

		topic := messaging.Topic(nuid.Next())
		publisher, err := conn.Publisher(topic)
		if err != nil {
			t.Fatal(err)
		}

		_, err = conn.Publisher(topic)
		if err != nil {
			t.Fatal(err)
		}

		publisher2, err := conn.Publisher(messaging.Topic(nuid.Next()))
		if err != nil {
			t.Fatal(err)
		}

		publisherCounts := connManager.PublisherCountsPerTopic()
		if publisherCounts[publisher.Topic()] != 1 {
			t.Errorf("There should be 1 publisher per topic per connection")
		}

		if publisherCounts[publisher2.Topic()] != 1 {
			t.Errorf("There should be 1 publisher per topic per connection")
		}

		conn.Close()
		if _, err := conn.Publisher(messaging.Topic(nuid.Next())); err == nil {
			t.Error("Creating publisher should fail when the connection is closed")
		}
	})

}

func TestUnsubscribe(t *testing.T) {
	runTestSingleServer(t, func(server server.NATSServer, config *server.NATSServerConfig, client messaging.Client, connManager nats.ConnManager) {
		conn, err := client.Connect()
		if err != nil {
			t.Fatal(err)
		}

		subscription, err := conn.Subscribe(messaging.Topic(nuid.Next()), nil)
		if err != nil {
			t.Fatal(err)
		}

		subscription.Unsubscribe()
		conn.Publish(subscription.Topic(), []byte("data"))
		if subscription.IsValid() {
			t.Error("subscription should been cancelled")
		}
		if msg := <-subscription.Channel(); msg != nil {
			t.Error("No message should have been reeived")
		}

		qsubscription, err := conn.QueueSubscribe(messaging.Topic(nuid.Next()), messaging.Topic(nuid.Next()).AsQueue(), nil)
		if err != nil {
			t.Fatal(err)
		}

		qsubscription.Unsubscribe()
		conn.Publish(qsubscription.Topic(), []byte("data"))
		if qsubscription.IsValid() {
			t.Error("subscription should been cancelled")
		}
		if msg := <-qsubscription.Channel(); msg != nil {
			t.Error("No message should have been reeived")
		}
	})
}

func TestSlowSubscriber(t *testing.T) {
	runTestSingleServer(t, func(server server.NATSServer, serverConfig *server.NATSServerConfig, client messaging.Client, connManager nats.ConnManager) {
		conn, _ := client.Connect("slow-consumer")
		topic := messaging.Topic(nuid.Next())
		subscriber, _ := conn.Subscribe(topic, &messaging.SubscriptionSettings{&messaging.PendingLimits{MsgLimit: 1, BytesLimit: 64}})

		msgLimit, byteLimit, err := subscriber.PendingLimits()
		if err != nil {
			t.Error(err)
		} else {
			if msgLimit != 1 {
				t.Errorf("Msg limit does not match : %d", msgLimit)
			}
			if byteLimit != 64 {
				t.Errorf("Byte limit does not match : %d", byteLimit)
			}
		}

		go func() {
			for {
				if !subscriber.IsValid() {
					return
				}
				msg := <-subscriber.Channel()
				log.Logger.Info().Msgf("TestSlowSubscriber() : received msg : %s", string(msg.Data))
				time.Sleep(50 * time.Millisecond)
			}

		}()

		pubConn, _ := client.Connect()
		publisher, _ := pubConn.Publisher(topic)
		for i := 0; i < 100; i++ {
			msg := fmt.Sprintf("MSG #%d", i)
			publisher.Publish([]byte(msg))
			log.Logger.Info().Msgf("TestSlowSubscriber() published msg : %s", msg)
		}
		time.Sleep(10 * time.Millisecond)

		msgsMaxPending, bytesMaxPending, err := subscriber.MaxPending()
		if err != nil {
			t.Error(err)
		} else {
			t.Logf("msgsMaxPending = %d, bytesMaxPending = %d", msgsMaxPending, bytesMaxPending)
			if msgsMaxPending == 0 || bytesMaxPending == 0 {
				t.Error("There should have been pending messages")
			}
		}
		subscriber.ClearMaxPending()
		msgsMaxPending, bytesMaxPending, err = subscriber.MaxPending()
		if err != nil {
			t.Error(err)
		} else {
			t.Logf("msgsMaxPending = %d, bytesMaxPending = %d", msgsMaxPending, bytesMaxPending)
		}

		info, err := subscriber.SubscriptionInfo()
		if err != nil {
			t.Error(err)
		} else {
			t.Logf("subscription info : %v", info)
		}

		subscriber.Unsubscribe()
	})
}

func runTestWithMultipleServer(t *testing.T, serverCount int, test func(servers []server.NATSServer, serverConfigs []*server.NATSServerConfig)) {
	metrics.ResetRegistry()
	defer metrics.ResetRegistry()

	serverConfigs := natstest.CreateNATSServerConfigsNoTLS(serverCount)
	servers := natstest.CreateNATSServers(t, serverConfigs)
	natstest.StartServers(servers)
	defer natstest.ShutdownServers(servers)

	test(servers, serverConfigs)
}

func runTestSingleServer(t *testing.T, test func(server server.NATSServer, serverConfig *server.NATSServerConfig, client messaging.Client, connManager nats.ConnManager)) {
	metrics.ResetRegistry()
	defer metrics.ResetRegistry()

	serverConfigs := natstest.CreateNATSServerConfigsNoTLS(1)
	servers := natstest.CreateNATSServers(t, serverConfigs)
	natstest.StartServers(servers)
	defer natstest.ShutdownServers(servers)

	connManager := nats.NewConnManager(natstest.ConnManagerSettings(serverConfigs[0]))
	client := nats.NewClient(connManager)
	defer client.CloseAllConns()

	test(servers[0], serverConfigs[0], client, connManager)
}

func testConnInfoLookup(t *testing.T, connManager nats.ConnManager) {
	for _, info := range connManager.ConnInfos() {
		if connInfo := connManager.ConnInfo(info.Id); connInfo == nil {
			t.Errorf("Lookup for ConnInfo(%v) failed", connInfo.Id)
		} else if info.Id != connInfo.Id {
			t.Errorf("Wrong ConnInfo(%v) was returned : %v", info.Id, connInfo.Id)
		}
	}

	if connInfo := connManager.ConnInfo(nuid.Next()); connInfo != nil {
		t.Error("Lookup should have found nothing for random id")
	}
}

func testManagedConnLookup(t *testing.T, connManager nats.ConnManager) {
	for _, info := range connManager.ConnInfos() {
		if conn := connManager.ManagedConn(info.Id); conn == nil {
			t.Errorf("Lookup for ManagedConn(%v) failed", conn.ID())
		} else {
			t.Logf("conn : %s", conn)
		}
	}
}

func checkMetrics(t *testing.T, client messaging.Client) {
	metrics := client.Metrics()
	if metrics == nil {
		t.Error("no metrics were returned")
	}
	if len(metrics.GaugeVecOpts) != len(nats.ConnManagerMetrics.GaugeVecOpts) {
		t.Errorf("gauge vectors count does not match : %d != %d", len(metrics.GaugeOpts), len(nats.ConnManagerMetrics.GaugeOpts))
	}
	if len(metrics.CounterVecOpts) != len(nats.ConnManagerMetrics.CounterVecOpts) {
		t.Errorf("counter vectors count does not match : %d != %d", len(metrics.GaugeOpts), len(nats.ConnManagerMetrics.GaugeOpts))
	}
}

func checkHealthChecks(t *testing.T, client messaging.Client) {
	healthchecks := client.HealthChecks()
	if len(healthchecks) == 0 {
		t.Error("there are 0 healthchecks")
	}

	for _, healthcheck := range healthchecks {
		result := healthcheck.Run()
		if !result.Success() {
			t.Error(result.Err)
		}
	}
}

func testAsyncRequestChannelWithContext(t *testing.T, conn messaging.Conn) {
	topic := messaging.Topic(nuid.Next())

	subscriber, err := conn.Subscribe(topic, &messaging.SubscriptionSettings{})
	if err != nil {
		t.Error(err)
		return
	}

	go func() {
		msg := <-subscriber.Channel()
		t.Logf("Received msg : %v", string(msg.Data))
		conn.Publish(msg.ReplyTo.AsTopic(), msg.Data)
	}()

	msg := "CIAO MUNDO #1"
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	responseWait := sync.WaitGroup{}
	responseWait.Add(1)
	conn.AsyncRequestWithContext(ctx, topic, []byte(msg), func(response messaging.Response) {
		defer responseWait.Done()
		if response.Error != nil {
			t.Error(response.Error)
		} else if string(response.Data) != msg {
			t.Error("Response message did not match")
		}
	})

	responseWait.Wait()
}

func testRequestChannelWithContext(t *testing.T, conn messaging.Conn) {
	topic := messaging.Topic(nuid.Next())

	subscriber, err := conn.Subscribe(topic, &messaging.SubscriptionSettings{})
	if err != nil {
		t.Error(err)
		return
	}

	go func() {
		msg := <-subscriber.Channel()
		t.Logf("Received msg : %v", string(msg.Data))
		conn.Publish(msg.ReplyTo.AsTopic(), msg.Data)
	}()

	msg := "CIAO MUNDO #1"
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	responseChannel := conn.RequestChannelWithContext(ctx, topic, []byte(msg))
	response := <-responseChannel
	if response.Error != nil {
		t.Error(response.Error)
	} else if string(response.Data) != msg {
		t.Error("Response message did not match")
	}
}

func testRequestWithContext(t *testing.T, conn messaging.Conn) {
	topic := messaging.Topic(nuid.Next())

	subscriber, err := conn.Subscribe(topic, &messaging.SubscriptionSettings{})
	if err != nil {
		t.Error(err)
		return
	}

	go func() {
		msg := <-subscriber.Channel()
		t.Logf("Received msg : %v", string(msg.Data))
		conn.Publish(msg.ReplyTo.AsTopic(), msg.Data)
	}()

	msg := "CIAO MUNDO #1"
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	response, err := conn.RequestWithContext(ctx, topic, []byte(msg))
	if err != nil {
		t.Error(err)
	}
	if string(response.Data) != msg {
		t.Error("Response message did not match")
	}
}

func testRequestChannel(t *testing.T, conn messaging.Conn) {
	topic := messaging.Topic(nuid.Next())

	subscriber, err := conn.Subscribe(topic, &messaging.SubscriptionSettings{})
	if err != nil {
		t.Error(err)
		return
	}

	go func() {
		msg := <-subscriber.Channel()
		t.Logf("Received msg : %v", string(msg.Data))
		conn.Publish(msg.ReplyTo.AsTopic(), msg.Data)
	}()

	msg := "CIAO MUNDO #1"
	responseChannel := conn.RequestChannel(topic, []byte(msg), time.Second)
	response := <-responseChannel
	if response.Error != nil {
		t.Error(response.Error)
	} else if string(response.Data) != msg {
		t.Error("Response message did not match")
	}
}

func testAsyncRequest(t *testing.T, conn messaging.Conn) {
	topic := messaging.Topic(nuid.Next())

	subscriber, err := conn.Subscribe(topic, &messaging.SubscriptionSettings{})
	if err != nil {
		t.Error(err)
		return
	}

	go func() {
		msg := <-subscriber.Channel()
		t.Logf("Received msg : %v", string(msg.Data))
		conn.Publish(msg.ReplyTo.AsTopic(), msg.Data)
	}()

	responseWait := sync.WaitGroup{}
	responseWait.Add(1)
	msg := "CIAO MUNDO #1"
	conn.AsyncRequest(topic, []byte(msg), time.Second, func(response messaging.Response) {
		defer responseWait.Done()
		if response.Error != nil {
			t.Error(response.Error)
		} else if string(response.Data) != msg {
			t.Error("Response message did not match")
		}
	})

	responseWait.Wait()
}

func testRequest(t *testing.T, conn messaging.Conn) {
	topic := messaging.Topic(nuid.Next())

	subscriber, err := conn.Subscribe(topic, &messaging.SubscriptionSettings{})
	if err != nil {
		t.Error(err)
		return
	}

	go func() {
		msg := <-subscriber.Channel()
		t.Logf("Received msg : %v", string(msg.Data))
		conn.Publish(msg.ReplyTo.AsTopic(), msg.Data)
	}()

	msg := "CIAO MUNDO #1"
	response, err := conn.Request(topic, []byte(msg), time.Second)
	if err != nil {
		t.Error(err)
	}
	if string(response.Data) != msg {
		t.Error("Response message did not match")
	}
}

func testPublishSubscribe_WithEmptySubscriptionSettings(t *testing.T, conn messaging.Conn) {
	topic := messaging.Topic(nuid.Next())

	subscriber, err := conn.Subscribe(topic, &messaging.SubscriptionSettings{})
	if err != nil {
		t.Error(err)
		return
	}
	if subscriber.Cluster() != conn.Cluster() {
		t.Errorf("subscriber Cluster does not match : %s", subscriber.Cluster())
	}

	publisher, _ := conn.Publisher(topic)

	publisher.Publish([]byte("CIAO MUNDO #1"))
	msg := <-subscriber.Channel()
	t.Logf("Received msg : %v", string(msg.Data))
	if msg.Topic != topic {
		t.Errorf("topic did not match : %v", msg.Topic)
	}
}

func testAutoUnsubscribe(t *testing.T, conn messaging.Conn) {
	topic := messaging.Topic(nuid.Next())

	subscriber, err := conn.Subscribe(topic, &messaging.SubscriptionSettings{})
	if err != nil {
		t.Error(err)
		return
	}
	if subscriber.Cluster() != conn.Cluster() {
		t.Errorf("subscriber Cluster does not match : %s", subscriber.Cluster())
	}
	subscriber.AutoUnsubscribe(1)

	publisher, _ := conn.Publisher(topic)

	publisher.Publish([]byte("CIAO MUNDO #1"))
	msg := <-subscriber.Channel()
	t.Logf("Received msg : %v", string(msg.Data))
	if msg.Topic != topic {
		t.Errorf("topic did not match : %v", msg.Topic)
	}

	publisher.Publish([]byte("CIAO MUNDO #1"))
	msg = <-subscriber.Channel()
	if msg != nil {
		t.Error("subscriber should have auto-unsubsctibed after receiving 1 message")
	}
}

func testPublishSubscribe_WithSubscriptionSettings_WithValidLimits(t *testing.T, conn messaging.Conn) {
	topic := messaging.Topic(nuid.Next())

	subscriber, err := conn.Subscribe(topic, &messaging.SubscriptionSettings{PendingLimits: &messaging.PendingLimits{BytesLimit: 1024, MsgLimit: 100}})
	if err != nil {
		t.Error(err)
		return
	}
	publisher, _ := conn.Publisher(topic)

	publisher.Publish([]byte("CIAO MUNDO #1"))
	msg := <-subscriber.Channel()
	t.Logf("Received msg : %v", string(msg.Data))
	if msg.Topic != topic {
		t.Errorf("topic did not match : %v", msg.Topic)
	}
}

func testPublishSubscribe_WithSubscriptionSettings_WithInvalidLimits(t *testing.T, conn messaging.Conn) {
	topic := messaging.Topic(nuid.Next())

	_, err := conn.Subscribe(topic, &messaging.SubscriptionSettings{PendingLimits: &messaging.PendingLimits{BytesLimit: 0, MsgLimit: 100}})
	if err == nil {
		t.Error("An error was expected because 0 is not allowed for BytesLimit")
	}

	_, err = conn.Subscribe(topic, &messaging.SubscriptionSettings{PendingLimits: &messaging.PendingLimits{BytesLimit: 1024, MsgLimit: 0}})
	if err == nil {
		t.Error("An error was expected because 0 is not allowed for MsgLimit")
	}
}

func testPublishSubscribe_WithSubscriptionSettings_Unlimited(t *testing.T, conn messaging.Conn) {
	topic := messaging.Topic(nuid.Next())

	subscriber, err := conn.Subscribe(topic, &messaging.SubscriptionSettings{PendingLimits: &messaging.PendingLimits{BytesLimit: -1, MsgLimit: -1}})
	if err != nil {
		t.Error(err)
		return
	}
	publisher, _ := conn.Publisher(topic)

	publisher.Publish([]byte("CIAO MUNDO #1"))
	msg := <-subscriber.Channel()
	t.Logf("Received msg : %v", string(msg.Data))
	if msg.Topic != topic {
		t.Errorf("topic did not match : %v", msg.Topic)
	}
}

func testPublishQueueSubscribe(t *testing.T, conn messaging.Conn) {
	topic := messaging.Topic(nuid.Next())
	subscriptions := make([]messaging.QueueSubscription, 2)
	for i := 0; i < len(subscriptions); i++ {
		subscriber, err := conn.QueueSubscribe(topic, topic.AsQueue(), nil)
		if err != nil {
			t.Fatal(err)
		}
		subscriptions[i] = subscriber
	}

	conn.Publish(topic, []byte("CIAO MUNDO #1"))

LOOP:
	for {
		for _, sub := range subscriptions {
			select {
			case msg := <-sub.Channel():
				t.Logf("Received msg : %v", string(msg.Data))
				break LOOP
			default:
				time.Sleep(time.Millisecond)
			}
		}
	}

	for _, sub := range subscriptions {
		select {
		case msg := <-sub.Channel():
			t.Logf("Received msg : %v", string(msg.Data))
		default:
			time.Sleep(time.Millisecond)
		}
	}

	msgReceivedCount := 0
	for _, sub := range subscriptions {
		count, _ := sub.Delivered()
		msgReceivedCount += int(count)
		t.Log(sub.QueueSubscriptionInfo())
	}
	if msgReceivedCount != 1 {
		t.Error("Only 1 of the subscribers should have received the message")
	}

}

func testPublishSubscribe(t *testing.T, conn messaging.Conn) {
	topic := messaging.Topic(nuid.Next())
	subscriber, err := conn.Subscribe(topic, nil)
	if err != nil {
		t.Error(err)
		return
	}
	publisher, _ := conn.Publisher(topic)

	publisher.Publish([]byte("CIAO MUNDO #1"))
	msg := <-subscriber.Channel()
	t.Logf("Received msg : %v", string(msg.Data))
	if msg.Topic != topic {
		t.Errorf("topic did not match : %v", msg.Topic)
	}

	conn.Publish(topic, []byte("CIAO MUNDO #2"))
	msg = <-subscriber.Channel()
	t.Logf("Received msg : %v", string(msg.Data))
	if msg.Topic != topic {
		t.Errorf("topic did not match : %v", msg.Topic)
	}

	replyTo := messaging.ReplyTo(nuid.Next())
	conn.PublishRequest(topic, replyTo, []byte("CIAO MUNDO #3"))

	msg = <-subscriber.Channel()
	t.Logf("Received msg : %v", string(msg.Data))
	if msg.Topic != topic {
		t.Errorf("topic did not match : %v", msg.Topic)
	}
	if msg.ReplyTo != replyTo {
		t.Errorf("replyTo did not match : %v != %v ", msg.ReplyTo, replyTo)
	}
}

func checkUniqueConnIds(t *testing.T, client messaging.Client) {
	ids := sets.NewStrings()
	for i := 0; i < 10; i++ {
		conn, err := client.Connect()
		if err != nil {
			t.Fatal(err)
		}
		ids.Add(conn.ID())
		conn.Close()
	}
	if ids.Size() != 10 {
		t.Errorf("Ids were not unique")
	}
}

func checkTags(t *testing.T, client messaging.Client) {
	tags := []string{"a", "b", "c"}
	conn, err := client.Connect(tags...)
	if err != nil {
		t.Fatal(err)
	}

	if len(conn.Tags()) != 3 {
		t.Errorf("tags were not set")
	} else {
		tagSet := sets.NewStrings()
		tagSet.AddAll(tags...)

		connTags := sets.NewStrings()
		connTags.AddAll(conn.Tags()...)

		if !connTags.Equals(tagSet) {
			t.Errorf("tags do not match : %v != %v", tags, conn.Tags())
		}
	}

	if client.ConnCount() < client.ConnCount("a") {
		t.Errorf("client.ConnCount() should return the total number of live connections : %d", client.ConnCount())
	}

	if count := client.ConnCount("a"); count != 1 {
		t.Errorf("count did not match %d != 1", count)
	}
	if count := client.ConnCount("a", "b"); count != 1 {
		t.Errorf("count did not match %d != 1", count)
	}
	if count := client.ConnCount(tags...); count != 1 {
		t.Errorf("count did not match %d != 1", count)
	}

	client.Connect(tags...)
	if count := client.ConnCount("a"); count != 2 {
		t.Errorf("count did not match %d != 1", count)
	}
	if count := client.ConnCount("a", "b"); count != 2 {
		t.Errorf("count did not match %d != 1", count)
	}
	if count := client.ConnCount(tags...); count != 2 {
		t.Errorf("count did not match %d != 1", count)
	}

	client.Connect("a", "y", "z")
	if count := client.ConnCount("a"); count != 3 {
		t.Errorf("count did not match %d != 1", count)
	}
	if count := client.ConnCount("a", "b"); count != 2 {
		t.Errorf("count did not match %d != 1", count)
	}
	if count := client.ConnCount(tags...); count != 2 {
		t.Errorf("count did not match %d != 1", count)
	}
	if client.ConnCount() < client.ConnCount("a") {
		t.Errorf("client.ConnCount() should return the total number of live connections : %d", client.ConnCount())
	}

}
