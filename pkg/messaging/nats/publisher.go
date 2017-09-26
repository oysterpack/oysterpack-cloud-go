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
	"encoding/json"
	"fmt"

	"github.com/nats-io/go-nats"
	"github.com/oysterpack/oysterpack.go/pkg/commons"
)

// Publisher encodes messages as gobs and publishes them to a subject
//
// The gobs implementation compiles a custom codec for each data type in the stream and is most efficient when a single
// Encoder is used to transmit a stream of values, amortizing the cost of compilation. In order to have a single Encoder,
// te bytes.Buffer, needs to be reset after the message is encoded. As a bonus, the bytes.Buffer memory storage is re-used
// throughout the publisher's lifetime - which minimizes allocations and garbage collection.
//
// For concurrent publishing, create a pool of publishers - each publishing from separate goroutines.
//
type Publisher struct {
	conn    *nats.Conn
	subject string

	messages chan *message

	shutdown chan struct{}
}

// NewPublisher creates a new Publisher.
func NewPublisher(conn *nats.Conn, subject string, chanBufSize int) *Publisher {
	publisher := &Publisher{
		conn:     conn,
		subject:  subject,
		messages: make(chan *message, chanBufSize),
		shutdown: make(chan struct{}),
	}
	go publisher.run()
	return publisher
}

type message struct {
	data    interface{}
	replyTo chan error
}

func (a *Publisher) run() {
	for {
		select {
		case msg := <-a.messages:
			a.publish(msg)
		case <-a.shutdown:
			a.closeMessagesChan()
			// drain queued messages
			for msg := range a.messages {
				a.publish(msg)
			}
			return
		}
	}
}

// Close is used to shutdown the backend goroutine that is publishing messages to NATS
func (a *Publisher) Close() {
	commons.CloseQuietly(a.shutdown)
}

func (a *Publisher) closeMessagesChan() {
	defer commons.IgnorePanic()
	close(a.messages)
}

func (a *Publisher) publish(msg *message) {
	bytes, err := json.Marshal(msg.data)
	if err != nil {
		msg.replyTo <- err
		return
	}
	if err := a.conn.Publish(a.subject, bytes); err != nil {
		msg.replyTo <- err
		return
	}
	close(msg.replyTo)
}

// Publish encodes the message as a gob and publishes it on the configured subject to NATS.
// An error is returned if encoding fails, or for a NATS publishing error.
func (a *Publisher) Publish(msg interface{}) (err error) {
	defer func() {
		if p := recover(); p != nil {
			err = fmt.Errorf("%v", p)
		}
	}()
	replyTo := make(chan error)
	a.messages <- &message{msg, replyTo}
	err = <-replyTo
	return
}
