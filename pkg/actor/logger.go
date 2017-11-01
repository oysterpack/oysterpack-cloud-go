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

package actor

import "github.com/oysterpack/oysterpack.go/pkg/logging"

type pkgobject struct{}

var logger = logging.NewPackageLogger(pkgobject{})

// used for logging
const (
	// fields
	LOG_FIELD_ACTOR      = "actor"
	LOG_FIELD_ACTOR_PATH = "path"
	LOG_FIELD_ACTOR_ID   = "id"
	LOG_FIELD_CHANNEL    = "channel"
	LOG_FIELD_MSG_TYPE   = "msg_type"

	// events
	LOG_EVENT_STARTED   logging.Event = "STARTED"
	LOG_EVENT_RESTARTED logging.Event = "RESTARTED"

	LOG_EVENT_DYING logging.Event = "DYING"
	LOG_EVENT_DEAD  logging.Event = "DEAD"

	LOG_EVENT_NIL_MSG logging.Event = "NIL_MSG"

	LOG_EVENT_REGISTERED   logging.Event = "REGISTERED"
	LOG_EVENT_UNREGISTERED logging.Event = "UNREGISTERED"

	LOG_EVENT_UNKNOWN_CHANNEL      logging.Event = "UNKNOWN_CHANNEL"
	LOG_EVENT_UNKNOWN_MESSAGE_TYPE logging.Event = "UNKNOWN_MESSAGE_TYPE"
)
