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

	"sync"

	"time"

	"context"

	"github.com/nats-io/nuid"
	"github.com/oysterpack/oysterpack.go/pkg/messaging"
	"github.com/oysterpack/oysterpack.go/pkg/messaging/nats"
	"github.com/oysterpack/oysterpack.go/pkg/messaging/natstest"
	"github.com/oysterpack/oysterpack.go/pkg/metrics"
)

func TestPublisher_PublishRequest(t *testing.T) {
	metrics.ResetRegistry()
	defer metrics.ResetRegistry()

	serverConfigs := natstest.CreateNATSServerConfigsNoTLS(1)
	servers := natstest.CreateNATSServers(t, serverConfigs)
	natstest.StartServers(servers)
	defer natstest.ShutdownServers(servers)

	connManager := nats.NewConnManager(natstest.ConnManagerSettings(serverConfigs[0]))
	client := nats.NewClient(connManager)
	defer client.CloseAllConns()

	conn, _ := client.Connect()

	topic := messaging.Topic(nuid.Next())

	subscriber, _ := conn.Subscribe(topic, nil)
	replyWait := sync.WaitGroup{}
	replyWait.Add(1)
	go func() {
		for subscriber.IsValid() {
			msg := <-subscriber.Channel()
			if msg != nil {
				conn.Publish(msg.ReplyTo.AsTopic(), msg.Data)
				replyWait.Done()
			}
		}
	}()

	replyTopic := messaging.ReplyTo(nuid.Next())

	replySubscriber, _ := conn.Subscribe(replyTopic.AsTopic(), nil)
	replyWait.Add(1)
	go func() {
		for replySubscriber.IsValid() {
			msg := <-replySubscriber.Channel()
			if msg != nil {
				conn.Publish(msg.ReplyTo.AsTopic(), msg.Data)
				replyWait.Done()
			}
		}
	}()

	publisher, _ := conn.Publisher(topic)
	if publisher.Cluster() != client.Cluster() {
		t.Errorf("publisher Cluster does not match the clients : %v", publisher.Cluster())
	}
	publisher.PublishRequest([]byte("HELLO"), replyTopic)

	replyWait.Wait()
	subscriber.Unsubscribe()
	replySubscriber.Unsubscribe()
}

func TestPublisher_Request(t *testing.T) {
	metrics.ResetRegistry()
	defer metrics.ResetRegistry()

	serverConfigs := natstest.CreateNATSServerConfigsNoTLS(1)
	servers := natstest.CreateNATSServers(t, serverConfigs)
	natstest.StartServers(servers)
	defer natstest.ShutdownServers(servers)

	connManager := nats.NewConnManager(natstest.ConnManagerSettings(serverConfigs[0]))
	client := nats.NewClient(connManager)
	defer client.CloseAllConns()

	conn, _ := client.Connect()

	topic := messaging.Topic(nuid.Next())

	subscriber, _ := conn.Subscribe(topic, nil)
	go func() {
		for subscriber.IsValid() {
			msg := <-subscriber.Channel()
			if msg != nil {
				conn.Publish(msg.ReplyTo.AsTopic(), msg.Data)
			}
		}
	}()

	publisher, _ := conn.Publisher(topic)
	response, err := publisher.Request([]byte("HELLO"), time.Millisecond*50)
	if err != nil {
		t.Error(err)
	} else {
		if string(response.Data) != "HELLO" {
			t.Errorf("received different response : %v", string(response.Data))
		}
	}
	subscriber.Unsubscribe()
}

func TestPublisher_RequestTimedOut(t *testing.T) {
	metrics.ResetRegistry()
	defer metrics.ResetRegistry()

	serverConfigs := natstest.CreateNATSServerConfigsNoTLS(1)
	servers := natstest.CreateNATSServers(t, serverConfigs)
	natstest.StartServers(servers)
	defer natstest.ShutdownServers(servers)

	connManager := nats.NewConnManager(natstest.ConnManagerSettings(serverConfigs[0]))
	client := nats.NewClient(connManager)
	defer client.CloseAllConns()

	conn, _ := client.Connect()

	topic := messaging.Topic(nuid.Next())

	subscriber, _ := conn.Subscribe(topic, nil)
	go func() {
		for subscriber.IsValid() {
			<-subscriber.Channel()
		}
	}()

	publisher, _ := conn.Publisher(topic)
	_, err := publisher.Request([]byte("HELLO"), time.Millisecond)
	if err == nil {
		t.Error("Request should have timed out")
	} else {
		t.Log(err)
	}
	subscriber.Unsubscribe()
}

