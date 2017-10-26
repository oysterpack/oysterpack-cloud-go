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
	"time"

	"github.com/json-iterator/go"
)

type Message struct {
	*Header
	Data interface{}
}

//
type Header struct {
	// unique message id - for message tracking purposes
	Id      string
	Created time.Time
	Reply   ChannelAddress
}

func (a *Header) String() string {
	json, _ := jsoniter.Marshal(a)
	return string(json)
}

// Address refers to an actor address.
//
// At least one of path or id must be set. Either can be set, depending on the use case.
//  - If path is set, then the message is sent to any actor with a matching path.
//    - Note: in a cluster, there may be more than 1 actor instance with the same address on different nodes
// 	- If id is set, then the message is sent to a specific actor with a matching id
// 	- If both path and id are set, then the id is used and path is ignored
type Address struct {
	// actor system name
	System string

	// actor path
	Path []string
	// actor id
	Id string
}

func (a *Address) String() string {
	json, _ := jsoniter.Marshal(a)
	return string(json)
}

// ChannelAddress is where to deliver specific types of messages to an actor.
// Each message channel supports a specific type of message
type ChannelAddress struct {
	Channel string
	Address
}

func (a *ChannelAddress) String() string {
	json, _ := jsoniter.Marshal(a)
	return string(json)
}
