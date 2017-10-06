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
	"context"
	"time"

	"github.com/nats-io/go-nats"
	"github.com/oysterpack/oysterpack.go/pkg/messaging"
	"github.com/prometheus/client_golang/prometheus"
)

// NewPublisher creates a new Publisher.
// Topic will be trimmed. If topic is blank, then a panic will be triggered.
func NewPublisher(conn *ManagedConn, topic messaging.Topic) messaging.Publisher {
	topic = topic.TrimSpace()
	if err := topic.Validate(); err != nil {
		logger.Panic().Err(err).Msg("")
	}

	return &publisher{
		conn,
		topic,
		topicMsgsPublishedCounter.WithLabelValues(conn.cluster.String(), string(topic)),
	}
}

// Publisher is used to publish data to the specified topic.
// Replies is used to communicate to the subscriber where to send replies to.
// Messages are sent in fire and forget fashion.
type publisher struct {
	conn  *ManagedConn
	topic messaging.Topic

	publishCounter prometheus.Counter
}

// Cluster returns the name of the cluster that the topic belongs to
func (a *publisher) Cluster() messaging.ClusterName {
	return a.conn.Cluster()
}

// Topic is where messages are published to
func (a *publisher) Topic() messaging.Topic {
	return a.topic
}

// Publish send the data to the above topic
func (a *publisher) Publish(data []byte) error {
	err := a.conn.Publish(string(a.topic), data)
	if err == nil {
		a.publishCounter.Inc()
	}
	return err
}

// PublishRequest send the data to the above topic
func (a *publisher) PublishRequest(data []byte, replyTo messaging.ReplyTo) error {
	err := a.conn.PublishRequest(string(a.topic), string(replyTo), data)
	if err == nil {
		a.publishCounter.Inc()
	}
	return err
}

// Request sends the request and waits for a response based on the specified timeout
func (a *publisher) Request(data []byte, timeout time.Duration) (*messaging.Message, error) {
	msg, err := a.conn.Request(string(a.topic), data, timeout)
	if err != nil {
		if err == nats.ErrTimeout {
			// means request message was published, but we timed out waiting for a response message
			a.publishCounter.Inc()
		}
		return nil, err
	}
	return toMessage(msg), nil
}

// RequestWithContext sends the request and waits for a response based on the context timeout
func (a *publisher) RequestWithContext(ctx context.Context, data []byte) (response *messaging.Message, err error) {
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}
	msg, err := a.conn.RequestWithContext(ctx, string(a.topic), data)
	if err != nil {
		if err == context.DeadlineExceeded || err == context.Canceled {
			// means request message was published, but we either timed out waiting for a response message or the request was cancelled
			a.publishCounter.Inc()
		}
		return nil, err
	}
	return toMessage(msg), nil
}

// AsyncRequest sends a message async using a request-response model.
// The handler is notified async with a response
func (a *publisher) AsyncRequest(data []byte, timeout time.Duration, handler func(messaging.Response)) {
	go func() {
		msg, err := a.Request(data, timeout)
		handler(messaging.Response{msg, err})
	}()
}

// RequestChannel sends a message async using a request-response model.
// The response is returned on the channel.
func (a *publisher) RequestChannel(data []byte, timeout time.Duration) <-chan messaging.Response {
	c := make(chan messaging.Response, 1)
	go func() {
		defer close(c)
		msg, err := a.Request(data, timeout)
		c <- messaging.Response{msg, err}
	}()
	return c
}

// AsyncRequestWithContext sends a message async using a request-response model.
// The handler is notified async with a response.
func (a *publisher) AsyncRequestWithContext(ctx context.Context, data []byte, handler func(messaging.Response)) {
	go func() {
		msg, err := a.RequestWithContext(ctx, data)
		handler(messaging.Response{msg, err})
	}()
}

// RequestChannelWithContext sends a message async using a request-response model.
// The response is returned on the channel.
func (a *publisher) RequestChannelWithContext(ctx context.Context, data []byte) <-chan messaging.Response {
	c := make(chan messaging.Response, 1)
	go func() {
		defer close(c)
		msg, err := a.RequestWithContext(ctx, data)
		c <- messaging.Response{msg, err}
	}()
	return c
}
