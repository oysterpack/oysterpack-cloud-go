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
	"time"
)

// Conn represents a messaging connection
type Conn interface {
	// ID is a unique identifier assigned to the connection for tracking pusposes
	ID() string

	// Cluster returns the name of the cluster that the connection belongs to
	Cluster() ClusterName

	Publisher(topic Topic) (Publisher, error)

	// Subscribe creates a new async topic subscription with the specified settings
	Subscribe(topic Topic, settings *SubscriptionSettings) (Subscription, error)

	// QueueSubscribe creates a new async queue subscription with the specified settings
	QueueSubscribe(topic Topic, queue Queue, settings *SubscriptionSettings) (QueueSubscription, error)

	// Close will close the connection to the server. This call will release all blocking calls
	Close()

	// Closed tests if a Conn has been closed.
	Closed() bool

	// Connected tests if a Conn is connected.
	Connected() bool

	// Reconnecting tests if a Conn is reconnecting.
	Reconnecting() bool

	// LastError reports the last error encountered via the connection and when it occurred
	LastError() *ConnErr

	// MaxPayload returns the size limit that a message payload can have.
	// This is set by the server configuration and delivered to the client upon connect.
	MaxPayload() int64
}

// ConnErr contains a conn error and when it happened
type ConnErr struct {
	Error     error
	Timestamp time.Time
}
