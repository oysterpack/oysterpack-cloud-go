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

import "errors"

// An ActorSystem is a hierarchical group of actors.
//  It is also the entry point for creating or looking up actors.
type ActorSystem struct {
	// The name of this actor system, used to distinguish multiple ones within the same process
	Name string

	root *Actor
}

// Tell sends a message to the given address.
//
// If the address is for this actor system, then first try to send the message to a local actor.
// If no such local actor exists, then try sending it to a remote actor.
//
// If the address is for another actor system, then check if this actor system is linked to the specified actor system.
// If not, then return an error - actor system unknown
func (a *ActorSystem) Tell(address *Address, msg *Message) error {
	//TODO
	return errors.New("NOT IMPLEMENTED")
}
