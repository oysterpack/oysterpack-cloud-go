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

package natspoc_test

import (
	"testing"

	"github.com/nats-io/go-nats"

	"sync"

	"encoding/json"

	natsop "github.com/oysterpack/oysterpack.go/pkg/messaging/natspoc"

	"github.com/oysterpack/oysterpack.go/pkg/messaging/natstest"
)

func TestNewSubscriber(t *testing.T) {
	ts := natstest.RunServer()
	defer ts.Shutdown()

	subject := "ping"

	wait := sync.WaitGroup{}

	// http://nats.io/documentation/server/gnatsd-prune/
	// NATS automatically handles a slow consumer. If a client is not processing messages quick enough, the NATS server cuts it off.
	// because the publisher1 is firing messages at full speed, if it takes too long for NATS to deliver a message to the subscriber, the messages will be dropped
	const subscriberCount = 3
	for i := 0; i < subscriberCount; i++ {
		nc, _ := nats.Connect(nats.DefaultURL)
		defer nc.Close()
		subscriber, err := natsop.NewSubscriber(nc, subject, 256, logMessage(&wait, i, t))
		if err != nil {
			t.Errorf("Failed to create subscriber : %v", err)
			return
		}
		defer subscriber.Close()
	}

	nc, _ := nats.Connect(nats.DefaultURL)
	defer nc.Close()
	publisher1 := natsop.NewPublisher(nc, subject, 0)
	defer publisher1.Close()

	publisher2 := natsop.NewPublisher(nc, subject, 0)
	defer publisher2.Close()

	for i := 1; i <= 100; i++ {
		wait.Add(subscriberCount)
		person := &Person{"alfio", "zappala", i}
		if err := publisher1.Publish(person); err != nil {
			t.Errorf("Publish failed : %v", err)
		} else {
			t.Logf("published message : %v", *person)
		}

		wait.Add(subscriberCount)
		if err := publisher2.Publish(person); err != nil {
			t.Errorf("Publish failed : %v", err)
		} else {
			t.Logf("published message : %v", *person)
		}
	}
	wait.Wait()
}

func TestNewQueueSubscriber(t *testing.T) {
	ts := natstest.RunServer()
	defer ts.Shutdown()

	nc, _ := nats.Connect(nats.DefaultURL)
	defer nc.Close()

	subject := "ping"
	queue := "ping_queue"

	wait := sync.WaitGroup{}

	// http://nats.io/documentation/server/gnatsd-prune/
	// NATS automatically handles a slow consumer. If a client is not processing messages quick enough, the NATS server cuts it off.
	// because the publisher is firing messages at full speed, if it takes too long for NATS to deliver a message to the subscriber, the messages will be dropped
	const subscriberCount = 3
	for i := 0; i < subscriberCount; i++ {
		subscriber, err := natsop.NewQueueSubscriber(nc, subject, queue, 100, logMessage(&wait, i, t))
		if err != nil {
			t.Errorf("Failed to create subscriber : %v", err)
			return
		}
		defer subscriber.Close()
	}

	publisher := natsop.NewPublisher(nc, subject, 100)
	defer publisher.Close()

	for i := 1; i <= 100; i++ {
		wait.Add(1)
		person := Person{"alfio", "zappala", i}
		if err := publisher.Publish(person); err != nil {
			t.Errorf("Publish failed : %v", err)
		} else {
			t.Logf("published message : %v", person)
		}
	}
	wait.Wait()
}

func logMessage(wait *sync.WaitGroup, id int, t *testing.T) natsop.MessageProcessor {
	msgReceivedCounter := 0
	return func(conn *nats.Conn, msg *nats.Msg) {
		defer wait.Done()
		person := &Person{}
		json.Unmarshal(msg.Data, person)
		msgReceivedCounter++
		t.Logf("[%d] Received message #%d: %v : %d", id, msgReceivedCounter, *person, len(msg.Data))
	}
}
