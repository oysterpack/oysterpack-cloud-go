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

// log fields
const (
	LOG_FIELD_MESSAGE_ID   = "msg_id"
	LOG_FIELD_MESSAGE_TYPE = "msg_type"

	LOG_EVENT_STARTED logging.Event = "STARTED"
	LOG_EVENT_DYING   logging.Event = "DYING"
	LOG_EVENT_DEAD    logging.Event = "DEAD"

	LOG_EVENT_REGISTERED   logging.Event = "REGISTERED"
	LOG_EVENT_UNREGISTERED logging.Event = "UNREGISTERED"

	LOG_EVENT_MESSAGE_DROPPED        logging.Event = "MSG_DROPPED"
	LOG_EVENT_MESSAGE_PROCESSING_ERR logging.Event = "MSG_ERR"

	LOG_EVENT_ACTOR_NOT_FOUND logging.Event = "ACTOR_NOT_FOUND"
)
