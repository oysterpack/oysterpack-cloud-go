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
	"github.com/oysterpack/oysterpack.go/pkg/messaging"
)

type subscription struct {
	sub *nats.Subscription
	c   chan *messaging.Message
}

// Subject that represents this subscription. This can be different
// than the received subject inside a Msg if this is a wildcard.
func (a *subscription) Topic() messaging.Topic {
	return messaging.Topic(a.sub.Subject)
}

// AutoUnsubscribe will issue an automatic Unsubscribe that is processed by the server when max messages have been received.
// This can be useful when sending a request to an unknown number of subscribers.
func (a *subscription) AutoUnsubscribe(max int) error {
	return a.sub.Unsubscribe()
}

// MaxPending returns the maximum number of queued messages and queued bytes seen so far.
func (a *subscription) MaxPending() (int, int, error) {
	return a.sub.MaxPending()
}

// ClearMaxPending resets the maximums seen so far.
func (a *subscription) ClearMaxPending() error {
	return a.sub.ClearMaxPending()
}

// PendingMessages returns the number of queued messages and queued bytes in the client for this subscription.
func (a *subscription) Pending() (int, int, error) {
	return a.sub.Pending()
}

// PendingLimits returns the current limits for this subscription. If no error is returned, a negative value indicates
// that the given metric is not limited.
func (a *subscription) PendingLimits() (int, int, error) {
	return a.sub.PendingLimits()
}

// SetPendingLimits sets the limits for pending msgs and bytes for this subscription. Zero is not allowed.
// Any negative value means that the given metric is not limited.
func (a *subscription) SetPendingLimits(msgLimit, bytesLimit int) error {
	return a.sub.SetPendingLimits(msgLimit, bytesLimit)
}

// Delivered returns the number of delivered messages for this subscription.
func (a *subscription) Delivered() (int64, error) {
	return a.sub.Delivered()
}

// Dropped returns the number of known dropped messages for this subscription. This will correspond to messages
// dropped by violations of PendingLimits. If the server declares the connection a SlowConsumer, this number may not be valid.
func (a *subscription) Dropped() (int, error) {
	return a.sub.Dropped()
}

// IsValid returns a boolean indicating whether the subscription is still active. This will return false if the subscription has already been closed.
func (a *subscription) IsValid() bool {
	return a.sub.IsValid()
}

// Unsubscribe will remove interest in the given subject.
func (a *subscription) Unsubscribe() error {
	return a.sub.Unsubscribe()
}

// Channel is used to receive the messages subscribed to
func (a *subscription) Channel() <-chan *messaging.Message {
	return a.c
}

func (a *subscription) SubscriptionInfo() (info *messaging.SubscriptionInfo, err error) {
	delivered, err := a.Delivered()
	dropped, err := a.Dropped()
	maxPendingMsgs, maxPendingBytes, err := a.MaxPending()
	pendingMsgs, pendingBytes, err := a.Pending()
	pendingMsgsLimit, pendingBytesLimit, err := a.PendingLimits()
	if err != nil {
		return
	}
	info = &messaging.SubscriptionInfo{
		Topic:         a.Topic(),
		Delivered:     delivered,
		Dropped:       dropped,
		Valid:         a.IsValid(),
		MaxPending:    messaging.PendingMessages{maxPendingMsgs, maxPendingBytes},
		Pending:       messaging.PendingMessages{pendingMsgs, pendingBytes},
		PendingLimits: messaging.PendingMessages{pendingMsgsLimit, pendingBytesLimit},
	}
	return
}

type queueSubscription struct {
	messaging.Subscription
	queue messaging.Queue
}

func (a *queueSubscription) QueueSubscriptionInfo() (*messaging.QueueSubscriptionInfo, error) {
	info, err := a.SubscriptionInfo()
	if err != nil {
		return nil, err
	}
	return &messaging.QueueSubscriptionInfo{
		SubscriptionInfo: info,
		Queue:            a.Queue(),
	}, nil
}

// Queue : All subscriptions with the same name will form a distributed queue, and each message will only be
// processed by one member of the group. All messages sent to the corresponding Topic will be delivered to the queue.
func (a *queueSubscription) Queue() messaging.Queue {
	return a.queue
}
