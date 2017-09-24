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

	"github.com/nats-io/go-nats"

	"encoding/gob"
	"fmt"
	"sync"

	natsop "github.com/oysterpack/oysterpack.go/pkg/messaging/nats"
)

func TestNewGobSubscriber(t *testing.T) {
	nc, _ := nats.Connect(nats.DefaultURL)
	defer nc.Close()

	subject := "ping"

	wait := sync.WaitGroup{}

	// http://nats.io/documentation/server/gnatsd-prune/
	// NATS automatically handles a slow consumer. If a client is not processing messages quick enough, the NATS server cuts it off.
	// because the publisher is firing messages at full speed, if it takes too long for NATS to deliver a message to the subscriber, the messages will be dropped
	msgReceviedCounter := 0
	subscriber, err := natsop.NewGobSubscriber(nc, subject, 100, func(conn *nats.Conn, msg *nats.Msg, decoder *gob.Decoder) {
		defer wait.Done()
		person := &Person{}
		decoder.Decode(person)
		msgReceviedCounter++
		fmt.Printf("Received message #%d: %v\n", msgReceviedCounter, *person)
	})
	if err != nil {
		t.Errorf("Failed to create subscriber : %v", err)
		return
	}
	defer subscriber.Close()

	publisher := natsop.NewGobPublisher(nc, subject, 0)
	defer publisher.Close()

	const msgCount = 100

	for i := 1; i <= msgCount; i++ {
		wait.Add(1)
		person := &Person{"alfio", "zappala", i}
		if err := publisher.Publish(person); err != nil {
			t.Errorf("Publish failed : %v", err)
		} else {
			fmt.Printf("published message : %v\n", *person)
		}
	}
	wait.Wait()
}

func TestChannels(t *testing.T) {
	nc, _ := nats.Connect(nats.DefaultURL)
	ec, _ := nats.NewEncodedConn(nc, nats.GOB_ENCODER)
	defer ec.Close()

	type person struct {
		Name    string
		Address string
		Age     int
	}

	recvCh := make(chan *person)
	ec.BindRecvChan("hello", recvCh)

	sendCh := make(chan *person)
	ec.BindSendChan("hello", sendCh)

	me := &person{Name: "derek", Age: 22, Address: "140 New Montgomery Street"}

	// Send via Go channels
	sendCh <- me

	// Receive via Go channels
	t.Logf("%v", <-recvCh)
}