func TestPublisher_RequestWithContext(t *testing.T) {
	metrics.ResetRegistry()
	defer metrics.ResetRegistry()

	serverConfigs := natstest.CreateNATSServerConfigsNoTLS(1)
	servers := natstest.CreateNATSServers(t, serverConfigs)
	natstest.StartServers(servers)
	defer natstest.ShutdownServers(servers)

	connManager := nats.NewConnManager(natstest.ConnManagerSettings(serverConfigs[0]))
	client := nats.NewClient(connManager)
	defer client.CloseAllConns()

	conn, _ := client.Connect()

	topic := messaging.Topic(nuid.Next())

	subscriber, _ := conn.Subscribe(topic, nil)
	go func() {
		for subscriber.IsValid() {
			msg := <-subscriber.Channel()
			if msg != nil {
				conn.Publish(msg.ReplyTo.AsTopic(), msg.Data)
			}
		}
	}()

	publisher, _ := conn.Publisher(topic)
	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*50)
	defer cancel()
	response, err := publisher.RequestWithContext(ctx, []byte("HELLO"))
	if err != nil {
		t.Error(err)
	} else {
		if string(response.Data) != "HELLO" {
			t.Errorf("received different response : %v", string(response.Data))
		}
	}
	subscriber.Unsubscribe()
}

func TestPublisher_RequestWithContext_Timesout(t *testing.T) {
	metrics.ResetRegistry()
	defer metrics.ResetRegistry()

	serverConfigs := natstest.CreateNATSServerConfigsNoTLS(1)
	servers := natstest.CreateNATSServers(t, serverConfigs)
	natstest.StartServers(servers)
	defer natstest.ShutdownServers(servers)

	connManager := nats.NewConnManager(natstest.ConnManagerSettings(serverConfigs[0]))
	client := nats.NewClient(connManager)
	defer client.CloseAllConns()

	conn, _ := client.Connect()

	topic := messaging.Topic(nuid.Next())

	subscriber, _ := conn.Subscribe(topic, nil)
	go func() {
		for subscriber.IsValid() {
			<-subscriber.Channel()
		}
	}()

	publisher, _ := conn.Publisher(topic)
	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond)
	defer cancel()
	_, err := publisher.RequestWithContext(ctx, []byte("HELLO"))
	if err == nil {
		t.Error("Request should have timed out")
	} else {
		t.Log(err)
	}
	subscriber.Unsubscribe()
}

func TestPublisher_AsyncRequest(t *testing.T) {
	metrics.ResetRegistry()
	defer metrics.ResetRegistry()

	serverConfigs := natstest.CreateNATSServerConfigsNoTLS(1)
	servers := natstest.CreateNATSServers(t, serverConfigs)
	natstest.StartServers(servers)
	defer natstest.ShutdownServers(servers)

	connManager := nats.NewConnManager(natstest.ConnManagerSettings(serverConfigs[0]))
	client := nats.NewClient(connManager)
	defer client.CloseAllConns()

	conn, _ := client.Connect()

	topic := messaging.Topic(nuid.Next())

	subscriber, _ := conn.Subscribe(topic, nil)
	go func() {
		for subscriber.IsValid() {
			msg := <-subscriber.Channel()
			if msg != nil {
				conn.Publish(msg.ReplyTo.AsTopic(), msg.Data)
			}
		}
	}()

	publisher, _ := conn.Publisher(topic)
	wait := sync.WaitGroup{}
	wait.Add(1)
	publisher.AsyncRequest([]byte("HELLO"), time.Millisecond*50, func(response messaging.Response) {
		defer wait.Done()
		if response.Error != nil {
			t.Error(response.Error)
		} else {
			if string(response.Data) != "HELLO" {
				t.Errorf("received different response : %v", string(response.Data))
			}
		}
	})
	wait.Wait()
	subscriber.Unsubscribe()
}

