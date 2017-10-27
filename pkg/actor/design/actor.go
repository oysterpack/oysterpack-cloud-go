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

package design

import (
	"encoding"
	"time"
)

type Actor interface {
	System() System

	// Path is the actor path within the system
	Path() []string

	// Name is the leaf node in the actor path
	Name() string

	// Id is the unique actor id withing the system
	Id() string

	// Created is the actor creation timestamp
	Created() time.Time

	// Parent returns the parent actor.
	// The actor system root actor returns nil.
	Parent() Actor

	// Children returns any child actors
	Children() []Actor

	// Child returns a child actor by name.
	// If no such actor exists then nil is returned
	Child(name string) Actor

	SelfRef() Ref

	// Ref returns an actor reference with a path relative to this actor.
	// If the actor does not exist locally, it checks if the actor exists remotely.
	// If the no such actor exists locally or remotely, then nil is returned.
	Ref(path []string, mode RefMode) Ref

	// NewChild will create a new child actor and return it
	//
	// The following type of errors can occur :
	// 	- ChildAlreadyExists
	NewChild(props Props, name string) (Actor, error)

	RestartStatistics() RestartStatistics

	Management() Management

	SetBehavior(receive Receive)
}

// RefMode represents an enum. It indicates what mode to use to create the actor ref.
type RefMode int

// RefMode values
const (
	// means the ref will connect to a specific actor instance in the cluster - local is preferred.
	REFMODE_BY_ID = iota
	// means the ref will send messages to any actor in the cluster - local is preferred.
	REFMODE_BY_PATH
)

type System interface {
	Actor

	// Inbox returns a new Inbox with the specified channel buffer size
	Inbox(chanBufSize int) Inbox

	// RefAddresses returns the addresses that are currently in use and watched
	RefAddresses() []Address
}

// Management groups together actor management related operations.
// TODO: Metrics
type Management interface {
	// Alive indicates if the actor is still alive.
	Alive() bool

	// Dead returns the channel that can be used to wait until the actor is dead.
	// The cause of death is returned on the channel.
	Dead() <-chan error

	// Dying returns the channel that can be used to wait until the actor is dying.
	Dying() <-chan struct{}

	// Kill puts the actor in a dying state for the given reason, closes the Dying channel, and sets Alive to false.
	//
	// Although Kill may be called multiple times, only the first non-nil error is recorded as the death reason.
	// If reason is ErrDying, the previous reason isn't replaced even if nil. It's a runtime error to call Kill with
	// ErrDying if t is not in a dying state.
	Kill(reason error)

	// A restart only swaps the Actor instance defined by the Props but the incarnation and hence the UID remains the same.
	// As long as the incarnation is the same, you can keep using the same ActorRef.
	Restart()

	// Pause will pause message processing.
	// stash indicates the max number of messages that are received while paused to stash.
	// When the actor is resumed, the messages will be unstashed.
	// If stash <= 0, then messages will simply be dropped.
	Pause(stash int)

	// Resume will resume message processing
	Resume()

	// Paused returns true if actor is currently paused
	Paused() bool

	// Ping sends a Ping message to the actor, and will return the PingResponse on the channel
	// Multiple actors may respond in a cluster. The channel will be closed upon timeout.
	Ping(timeout time.Duration) <-chan PingResponse

	// Remote returns true if the referenced actor is remote
	Remote() bool
}

// Ref is an actor reference. The actor reference can be used in 2 modes :
//	1. Refer to a single actor instance by id
//
type Ref interface {
	Management() Management

	Address() Address

	ChannelNames() []string

	ValidateMessage(channel string, msg interface{}) error

	// Message is used to send messages to the actor
	Message(channel string, msg interface{}) error
	// Messages enables streaming messages to the actor over the channel.
	// The returned channel can be monitored for client message failures. For example, message may be invalid, or the actor may be dead.
	Messages(channel string, msgs <-chan interface{}) <-chan Failure

	// Message is used to send messages to the actor
	Request(channel string, msg interface{}, replyTo ChannelAddress) error
	// Messages enables streaming messages to the actor over the channel.
	// The returned channel can be monitored for client message failures. For example, message may be invalid, or the actor may be dead.
	Requests(channel string, msgs <-chan interface{}, replyTo ChannelAddress) <-chan Failure

	Channel(name string) Channel
}

type Channel interface {
	Address() ChannelAddress

	Remote() bool

	ValidateMessage(msg interface{}) error

	// Message is used to send messages to the actor
	Message(msg interface{}) error
	// Messages enables streaming messages to the actor over the channel.
	// The returned channel can be monitored for client message failures. For example, message may be invalid, or the actor may be dead.
	Messages(msgs <-chan interface{}) <-chan Failure

	// Message is used to send messages to the actor
	Request(msg interface{}, replyTo ChannelAddress) error
	// Messages enables streaming messages to the actor over the channel.
	// The returned channel can be monitored for client message failures. For example, message may be invalid, or the actor may be dead.
	Requests(msgs <-chan interface{}, replyTo ChannelAddress) <-chan Failure
}

// Address is used to locate an actor within the specified system
// Within a local actor system, the path will be unique. If there exists no actor at the specified address locally, then
// the address refers to a remote actor. For remote actors the semantics are different based on the Path and Id :
//	- if just path is set, then the address refers to all remote actors at the specified path, i.e., it refers to the actor pool
//  - if just id is set, then the address will refer to a specific remote actor instance
//	- if both path and id are set, then id will be used
type Address interface {
	// System is the actor system name
	System() string
	// Path is the actor path within the system
	Path() []string
	// Id is the actor id within the system
	Id() string
}

type ChannelAddress interface {
	Address

	Channel() string
}

type SupervisorStrategy interface {
}

type Decider func(err error) Directive

// Directive represents an enum. The directive specifies how the actor failure should be handled.
type Directive int

// Directive enum values
const (
	DIRECTIVE_RESUME = iota
	DIRECTIVE_RESTART
	DIRECTIVE_STOP
	DIRECTIVE_ESCALATE
)

type Inbox interface {
	Address() ChannelAddress
}

type SystemMessage interface {
	SystemMessage()
}

type LifeCycleMessage interface {
	SystemMessage
	LifeCycleMessage()
}

type Started struct{}

func (a *Started) SystemMessage()    {}
func (a *Started) LifeCycleMessage() {}

type Stopping struct{}

func (a *Stopping) SystemMessage()    {}
func (a *Stopping) LifeCycleMessage() {}

type Stopped struct{}

func (a *Stopped) SystemMessage()    {}
func (a *Stopped) LifeCycleMessage() {}

type Restarting struct{}

func (a *Restarting) SystemMessage()    {}
func (a *Restarting) LifeCycleMessage() {}

type Ping struct{}

func (a *Ping) SystemMessage() {}

type Failure struct {
	Message interface{}
	Err     error
}

func (a *Failure) SystemMessage() {}

type PingResponse struct {
	Address
}

func (a *PingResponse) SystemMessage() {}

type RestartStatistics interface {
}

type InstanceProducer interface {
}

type Instance interface {
	Receive(msgCtx MessageContext) error
}

type Message interface {
	encoding.BinaryMarshaler
	encoding.BinaryUnmarshaler
}

type MessageFactory func() Message

type Envelope interface {
	Message

	Channel() string
	Message() Message
	ReplyTo() ChannelAddress
}

type MessageContext interface {
	Actor
	Envelope() Envelope
}

type Receive func(msgCtx MessageContext) error

type Props interface {
}

type ChannelMiddleware func(next Receive) Receive
