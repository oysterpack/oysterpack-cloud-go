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

	"github.com/oysterpack/oysterpack.go/pkg/actor/msgs"
	"zombiezen.com/go/capnproto2"
)

func NewAddress(path []string, id *string) (*Address, error) {
	address := &Address{strings.Join(path, "/"), id}
	if err := address.Validate(); err != nil {
		return nil, err
	}
	return address, nil
}

// Address refers to an actor address.
//
//  - If id is not set, then the message is sent to any actor with a matching path.
//    - Note: in a cluster, there may be more than 1 actor with the same address on different nodes
// 	- If id is set, then the message is sent to a specific actor with a matching id
type Address struct {
	path string
	id   *string
}

// Path is the actor path in the actor system hierarchy.
// The first element is the name of the actor system.
// The last element is the actor name
func (a *Address) Path() string {
	return a.path
}

// Id is the unique actor id.
// The Id is optional. When it is not set it refers to any actor at the specified path.
func (a *Address) Id() *string {
	return a.id
}

func (a *Address) Validate() error {
	if a.path == "" {
		return errors.New("Address.Path is required")
	}

	if a.id != nil && *a.id == "" {
		return errors.New("Address.Id cannot be blank")
	}

	return nil
}

func (a *Address) ToCapnpMessage(seg *capnp.Segment) (msgs.Address, error) {
	addr, err := msgs.NewAddress(seg)
	if err != nil {
		return addr, err
	}
	if err := addr.SetPath(a.path); err != nil {
		return addr, err
	}
	if a.id != nil {
		if err := addr.SetId(*a.id); err != nil {
			return addr, err
		}
	}
	return addr, nil
}
