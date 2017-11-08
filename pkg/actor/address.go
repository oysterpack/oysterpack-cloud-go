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

	"github.com/oysterpack/oysterpack.go/pkg/actor/msgs"
	"zombiezen.com/go/capnproto2"
)

func NewAddress(path string, id string) Address {
	return Address{path, &id}
}

// Address refers to an actor address.
//
// Path is the actor path in the actor system hierarchy. The first element is the name of the actor system.
// The last element is the actor name
//
// Id is the unique actor id. The Id is optional. When it is not set it refers to any actor at the specified path.
//
// If id is not set, then the message is sent to any actor with a matching path.
// If id is set, then the message is sent to a specific actor with a matching id
//
// Note: in a cluster, there may be more than 1 actor with the same address on different nodes
type Address struct {
	Path string
	Id   *string
}

func (a *Address) Validate() error {
	if a.Path == "" {
		return errors.New("Address.Path is required")
	}

	if a.Id != nil && *a.Id == "" {
		return errors.New("Address.Id cannot be blank")
	}

	return nil
}

func (a *Address) ToCapnpMessage(seg *capnp.Segment) (msgs.Address, error) {
	addr, err := msgs.NewAddress(seg)
	if err != nil {
		return addr, err
	}
	if err := addr.SetPath(a.Path); err != nil {
		return addr, err
	}
	if a.Id != nil {
		if err := addr.SetId(*a.Id); err != nil {
			return addr, err
		}
	}
	return addr, nil
}
