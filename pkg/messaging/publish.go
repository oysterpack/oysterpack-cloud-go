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

package messaging

import (
	"context"
	"time"
)

// TopicPublisher is used to publish data to the specified topic.
// Replies is used to communicate to the subscriber where to send replies to.
// Messages are sent in fire and forget fashion.
type TopicPublisher struct {
	conn    Conn
	topic   Topic
	replyTo ReplyTo
}

// NewTopicPublisher creates a new TopicPublisher
func NewTopicPublisher(conn Conn, topic Topic, replyTo ReplyTo) *TopicPublisher {
	return &TopicPublisher{conn, topic, replyTo}
}

// Cluster returns the name of the cluster that the topic belongs to
func (a *TopicPublisher) Cluster() ClusterName {
	return a.conn.Cluster()
}

// Topic is where messages are published to
func (a *TopicPublisher) Topic() Topic {
	return a.topic
}

// ReplyTo specified the name of the topic to send replies to
func (a *TopicPublisher) ReplyTo() ReplyTo {
	return a.replyTo
}

// Publish send the data to the above topic
func (a *TopicPublisher) Publish(data []byte) error {
	return a.conn.Publish(a.topic, data)
}

// RequestPublisher implements a request-response messaging model.
type RequestPublisher struct {
	conn  Conn
	topic Topic
}

// Cluster returns the name of the cluster that the topic belongs to
func (a *RequestPublisher) Cluster() ClusterName {
	return a.conn.Cluster()
}

// Topic is where messages are sent to.
func (a *RequestPublisher) Topic() Topic {
	return a.topic
}

// Request sends the request and waits for a response based on the specified timeout
func (a *RequestPublisher) Request(data []byte, timeout time.Duration) (response *Message, err error) {
	return a.conn.Request(a.topic, data, timeout)
}

// RequestWithContext sends the request and waits for a response based on the context timeout
func (a *RequestPublisher) RequestWithContext(ctx context.Context, data []byte) (response *Message, err error) {
	return a.conn.RequestWithContext(ctx, a.topic, data)
}

// AsyncRequest sends a message async using a request-response model.
// The handler is notified async with a response
func (a *RequestPublisher) AsyncRequest(data []byte, timeout time.Duration, handler func(Response)) {
	go func() {
		msg, err := a.Request(data, timeout)
		handler(Response{msg, err})
	}()
}

// RequestChannel sends a message async using a request-response model.
// The response is returned on the channel.
func (a *RequestPublisher) RequestChannel(data []byte, timeout time.Duration) <-chan Response {
	c := make(chan Response, 1)
	go func() {
		defer close(c)
		msg, err := a.Request(data, timeout)
		c <- Response{msg, err}
	}()
	return c
}

// AsyncRequestWithContext sends a message async using a request-response model.
// The handler is notified async with a response.
func (a *RequestPublisher) AsyncRequestWithContext(ctx context.Context, data []byte, handler func(Response)) {
	go func() {
		msg, err := a.RequestWithContext(ctx, data)
		handler(Response{msg, err})
	}()
}

// RequestChannelWithContext sends a message async using a request-response model.
// The response is returned on the channel.
func (a *RequestPublisher) RequestChannelWithContext(ctx context.Context, data []byte) <-chan Response {
	c := make(chan Response, 1)
	go func() {
		defer close(c)
		msg, err := a.RequestWithContext(ctx, data)
		c <- Response{msg, err}
	}()
	return c
}
