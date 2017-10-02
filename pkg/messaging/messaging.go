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

// Topic represents the name of a messaging topic
type Topic string

// ReplyTo is the topic name to send replies to
type ReplyTo string

// Message represents the message envelope
type Message struct {
	// Topic is the topic that the message is published to or received from
	Topic
	// ReplyTo is intended to support a request-response model
	ReplyTo
	// Data is the message data
	Data []byte
}

// Queue represents the name of a messaging queue
type Queue string

// Subscription represents a topic subscriber's subscription.
// It is used to receive messages via the Channel.
// It is also used to track the subscription
type Subscription interface {
	// Subject that represents this subscription. This can be different
	// than the received subject inside a Msg if this is a wildcard.
	Topic() Topic

	// AutoUnsubscribe will issue an automatic Unsubscribe that is processed by the server when max messages have been received.
	// This can be useful when sending a request to an unknown number of subscribers.
	AutoUnsubscribe(max int) error

	// MaxPending returns the maximum number of queued messages and queued bytes seen so far.
	MaxPending() (int, int, error)

	// ClearMaxPending resets the maximums seen so far.
	ClearMaxPending() error

	// PendingMessages returns the number of queued messages and queued bytes in the client for this subscription.
	Pending() (int, int, error)

	// PendingLimits returns the current limits for this subscription. If no error is returned, a negative value indicates
	// that the given metric is not limited.
	PendingLimits() (int, int, error)

	// SetPendingLimits sets the limits for pending msgs and bytes for this subscription. Zero is not allowed.
	// Any negative value means that the given metric is not limited.
	SetPendingLimits(msgLimit, bytesLimit int) error

	// Delivered returns the number of delivered messages for this subscription.
	Delivered() (int64, error)

	// Dropped returns the number of known dropped messages for this subscription. This will correspond to messages
	// dropped by violations of PendingLimits. If the server declares the connection a SlowConsumer, this number may not be valid.
	Dropped() (int, error)

	// IsValid returns a boolean indicating whether the subscription is still active. This will return false if the subscription has already been closed.
	IsValid() bool

	// Unsubscribe will remove interest in the given subject.
	Unsubscribe() error

	// Channel is used to receive the messages subscribed to
	Channel() <-chan *Message

	// SubscriptionInfo collects all info at once and returns it
	SubscriptionInfo() (*SubscriptionInfo, error)
}

// QueueSubscription represents a queue subscriber subscription
type QueueSubscription interface {
	Subscription

	// Queue : All subscriptions with the same name will form a distributed queue, and each message will only be
	// processed by one member of the group. All messages sent to the corresponding Topic will be delivered to the queue.
	Queue() Queue

	// QueueSubscriptionInfo collects all info at once and returns it
	QueueSubscriptionInfo() (*QueueSubscriptionInfo, error)
}

// Conn represents a messaging connection
type Conn interface {
	Publish(topic Topic, data []byte) error

	PublishRequest(topic Topic, replyTo ReplyTo, data []byte) error

	PublishMessage(msg *Message) error

	Request(topic Topic, data []byte, timeout time.Duration) (response *Message, err error)

	RequestWithContext(ctx context.Context, topic Topic, data []byte) (response *Message, err error)

	Subscribe(topic Topic, settings SubscriptionSettings) (Subscription, error)

	QueueSubscribe(topic Topic, queue Queue, settings SubscriptionSettings) (QueueSubscription, error)
}

// SubscriptionSettings are used for subscribing to topics or queues
type SubscriptionSettings struct {
	// OPTIONAL - if nil, then defaults are applied
	*PendingLimits
}

// PendingLimits are used to configure message buffering on the client subscriber side.
type PendingLimits struct {
	// Zero is not allowed. Any negative value means there is no limit.
	MsgLimit int
	// Zero is not allowed. Any negative value means tehre is no limit.
	BytesLimit int
}

// TopicPublisher is used to publish data to the specified topic.
// Replies is used to communicate to the subscriber where to send replies to.
// Messages are sent in fire and forget fashion.
type TopicPublisher interface {
	// Topic is where messages are published to
	Topic() Topic

	// ReplyTo
	ReplyTo() ReplyTo

	// Publish send the data to the above topic
	Publish(data []byte) error
}

// RequestPublisher implements a request-response messaging model.
type RequestPublisher interface {
	// Topic is where messages are sent to.
	Topic() Topic

	// Request sends the request and waits for a response based on the specified timeout
	Request(data []byte, timeout time.Duration) (response *Message, err error)

	// RequestWithContext sends the request and waits for a response based on the context timeout
	RequestWithContext(ctx context.Context, data []byte) (response *Message, err error)
}

// SubscriptionInfo topic subscription info
type SubscriptionInfo struct {
	Topic         Topic
	Delivered     int64
	Dropped       int
	Valid         bool
	MaxPending    PendingMessages
	Pending       PendingMessages
	PendingLimits PendingMessages
}

// SubscriptionInfo queue subscription info
type QueueSubscriptionInfo struct {
	*SubscriptionInfo
	Queue Queue
}

// PendingMessages returns counts for the number of pending messages and bytes
type PendingMessages struct {
	// number of messages pending
	Count     int
	// number of bytes for the pending messages
	Bytes int
}
