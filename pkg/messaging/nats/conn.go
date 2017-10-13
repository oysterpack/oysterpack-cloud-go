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
func NewConn(managedConn *ManagedConn) (messaging.Conn, error) {
	if managedConn.IsClosed() {
		return nil, messaging.ErrConnectionIsClosed
	}
	return &conn{managedConn}, nil
}

type conn struct {
	managedConn *ManagedConn
}

func (a *conn) Tags() []string {
	return a.managedConn.Tags()
}

func (a *conn) ID() string {
	return a.managedConn.ID()
}

func (a *conn) Cluster() messaging.ClusterName {
	return a.managedConn.Cluster()
}

func (a *conn) Publish(topic messaging.Topic, data []byte) error {
	return a.managedConn.Publish(string(topic), data)
}

func (a *conn) PublishRequest(topic messaging.Topic, replyTo messaging.ReplyTo, data []byte) error {
	return a.managedConn.PublishRequest(string(topic), string(replyTo), data)
}

func (a *conn) Request(topic messaging.Topic, data []byte, timeout time.Duration) (*messaging.Message, error) {
	msg, err := a.managedConn.Request(string(topic), data, timeout)
	if err != nil {
		return nil, err
	}
	return toMessage(msg), nil
}

func (a *conn) AsyncRequest(topic messaging.Topic, data []byte, timeout time.Duration, handler func(messaging.Response)) {
	go func() {
		msg, err := a.Request(topic, data, timeout)
		handler(messaging.Response{msg, err})
	}()
}

func (a *conn) RequestChannel(topic messaging.Topic, data []byte, timeout time.Duration) <-chan messaging.Response {
	c := make(chan messaging.Response, 1)
	go func() {
		defer close(c)
		msg, err := a.Request(topic, data, timeout)
		c <- messaging.Response{msg, err}
	}()
	return c
}

func (a *conn) RequestWithContext(ctx context.Context, topic messaging.Topic, data []byte) (response *messaging.Message, err error) {
	msg, err := a.managedConn.RequestWithContext(ctx, string(topic), data)
	if err != nil {
		return nil, err
	}
	return toMessage(msg), nil
}

func (a *conn) AsyncRequestWithContext(ctx context.Context, topic messaging.Topic, data []byte, handler func(messaging.Response)) {
	go func() {
		msg, err := a.RequestWithContext(ctx, topic, data)
		handler(messaging.Response{msg, err})
	}()
}

func (a *conn) RequestChannelWithContext(ctx context.Context, topic messaging.Topic, data []byte) <-chan messaging.Response {
	c := make(chan messaging.Response, 1)
	go func() {
		defer close(c)
		msg, err := a.RequestWithContext(ctx, topic, data)
		c <- messaging.Response{msg, err}
	}()
	return c
}

func (a *conn) Subscribe(topic messaging.Topic, settings *messaging.SubscriptionSettings) (messaging.Subscription, error) {
	return a.managedConn.TopicSubscribe(topic, settings)
}

func (a *conn) QueueSubscribe(topic messaging.Topic, queue messaging.Queue, settings *messaging.SubscriptionSettings) (messaging.QueueSubscription, error) {
	return a.managedConn.TopicQueueSubscribe(topic, queue, settings)
}

func (a *conn) Publisher(topic messaging.Topic) (messaging.Publisher, error) {
	return a.managedConn.Publisher(topic)
}

// Close will close the connection to the server. This call will release all blocking calls
func (a *conn) Close() {
	a.managedConn.Close()
}

// Closed tests if a Conn has been closed.
func (a *conn) Closed() bool {
	return a.managedConn.IsClosed()
}

// Connected tests if a Conn is connected.
func (a *conn) Connected() bool {
	return a.managedConn.IsConnected()
}

// Reconnecting tests if a Conn is reconnecting.
func (a *conn) Reconnecting() bool {
	return a.managedConn.IsReconnecting()
}

// Status returns the connection status
func (a *conn) Status() (status messaging.ConnStatus) {
	switch a.managedConn.Status() {
	case nats.DISCONNECTED:
		status = messaging.DISCONNECTED
	case nats.CONNECTED:
		status = messaging.CONNECTED
	case nats.CLOSED:
		status = messaging.CLOSED
	case nats.RECONNECTING:
		status = messaging.RECONNECTING
	case nats.CONNECTING:
		status = messaging.CONNECTING
	default:
		logger.Panic().Int("status", int(a.managedConn.Status())).Msg("Unknown NATS connection status")
	}
	return
}

// LastError reports the last error encountered via the connection.
func (a *conn) LastError() (err *messaging.ConnErr) {
	if a.managedConn.LastError() != nil {
		err = &messaging.ConnErr{a.managedConn.LastError(), a.managedConn.lastErrorTime}
	}
	return
}

// MaxPayload returns the size limit that a message payload can have.
// This is set by the server configuration and delivered to the client upon connect.
func (a *conn) MaxPayload() int64 {
	return a.managedConn.MaxPayload()
}

func checkSubscriptionSettings(settings *messaging.SubscriptionSettings) error {
	if settings == nil {
		return nil
	}

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
