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

// Subscription represents a topic subscriber's subscription.
// It is used to receive messages via the Channel.
// It is also used to track the subscription
type Subscription interface {
	ID() string

	// Cluster returns the name of the cluster that the subscription belongs to
	Cluster() ClusterName

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

// QueueSubscriptionInfo queue subscription info
type QueueSubscriptionInfo struct {
	*SubscriptionInfo
	Queue Queue
}

// PendingMessages returns counts for the number of pending messages and bytes
type PendingMessages struct {
	// number of messages pending
	Count int
	// number of bytes for the pending messages
	Bytes int
}
