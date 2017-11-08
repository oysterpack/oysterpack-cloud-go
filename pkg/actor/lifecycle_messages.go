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

import (
	"github.com/oysterpack/oysterpack.go/pkg/actor/msgs"
)

// Actor lifecycle
// STARTED -> STOPPING -> STOPPED
var (
	STARTED  = Started{}
	STOPPING = Stopping{}
	STOPPED  = Stopped{}
)

// Started is used to notify the MessageProcessor that Actor has been started.
// This provides the MessageProcessor an opportunity to initialize.
type Started struct {
	EmptyMessage
}

func (a Started) MessageType() MessageType {
	return MessageType(msgs.Started_TypeID)
}

// Stopping is used to notify the MessageProcessor that the Actor is dying.
// The MessageProcessor will not receive any more user messages. This provides the MessageProcessor an opportunity
// to perform any shutdown activities before the Actor initiates shutdown procedures.
// This event notification occurs before the children are killed.
type Stopping struct {
	EmptyMessage
}

func (a Stopping) MessageType() MessageType {
	return MessageType(msgs.Stopping_TypeID)
}

// Stopped notifies the MessageProcessor that the Actor has been stopped and all children have been killed.
type Stopped struct {
	EmptyMessage
}

func (a Stopped) MessageType() MessageType {
	return MessageType(msgs.Stopped_TypeID)
}