func TestPublisher_AsyncRequestWithContext(t *testing.T) {
	metrics.ResetRegistry()
	defer metrics.ResetRegistry()

	serverConfigs := natstest.CreateNATSServerConfigsNoTLS(1)
	servers := natstest.CreateNATSServers(t, serverConfigs)
	natstest.StartServers(servers)
	defer natstest.ShutdownServers(servers)

	connManager := nats.NewConnManager(natstest.ConnManagerSettings(serverConfigs[0]))
	client := nats.NewClient(connManager)
	defer client.CloseAllConns()

	conn, _ := client.Connect()

	topic := messaging.Topic(nuid.Next())

	subscriber, _ := conn.Subscribe(topic, nil)
	go func() {
		for subscriber.IsValid() {
			msg := <-subscriber.Channel()
			if msg != nil {
				conn.Publish(msg.ReplyTo.AsTopic(), msg.Data)
			}
		}
	}()

	publisher, _ := conn.Publisher(topic)
	wait := sync.WaitGroup{}
	wait.Add(1)
	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond)
	defer cancel()
	publisher.AsyncRequestWithContext(ctx, []byte("HELLO"), func(response messaging.Response) {
		defer wait.Done()
		if response.Error != nil {
			t.Error(response.Error)
		} else {
			if string(response.Data) != "HELLO" {
				t.Errorf("received different response : %v", string(response.Data))
			}
		}
	})
	wait.Wait()
	subscriber.Unsubscribe()
}

func TestPublisher_RequestChannel(t *testing.T) {
	metrics.ResetRegistry()
	defer metrics.ResetRegistry()

	serverConfigs := natstest.CreateNATSServerConfigsNoTLS(1)
	servers := natstest.CreateNATSServers(t, serverConfigs)
	natstest.StartServers(servers)
	defer natstest.ShutdownServers(servers)

	connManager := nats.NewConnManager(natstest.ConnManagerSettings(serverConfigs[0]))
	client := nats.NewClient(connManager)
	defer client.CloseAllConns()

	conn, _ := client.Connect()

	topic := messaging.Topic(nuid.Next())

	subscriber, _ := conn.Subscribe(topic, nil)
	go func() {
		for subscriber.IsValid() {
			msg := <-subscriber.Channel()
			if msg != nil {
				conn.Publish(msg.ReplyTo.AsTopic(), msg.Data)
			}
		}
	}()

	publisher, _ := conn.Publisher(topic)
	response := <-publisher.RequestChannel([]byte("HELLO"), time.Millisecond*50)
	if response.Error != nil {
		t.Error(response.Error)
	} else {
		if string(response.Data) != "HELLO" {
			t.Errorf("received different response : %v", string(response.Data))
		}
	}
	subscriber.Unsubscribe()
}

func TestPublisher_RequestChannelWithContext(t *testing.T) {
	metrics.ResetRegistry()
	defer metrics.ResetRegistry()

	serverConfigs := natstest.CreateNATSServerConfigsNoTLS(1)
	servers := natstest.CreateNATSServers(t, serverConfigs)
	natstest.StartServers(servers)
	defer natstest.ShutdownServers(servers)

	connManager := nats.NewConnManager(natstest.ConnManagerSettings(serverConfigs[0]))
	client := nats.NewClient(connManager)
	defer client.CloseAllConns()

	conn, _ := client.Connect()

	topic := messaging.Topic(nuid.Next())

	subscriber, _ := conn.Subscribe(topic, nil)
	go func() {
		for subscriber.IsValid() {
			msg := <-subscriber.Channel()
			if msg != nil {
				conn.Publish(msg.ReplyTo.AsTopic(), msg.Data)
			}
		}
	}()

	publisher, _ := conn.Publisher(topic)
	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*50)
	defer cancel()
	response := <-publisher.RequestChannelWithContext(ctx, []byte("HELLO"))
	if response.Error != nil {
		t.Error(response.Error)
	} else {
		if string(response.Data) != "HELLO" {
			t.Errorf("received different response : %v", string(response.Data))
		}
	}
	subscriber.Unsubscribe()
}
