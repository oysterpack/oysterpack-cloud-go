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
)

// Actor represents an actor in the Actor Model. Actors are objects which encapsulate state and behavior, they communicate exclusively by exchanging messages.
//
// The actors model provide a high level abstraction for writing concurrent and distributed systems. This approach simplifies
// the burden imposed on engineers, such as explicit locks and concurrent access to shared state, as actors receive messages synchronously.
//
// Actors communicate via messages over channels.
// All messages must be capnp messages to support distributed messaging.
// Actors are defined by an actor path. The actor path must be unique within a local process.
type Actor interface {
	// System the system that the actor belongs to
	System() ActorSystem

	// Address is where the actor lives. Actors receive messages on this address.
	Address()

	// Created is the actor creation time
	Created() time.Time

	// Parent actor parent. The root actor of the system will have no parent.
	Parent() Actor

	// ChildrenNames actor children names
	ChildrenNames() []string

	// Child returns the child actor for the specified name
	Child(name string) Actor

	Channels() Channels

	// Alive indicates if the actor is still alive.
	Alive() bool

	// Dead returns the channel that can be used to wait until the actor is dead
	Dead() <-chan struct{}

	// Dying returns the channel that can be used to wait until Actor.Kill() is called.
	Dying() <-chan struct{}

	// Err returns the death reason, or ErrStillAlive if the tomb is not in a dying or dead state.
	Err() (reason error)

	// Kill puts the actor in a dying state for the given reason, closes the Dying channel, and sets Alive to false.
	//
	// Although Kill may be called multiple times, only the first non-nil error is recorded as the death reason.
	// If reason is ErrDying, the previous reason isn't replaced even if nil. It's a runtime error to call Kill with
	// ErrDying if t is not in a dying state.
	Kill(reason error)

	// Wait blocks until the actor has finished running, and then returns the reason for their death.
	Wait() error
}

type Channels interface {
	// ChannelNames return the message channel names
	ChannelNames() []string

	// ChannelMessageType returns the type that the message sent on the channel will be converted to by this actor.
	ChannelMessageType(channel string) string

	// ChannelMessageTypes returns the supported types per channel
	ChannelMessageTypes(channel string) map[string]string

	// Send the message to the actor on the specified channel asynchronously. The method may block depending on the channel's buffer.
	//
	// An error can occur under the following conditions:
	// 	- validation fails - see Actor.ValidateMessage() for the types of errors that can be returned
	//  - the actor is not alive -> ActorNotAliveError
	Send(channel string, msg interface{}) error

	SendViaChannel(channel string, msgs <-chan interface{}) error

	// SendRequest sends a messages asynchronously to the actor. The actor is expected to send a reply back on the specified address.
	// `reply` must not be nil.
	SendRequest(channel string, msg interface{}, reply *ChannelAddress) error

	// SendRequestViaChannel sends messages asynchronously to the actor. The actor is expected to send a reply back on the specified address.
	// `reply` must not be nil.
	SendRequestViaChannel(channel string, msgs <-chan interface{}, reply *ChannelAddress) error

	// ValidateMessage will check if the message is valid for the channel, i.e., can it be sent on the channel
	//
	// The following types of errors may be returned:
	// 	- InvalidChannelError
	//	- InvalidChannelMessageTypeError
	//	- InvalidMessageError
	ValidateMessage(channel string, msg interface{}) error
}

type ActorRef interface {
	Address()

	// ChannelMessageType returns the type that the message sent on the channel will be converted to by this actor.
	ChannelMessageType(channel string) string

	// ChannelMessageTypes returns the supported types per channel
	ChannelMessageTypes(channel string) map[string]string

	// Message sends the message to the actor on the specified channel asynchronously. The operation is back pressured.
	// If the actor is not ready to receive the message, then the operation will be block
	//
	// An error can occur under the following conditions:
	// 	- validation fails - see Actor.ValidateMessage() for the types of errors that can be returned
	//  - the actor is not alive -> ActorNotAliveError
	Message(channel string, msg interface{}) error

	//
	Messages(channel string, msgs <-chan interface{}) error

	// Request sends a message asynchronously to the actor. The actor is expected to send a reply back on the specified address.
	// `reply` must not be nil.
	Request(channel string, msg interface{}, reply *ChannelAddress) error

	// Requests sends messages asynchronously to the actor. The actor is expected to send a reply back on the specified address.
	// `reply` must not be nil.
	Requests(channel string, msgs <-chan interface{}, reply *ChannelAddress) error

	// ValidateMessage will check if the message is valid for the channel, i.e., can it be sent on the channel
	//
	// The following types of errors may be returned:
	// 	- InvalidChannelError
	//	- InvalidChannelMessageTypeError
	//	- InvalidMessageError
	ValidateMessage(channel string, msg interface{}) error

	// Remote returns true if the actor is referencing a remote actor, i.e., outside the local actor system.
	// More accurately, the actor is accessed remotely. There may exist a local actor with the same address, but the
	// communication is remote, i.e., via the network.
	Remote() bool
}
