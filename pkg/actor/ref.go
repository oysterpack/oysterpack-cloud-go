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

// Ref is an actor reference. The referenced actor may be local or remote.
type Ref interface {
	// Address is used to send messages to the actor.
	// Address.Id is optional. If the Address.Id is not specified, then the message may be handled by any actor that is
	// registered at the specified path.
	Address() *Address

	// Channels that the actor supports
	Channels() []Channel

	// Tell sends a message with fire and forget semantics.
	Tell(msg *Envelope) error
}
