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

	"github.com/nats-io/nuid"
	"github.com/oysterpack/oysterpack.go/pkg/commons/collections/sets"
	"github.com/oysterpack/oysterpack.go/pkg/messaging"
	"github.com/oysterpack/oysterpack.go/pkg/messaging/nats"
	"github.com/oysterpack/oysterpack.go/pkg/messaging/natstest"
	"github.com/oysterpack/oysterpack.go/pkg/metrics"
)

func TestNewConn(t *testing.T) {
	metrics.ResetRegistry()
	defer metrics.ResetRegistry()

	serverConfigs := natstest.CreateNATSServerConfigsNoTLS(1)
	servers := natstest.CreateNATSServers(t, serverConfigs)
	natstest.StartServers(servers)
	defer natstest.ShutdownServers(servers)

	connManager := nats.NewConnManager(natstest.ConnManagerSettings(serverConfigs[0]))
	client := nats.NewClient(connManager)
	defer client.CloseAllConns()

	if client.Cluster() != servers[0].Cluster() {
		t.Errorf("client.Cluster() does not match : %v != %v", client.Cluster(), servers[0].Cluster())
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
	if int(conn.MaxPayload()) != serverConfigs[0].MaxPayload {
		t.Errorf("MaxPayload did not match the server config : %d != %d", conn.MaxPayload(), serverConfigs[0].MaxPayload)
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
	publisher, _ := conn.Publisher(topic)

	publisher.Publish([]byte("CIAO MUNDO #1"))
	msg := <-subscriber.Channel()
	t.Logf("Received msg : %v", string(msg.Data))
	if msg.Topic != topic {
		t.Errorf("topic did not match : %v", msg.Topic)
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
