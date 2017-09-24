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

// GobPublisher encodes messages as gobs and publishes them to a subject
//
// The gobs implementation compiles a custom codec for each data type in the stream and is most efficient when a single
// Encoder is used to transmit a stream of values, amortizing the cost of compilation. In order to have a single Encoder,
// te bytes.Buffer, needs to be reset after the message is encoded. As a bonus, the bytes.Buffer memory storage is re-used
// throughout the publisher's lifetime - which minimizes allocations and garbage collection.
//
// For concurrent publishing, create a pool of publishers - each publishing from separate goroutines.
//
type GobPublisher struct {
	conn    *nats.Conn
	subject string

	buf      bytes.Buffer
	messages chan *message
	encoder  *gob.Encoder

	shutdown chan struct{}
}

// NewGobPublisher creates a new GobPublisher.
func NewGobPublisher(conn *nats.Conn, subject string, chanBufSize int) *GobPublisher {
	publisher := &GobPublisher{
		conn:     conn,
		subject:  subject,
		messages: make(chan *message, chanBufSize),
		shutdown: make(chan struct{}),
	}
	publisher.encoder = gob.NewEncoder(&publisher.buf)
	go publisher.run()
	return publisher
}

type message struct {
	data    interface{}
	replyTo chan error
}

func (a *GobPublisher) run() {
	for {
		select {
		case msg := <-a.messages:
			a.publish(msg)
		case <-a.shutdown:
			return
		}
	}
}

// Close is used to shutdown the backend goroutine that is publishing messages to NATS
func (a *GobPublisher) Close() {
	commons.CloseQuietly(a.shutdown)
}

func (a *GobPublisher) publish(msg *message) {
	a.buf.Reset()
	if err := a.encoder.Encode(msg.data); err != nil {
		msg.replyTo <- err
		return
	}
	if err := a.conn.Publish(a.subject, a.buf.Bytes()); err != nil {
		msg.replyTo <- err
		return
	}
	close(msg.replyTo)
}

// Publish encodes the message as a gob and publishes it on the configured subject to NATS.
// An error is returned if encoding fails, or for a NATS publishing error.
func (a *GobPublisher) Publish(msg interface{}) error {
	replyTo := make(chan error)
	a.messages <- &message{msg, replyTo}
	return <-replyTo
}
