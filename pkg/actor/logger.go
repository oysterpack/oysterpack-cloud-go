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
	ACTOR_PATH = "path"
	CHANNEL    = "channel"
	MSG_TYPE   = "msg_type"

	// events
	ACTOR_STARTED logging.Event = "ACTOR_STARTED"
	ACTOR_DYING   logging.Event = "ACTOR_DYING"
	ACTOR_DEAD    logging.Event = "ACTOR_DEAD"

	MESSAGE_PROCESSOR_FAILURE         logging.Event = "MESSAGE_PROCESSOR_FAILURE"
	MESSAGE_PROCESSOR_CHANNEL_STARTED logging.Event = "MESSAGE_PROCESSOR_CHANNEL_STARTED"
	MESSAGE_PROCESSOR_CHANNEL_DYING   logging.Event = "MESSAGE_PROCESSOR_CHANNEL_DYING"
	MESSAGE_PROCESSOR_STARTED         logging.Event = "MESSAGE_PROCESSOR_STARTED"
	MESSAGE_PROCESSOR_DYING           logging.Event = "MESSAGE_PROCESSOR_DYING"
)
