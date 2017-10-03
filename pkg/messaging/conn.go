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

// Conn represents a messaging connection
type Conn interface {
	// Cluster returns the name of the cluster that the connection belongs to
	Cluster() ClusterName

	// Publish publishes data to the specified Topic
	Publish(topic Topic, data []byte) error

	// PublishRequest publishes data to the specified Topic and specifies which topic to reply to
	PublishRequest(topic Topic, replyTo ReplyTo, data []byte) error

	// PublishMessage publishes the message data to the specified topic, with an optional reply to
	PublishMessage(msg *Message) error

	Publisher(topic Topic) (*TopicPublisher,error)

	// Request sends a message using a request-response model.
	// The method will block and wait for a reponse until the timeout expires.
	Request(topic Topic, data []byte, timeout time.Duration) (response *Message, err error)

	// AsyncRequest sends a message async using a request-response model.
	// The handler is notified async with a response.
	AsyncRequest(topic Topic, data []byte, timeout time.Duration, handler func(Response))

	// AsyncRequest sends a message async using a request-response model.
	// The response is returned on the channel.
	RequestChannel(topic Topic, data []byte, timeout time.Duration) <-chan Response

	// RequestWithContext sends a message using a request-response model.
	// The method will block and wait for a reponse until the context is cancelled.
	RequestWithContext(ctx context.Context, topic Topic, data []byte) (response *Message, err error)

	// AsyncRequestWithContext sends a message async using a request-response model.
	// The handler is notified async with a response.
	AsyncRequestWithContext(ctx context.Context, topic Topic, data []byte, handler func(Response))

	// RequestChannelWithContext sends a message async using a request-response model.
	// The response is returned on the channel.
	RequestChannelWithContext(ctx context.Context, topic Topic, data []byte) <-chan Response

	// Subscribe creates a new async topic subscription with the specified settings
	Subscribe(topic Topic, settings SubscriptionSettings) (Subscription, error)

	// QueueSubscribe creates a new async queue subscription with the specified settings
	QueueSubscribe(topic Topic, queue Queue, settings SubscriptionSettings) (QueueSubscription, error)

	// Close will close the connection to the server. This call will release all blocking calls
	Close()

	// Closed tests if a Conn has been closed.
	Closed() bool

	// Connected tests if a Conn is connected.
	Connected() bool

	// Reconnecting tests if a Conn is reconnecting.
	Reconnecting() bool

	// LastError reports the last error encountered via the connection and when it occurred
	LastError() (error,time.Time)

	// MaxPayload returns the size limit that a message payload can have.
	// This is set by the server configuration and delivered to the client upon connect.
	MaxPayload() int64
}
