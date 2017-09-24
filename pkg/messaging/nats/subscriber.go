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
	"bytes"
	"encoding/gob"

	"github.com/nats-io/go-nats"
	"github.com/oysterpack/oysterpack.go/pkg/commons"
)

// MessageProcessor used to process messages received from NATS
// The nats connection is provided if needed to send a reply message.
// The provided decoder is provided for decoding the msg data into an app specific struct.
type MessageProcessor func(conn *nats.Conn, msg *nats.Msg, decoder *gob.Decoder)

type GobSubscriber struct {
	conn         *nats.Conn
	subscription *nats.Subscription
	subject      string

	buf      bytes.Buffer
	messages chan *nats.Msg
	decoder  *gob.Decoder
	process  MessageProcessor

	shutdown chan struct{}
}

func NewGobSubscriber(conn *nats.Conn, subject string, chanBufSize int, process MessageProcessor) (subscriber *GobSubscriber, err error) {
	subscriber = &GobSubscriber{
		conn:    conn,
		subject: subject,

		messages: make(chan *nats.Msg, chanBufSize),
		process:  process,

		shutdown: make(chan struct{}),
	}
	subscriber.decoder = gob.NewDecoder(&subscriber.buf)
	subscriber.subscription, err = subscriber.conn.ChanSubscribe(subject, subscriber.messages)
	if err != nil {
		subscriber = nil
		return
	}
	go subscriber.run()
	return
}

func (a *GobSubscriber) run() {
	for {
		select {
		case msg := <-a.messages:
			a.receive(msg)
		case <-a.shutdown:
			// drain and process any queued up messages
			for msg := range a.messages {
				a.receive(msg)
			}
			return
		}
	}
}

func (a *GobSubscriber) receive(msg *nats.Msg) {
	a.buf.Reset()
	a.buf.Write(msg.Data)
	a.process(a.conn, msg, a.decoder)
}

func (a *GobSubscriber) Close() {
	commons.CloseQuietly(a.shutdown)
	func() {
		defer commons.IgnorePanic()
		a.subscription.Unsubscribe()
	}()
}
