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

package nats

import (
	"github.com/nats-io/go-nats"
	"github.com/oysterpack/oysterpack.go/pkg/commons"
)

// MessageProcessor used to process messages received from NATS
// The nats connection is provided if needed to send a reply message.
// The provided decoder is provided for decoding the gob encoded msg data into an app specific struct.
type MessageProcessor func(conn *nats.Conn, msg *nats.Msg)

// Subscriber subcribes to the specified subject and relays messages to the associated MessageProcessor
type Subscriber struct {
	conn         *nats.Conn
	subscription *nats.Subscription
	subject      string
	queue        string

	messages chan *nats.Msg
	process  MessageProcessor

	shutdown chan struct{}
}

// NewSubscriber creates a new Subscriber
func NewSubscriber(conn *nats.Conn, subject string, chanBufSize int, process MessageProcessor) (subscriber *Subscriber, err error) {
	subscriber = &Subscriber{
		conn:    conn,
		subject: subject,

		messages: make(chan *nats.Msg, chanBufSize),
		process:  process,

		shutdown: make(chan struct{}),
	}
	subscriber.subscription, err = subscriber.conn.ChanSubscribe(subject, subscriber.messages)
	if err != nil {
		subscriber = nil
		return
	}
	go subscriber.run()
	return
}

// NewSubscriber creates a new Subscriber
func NewQueueSubscriber(conn *nats.Conn, subject string, queue string, chanBufSize int, process MessageProcessor) (subscriber *Subscriber, err error) {
	subscriber = &Subscriber{
		conn:    conn,
		subject: subject,
		queue:   queue,

		messages: make(chan *nats.Msg, chanBufSize),
		process:  process,

		shutdown: make(chan struct{}),
	}
	subscriber.subscription, err = subscriber.conn.ChanQueueSubscribe(subject, queue, subscriber.messages)
	if err != nil {
		subscriber = nil
		return
	}
	go subscriber.run()
	return
}

func (a *Subscriber) run() {
	for {
		select {
		case msg := <-a.messages:
			a.receive(msg)
		case <-a.shutdown:
			// drain and process any queued up messages
			for msg := range a.messages {
				a.receive(msg)
			}
			close(a.messages)
			return
		}
	}
}

func (a *Subscriber) receive(msg *nats.Msg) {
	a.process(a.conn, msg)
}

// Close triggers the goroutine to shutdown and cancels the NATS subject subscription.
func (a *Subscriber) Close() {
	commons.CloseQuietly(a.shutdown)
	func() {
		defer commons.IgnorePanic()
		a.subscription.Unsubscribe()
	}()
}
