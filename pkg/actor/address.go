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
	"errors"

	"strings"

	"fmt"

	"github.com/json-iterator/go"
	"github.com/oysterpack/oysterpack.go/pkg/actor/msgs"
	"zombiezen.com/go/capnproto2"
)

// Address refers to an actor address.
//
//  - If id is not set, then the message is sent to any actor with a matching path.
//    - Note: in a cluster, there may be more than 1 actor msgProcessor with the same address on different nodes
// 	- If id is set, then the message is sent to a specific actor with a matching id
// 	- If both path and id are set, then the id is used and path is ignored
type Address struct {
	// actor path - required
	Path []string
	// actor id
	Id string
}

func (a *Address) Validate() error {
	if len(a.Path) == 0 {
		return errors.New("Path is required")
	}
	if a.Id != strings.TrimSpace(a.Id) {
		return errors.New("Id must not have any whitespace padding")
	}
	for i := range a.Path {
		if a.Path[i] != strings.TrimSpace(a.Path[i]) {
			return errors.New("Path elements must not have any whitespace padding")
		}
		if a.Path[i] == "" {
			return fmt.Errorf("Path element [%d] cannot be blank : %v", i, a.Path)
		}
	}
	return nil
}

// Write the address to the specified address message
func (a *Address) Write(msg msgs.Address) error {
	if len(a.Path) > 0 {
		path, err := capnp.NewTextList(msg.Segment(), int32(len(a.Path)))
		if err != nil {
			return err
		}
		for i, p := range a.Path {
			if err := path.Set(i, p); err != nil {
				return err
			}
		}
		if err = msg.SetPath(path); err != nil {
			return err
		}
	}

	if a.Id != "" {
		if err := msg.SetId(a.Id); err != nil {
			return err
		}
	}

	return nil
}

func (a *Address) String() string {
	json, _ := jsoniter.Marshal(a)
	return string(json)
}

func NewAddress(addr msgs.Address) (*Address, error) {
	address := &Address{}
	if addr.HasPath() {
		pathList, err := addr.Path()
		if err != nil {
			return nil, err
		}
		address.Path = make([]string, pathList.Len())
		for i := 0; i < pathList.Len(); i++ {
			address.Path[i], err = pathList.At(i)
			if err != nil {
				return nil, err
			}
		}
	}

	if addr.HasId() {
		id, err := addr.Id()
		if err != nil {
			return nil, err
		}
		address.Id = id
	}
	return address, nil
}

type Channel string

func (a Channel) String() string {
	return string(a)
}

func (a Channel) Validate() error {
	if strings.TrimSpace(a.String()) == "" {
		return errors.New("Channel cannot be blank")
	}
	return nil
}

// ChannelAddress is where to deliver specific types of messages to an actor.
// Each message channel supports a specific type of message
type ChannelAddress struct {
	Channel Channel
	*Address
}

// Validate the channel address. If channel address is nil, then skip validation.
func (a *ChannelAddress) Validate() error {
	if a == nil {
		return nil
	}

	if err := a.Channel.Validate(); err != nil {
		return err
	}

	if a.Address == nil {
		return errors.New("ChannelAddress.Address is required")
	}

	if err := a.Address.Validate(); err != nil {
		return err
	}

	return nil
}

func (a *ChannelAddress) String() string {
	json, _ := jsoniter.Marshal(a)
	return string(json)
}

func (a *ChannelAddress) Write(msg msgs.ChannelAddress) error {
	if err := msg.SetChannel(a.Channel.String()); err != nil {
		return err
	}

	address, err := msgs.NewAddress(msg.Segment())
	if err != nil {
		return err
	}
	a.Address.Write(address)
	if err = msg.SetAddress(address); err != nil {
		return err
	}

	return nil
}

func NewChannelAddress(addr msgs.ChannelAddress) (*ChannelAddress, error) {
	channelAddr := &ChannelAddress{}
	channel, err := addr.Channel()
	if err != nil {
		return nil, err
	}
	channelAddr.Channel = Channel(channel)
	replytoAddr, err := addr.Address()
	if err != nil {
		return nil, err
	}
	channelAddr.Address, err = NewAddress(replytoAddr)
	if err != nil {
		return nil, err
	}
	return channelAddr, nil
}
