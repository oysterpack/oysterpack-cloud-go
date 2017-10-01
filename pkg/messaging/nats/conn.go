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

	"fmt"

	"github.com/nats-io/go-nats"
	"github.com/oysterpack/oysterpack.go/pkg/messaging"
)

// NewConn is a constructor method for new nats based connections
func NewConn(connect Connect) (messaging.Conn, error) {
	nc, err := connect()
	if err != nil {
		return nil, err
	}
	return &conn{nc: nc}, nil
}

type conn struct {
	nc *ManagedConn
}

func (a *conn) Publish(topic messaging.Topic, data []byte) error {
	return a.nc.Publish(string(topic), data)
}

func (a *conn) PublishRequest(topic messaging.Topic, replyTo messaging.ReplyTo, data []byte) error {
	return a.nc.PublishRequest(string(topic), string(replyTo), data)
}

func (a *conn) PublishMessage(msg *messaging.Message) error {
	return a.PublishRequest(msg.Topic, msg.ReplyTo, msg.Data)
}

func (a *conn) Request(topic messaging.Topic, data []byte, timeout time.Duration) (*messaging.Message, error) {
	msg, err := a.nc.Request(string(topic), data, timeout)
	if err != nil {
		return nil, err
	}
	return toMessage(msg), nil
}

func (a *conn) RequestWithContext(ctx context.Context, topic messaging.Topic, data []byte) (response *messaging.Message, err error) {
	msg, err := a.nc.RequestWithContext(ctx, string(topic), data)
	if err != nil {
		return nil, err
	}
	return toMessage(msg), nil
}

func (a *conn) Subscribe(topic messaging.Topic, settings messaging.SubscriptionSettings) (messaging.Subscription, error) {
	if err := checkSubscriptionSettings(settings); err != nil {
		return nil, err
	}
	c := make(chan *messaging.Message)
	sub, err := a.nc.Subscribe(string(topic), func(msg *nats.Msg) {
		c <- toMessage(msg)
	})
	if settings.PendingLimits != nil {
		sub.SetPendingLimits(settings.MsgLimit, settings.BytesLimit)
	}
	if err != nil {
		return nil, err
	}
	return &subscription{sub, c}, nil
}

func (a *conn) QueueSubscribe(topic messaging.Topic, queue messaging.Queue, settings messaging.SubscriptionSettings) (messaging.QueueSubscription, error) {
	if err := checkSubscriptionSettings(settings); err != nil {
		return nil, err
	}
	c := make(chan *messaging.Message)
	sub, err := a.nc.QueueSubscribe(string(topic), string(queue), func(msg *nats.Msg) {
		c <- toMessage(msg)
	})
	if settings.PendingLimits != nil {
		sub.SetPendingLimits(settings.MsgLimit, settings.BytesLimit)
	}
	if err != nil {
		return nil, err
	}
	return &queueSubscription{&subscription{sub, c}, queue}, nil
}

func checkSubscriptionSettings(settings messaging.SubscriptionSettings) error {
	if settings.PendingLimits != nil {
		errorMsgFormat := " %q : Zero is not allowed. Any negative value means there is no limit."
		if settings.MsgLimit == 0 {
			return fmt.Errorf(errorMsgFormat, "MsgLimit")
		}
		if settings.BytesLimit == 0 {
			return fmt.Errorf(errorMsgFormat, "BytesLimit")
		}
	}
	return nil
}

func toMessage(msg *nats.Msg) *messaging.Message {
	return &messaging.Message{
		Topic:   messaging.Topic(msg.Subject),
		Data:    msg.Data,
		ReplyTo: messaging.ReplyTo(msg.Reply),
	}
}
