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

func (a *topicSubscriptions) collectMetrics() map[messaging.Topic]*subscriptionMetrics {
	a.removeInvalid()
	a.RLock()
	defer a.RUnlock()
	topicMetrics := map[messaging.Topic]*subscriptionMetrics{}
	for _, s := range a.subscriptions {
		if !s.IsValid() {
			continue
		}
		metrics := &subscriptionMetrics{topic: s.Topic()}
		msgs, bytes, err := s.Pending()
		if err == nil {
			metrics.pendingMsgs = msgs
			metrics.pendingBytes = bytes
		}

		msgs, bytes, err = s.MaxPending()
		if err == nil {
			metrics.pendingMsgs = msgs
			metrics.pendingBytes = bytes
		}

		dropped, err := s.Dropped()
		if err == nil {
			metrics.dropped = dropped
		}

		delivered, err := s.Delivered()
		if err == nil {
			metrics.delivered = delivered
		}

		aggregatedMetrics, exists := topicMetrics[metrics.topic]
		if exists {
			aggregatedMetrics.add(metrics)
		} else {
			topicMetrics[metrics.topic] = metrics
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

func (a *queueSubscriptions) collectMetrics() map[topicQueueKey]*queueSubscriptionMetrics {
	a.removeInvalid()
	a.RLock()
	defer a.RUnlock()
	queueMetrics := map[topicQueueKey]*queueSubscriptionMetrics{}
	for _, s := range a.subscriptions {
		metrics := &queueSubscriptionMetrics{subscriptionMetrics: &subscriptionMetrics{topic: s.Topic()}, queue: s.Queue()}

		msgs, bytes, err := s.Pending()
		if err == nil {
			metrics.pendingMsgs = msgs
			metrics.pendingBytes = bytes
		}

		msgs, bytes, err = s.MaxPending()
		if err == nil {
			metrics.pendingMsgsMax = msgs
			metrics.pendingBytesMax = bytes
		}

		dropped, err := s.Dropped()
		if err == nil {
			metrics.dropped = dropped
		}

		delivered, err := s.Delivered()
		if err == nil {
			metrics.delivered = delivered
		}

		key := topicQueueKey{s.Topic(), s.queue}
		aggregatedMetrics, exists := queueMetrics[key]
		if exists {
			aggregatedMetrics.add(metrics.subscriptionMetrics)
		} else {
			queueMetrics[key] = metrics
		}
	}
	return queueMetrics
}

type subscriptionMetrics struct {
	topic                           messaging.Topic
	pendingMsgs, pendingBytes       int
	pendingMsgsMax, pendingBytesMax int
	dropped                         int
	delivered                       int64
}

func (a *subscriptionMetrics) String() string {
	return fmt.Sprintf("%v", *a)
}

func (a *subscriptionMetrics) add(b *subscriptionMetrics) {
	a.pendingMsgs += b.pendingMsgs
	a.pendingBytes += b.pendingBytes
	a.dropped += b.dropped
	a.delivered += b.delivered

	if b.pendingMsgsMax > a.pendingMsgsMax {
		a.pendingMsgsMax = b.pendingMsgsMax
	}
	if b.pendingBytesMax > a.pendingBytesMax {
		a.pendingBytesMax = b.pendingBytesMax
	}
}

type queueSubscriptionMetrics struct {
	*subscriptionMetrics
	queue messaging.Queue
}

type topicQueueKey struct {
	topic messaging.Topic
	queue messaging.Queue
}
