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
	"sync"

	"fmt"

	"github.com/oysterpack/oysterpack.go/pkg/messaging"
)

type topicSubscriptions struct {
	sync.RWMutex
	// key = subscription.id
	subscriptions map[string]*subscription
}

func newTopicSubscriptions() *topicSubscriptions {
	return &topicSubscriptions{subscriptions: make(map[string]*subscription)}
}

func (a *topicSubscriptions) add(sub *subscription) {
	a.Lock()
	a.subscriptions[sub.id] = sub
	a.Unlock()
}

func (a *topicSubscriptions) remove(id string) {
	a.Lock()
	delete(a.subscriptions, id)
	a.Unlock()
}

func (a *topicSubscriptions) get(id string) *subscription {
	a.RLock()
	defer a.RUnlock()
	return a.subscriptions[id]
}

func (a *topicSubscriptions) removeInvalid() {
	a.Lock()
	invalidIds := []string{}
	for id, s := range a.subscriptions {
		if !s.IsValid() {
			invalidIds = append(invalidIds, id)
		}
	}
	for _, id := range invalidIds {
		delete(a.subscriptions, id)
	}
	a.Unlock()
}

func (a *topicSubscriptions) collectMetrics() map[messaging.Topic]*SubscriptionMetrics {
	a.removeInvalid()
	a.RLock()
	defer a.RUnlock()
	topicMetrics := map[messaging.Topic]*SubscriptionMetrics{}
	for _, s := range a.subscriptions {
		if !s.IsValid() {
			continue
		}
		metrics := &SubscriptionMetrics{Topic: s.Topic(), SubscriberCount: 1}
		msgs, bytes, err := s.Pending()
		if err == nil {
			metrics.PendingMsgs = msgs
			metrics.PendingBytes = bytes
		}

		msgs, bytes, err = s.MaxPending()
		if err == nil {
			metrics.PendingMsgs = msgs
			metrics.PendingBytes = bytes
		}

		dropped, err := s.Dropped()
		if err == nil {
			metrics.Dropped = dropped
		}

		delivered, err := s.Delivered()
		if err == nil {
			metrics.Delivered = delivered
		}

		aggregatedMetrics, exists := topicMetrics[metrics.Topic]
		if exists {
			aggregatedMetrics.add(metrics)
		} else {
			topicMetrics[metrics.Topic] = metrics
		}
	}
	return topicMetrics
}

type queueSubscriptions struct {
	sync.RWMutex
	subscriptions map[string]*queueSubscription
}

func newQueueSubscriptions() *queueSubscriptions {
	return &queueSubscriptions{subscriptions: make(map[string]*queueSubscription)}
}

func (a *queueSubscriptions) add(sub *queueSubscription) {
	a.Lock()
	a.subscriptions[sub.ID()] = sub
	a.Unlock()
}

func (a *queueSubscriptions) remove(id string) {
	a.Lock()
	delete(a.subscriptions, id)
	a.Unlock()
}

func (a *queueSubscriptions) get(id string) *queueSubscription {
	a.RLock()
	defer a.RUnlock()
	return a.subscriptions[id]
}

func (a *queueSubscriptions) removeInvalid() {
	a.Lock()
	invalidIds := []string{}
	for id, s := range a.subscriptions {
		if !s.IsValid() {
			invalidIds = append(invalidIds, id)
		}
	}
	for _, id := range invalidIds {
		delete(a.subscriptions, id)
	}
	a.Unlock()
}

func (a *queueSubscriptions) collectMetrics() map[TopicQueueKey]*QueueSubscriptionMetrics {
	a.removeInvalid()
	a.RLock()
	defer a.RUnlock()
	queueMetrics := map[TopicQueueKey]*QueueSubscriptionMetrics{}
	for _, s := range a.subscriptions {
		metrics := &QueueSubscriptionMetrics{SubscriptionMetrics: &SubscriptionMetrics{Topic: s.Topic(), SubscriberCount: 1}, Queue: s.Queue()}

		msgs, bytes, err := s.Pending()
		if err == nil {
			metrics.PendingMsgs = msgs
			metrics.PendingBytes = bytes
		}

		msgs, bytes, err = s.MaxPending()
		if err == nil {
			metrics.PendingMsgsMax = msgs
			metrics.PendingBytesMax = bytes
		}

		dropped, err := s.Dropped()
		if err == nil {
			metrics.Dropped = dropped
		}

		delivered, err := s.Delivered()
		if err == nil {
			metrics.Delivered = delivered
		}

		key := TopicQueueKey{s.Topic(), s.queue}
		aggregatedMetrics, exists := queueMetrics[key]
		if exists {
			aggregatedMetrics.add(metrics.SubscriptionMetrics)
		} else {
			queueMetrics[key] = metrics
		}
	}
	return queueMetrics
}

// SubscriptionMetrics aggregates subscription metrics for a topic
type SubscriptionMetrics struct {
	Topic                           messaging.Topic
	SubscriberCount                 int
	PendingMsgs, PendingBytes       int
	PendingMsgsMax, PendingBytesMax int
	Dropped                         int
	Delivered                       int64
}

func (a *SubscriptionMetrics) String() string {
	return fmt.Sprintf("%v", *a)
}

func (a *SubscriptionMetrics) add(b *SubscriptionMetrics) {
	if a.Topic != b.Topic {
		logger.Panic().Msgf("It is illegal to combine metrics from different topics - that would result in reporting false metrics")
	}

	a.SubscriberCount += b.SubscriberCount
	a.PendingMsgs += b.PendingMsgs
	a.PendingBytes += b.PendingBytes
	a.Dropped += b.Dropped
	a.Delivered += b.Delivered

	if b.PendingMsgsMax > a.PendingMsgsMax {
		a.PendingMsgsMax = b.PendingMsgsMax
	}
	if b.PendingBytesMax > a.PendingBytesMax {
		a.PendingBytesMax = b.PendingBytesMax
	}
}

// QueueSubscriptionMetrics aggregates subscription metrics for a topic queue
type QueueSubscriptionMetrics struct {
	*SubscriptionMetrics
	Queue messaging.Queue
}

// TopicQueueKey topic queue key
type TopicQueueKey struct {
	Topic messaging.Topic
	Queue messaging.Queue
}
