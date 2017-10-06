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
	"time"

	"github.com/json-iterator/go"
	"github.com/nats-io/go-nats"
	"github.com/oysterpack/oysterpack.go/pkg/messaging"
)

// Connect is an alias for a NATS connect function
type Connect func() (conn *ManagedConn, err error)

// Connect Options
var (
	// DefaultConnectTimeout is the default timeout used when creating a new NATS connection
	DefaultConnectTimeout   = nats.Timeout(5 * time.Second)
	DefaultReConnectTimeout = nats.ReconnectWait(2 * time.Second)
	AlwaysReconnect         = nats.MaxReconnects(-1)
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

// Events
const (
	EVENT_CONN_CLOSED             = "conn_closed"
	EVENT_CONN_DISCONNECT         = "conn_disconnect"
	EVENT_CONN_RECONNECT          = "conn_reconnect"
	EVENT_CONN_DISCOVERED_SERVERS = "conn_discovered_servers"
	EVENT_CONN_ERR                = "conn_err"
)

// log event fields
const (
	CONN_ID            = "conn_id"
	CONN_TAGS          = "tags"
	DISCOVERED_SERVERS = "discovered_servers"
	SUBSCRIPTION_VALID = "sub_valid"
	MAX_QUEUED_MSGS    = "max_queued_msgs"
	MAX_QUEUED_BYTES   = "max_queued_bytes"
	QUEUED_MSGS        = "queued_msgs"
	QUEUED_BYTES       = "queued_bytes"
	QUEUED_MSGS_LIMIT  = "queued_msgs_limit"
	QUEUED_BYTES_LIMIT = "queued_bytes_limit"
	DELIVERED          = "delivered"
	DROPPED            = "dropped"
	DISCONNECTS        = "disconnects"
	RECONNECTS         = "reconnects"
)

// common Conn tags
const (
	PUBLISHER        = "publisher"
	TOPIC_SUBSCRIBER = "topic_subscriber"
	QUEUE_SUBSCRIBER = "queue_subscriber"
)

func toMessage(msg *nats.Msg) *messaging.Message {
	return &messaging.Message{
		Topic:   messaging.Topic(msg.Subject),
		Data:    msg.Data,
		ReplyTo: messaging.ReplyTo(msg.Reply),
	}
}
