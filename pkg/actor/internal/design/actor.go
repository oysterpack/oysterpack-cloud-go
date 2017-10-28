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
	//
	// If the actor system that it belongs to is clustered, then the following lookup is performed :
	// 	1. if the actor exists locally, then return a local actor ref
	//  2. if the actor exists remotely, then return a remote actor ref
	//  3. if the actor does not currently exist in the cluster, then return nil
	//
	// The returned ref will be bound to a specific actor instance.
	Ref(path []string) Ref

	// NewChild will create a new child actor and return it
	//
	// The following type of errors can occur :
	// 	- ChildAlreadyExists
	NewChild(settings Settings, name string) (Actor, error)

	Failures() Failures

	Management() Management

	SetBehavior(receive Receive)

	Metrics() Metrics
}

// System represents an actor system. Think of the system as the root actor.
// The root actor's name is the actor system's name.
type System interface {
	Actor

	// Inbox returns a new Inbox with the specified channel buffer size
	Inbox(chanBufSize int) Inbox

	// RefAddresses returns the addresses that have been requested. The corresponding Ref(s) are cached.
	// They are watched, and when they terminate, the cached Ref is removed.
	RefAddresses() []Address
}

// Management groups together actor management related operations.
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

	// Ping sends a Ping message to the actor, and will return the Pong on the channel
	// Multiple actors may respond in a cluster. The channel will be closed upon timeout.
	Ping(timeout time.Duration) <-chan Pong

	// Remote returns true if the referenced actor is remote
	Remote() bool
}

type Metrics interface {
	// TODO
}

// Ref is an actor reference.
// If the referenced actor dies, the ref will try to automatically reconnect in the background.
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
	HandleFailure(ctx MessageContext, err error) Directive
}

type Decider func(err error) Directive

func RestartDecider(_ error) Directive {
	return DIRECTIVE_RESTART
}

func ResumeDecider(_ error) Directive {
	return DIRECTIVE_RESUME
}

// Directive represents an enum. The directive specifies how the actor failure should be handled.
type Directive int

// Directive enum values
const (
	DIRECTIVE_RESUME = iota
	DIRECTIVE_RESTART
	DIRECTIVE_STOP
	DIRECTIVE_ESCALATE
)

type SupervisorStrategyFactory interface {
	DefaultSupervisorStrategy() SupervisorStrategy

	NewAllForOneStrategy(maxNrOfRetries int, withinDuration time.Duration, decider Decider) SupervisorStrategy

	NewExponentialBackoffStrategy(backoffWindow time.Duration, initialBackoff time.Duration) SupervisorStrategy

	NewOneForOneStrategy(maxNrOfRetries int, withinDuration time.Duration, decider Decider) SupervisorStrategy

	RestartingStrategy() SupervisorStrategy
}

// Inbox is used to receive messages outside of the actor system.
// An Inbox is an anonymous actor created by the actor system.
type Inbox interface {
	// Address returns the inbox address
	Address() Address

	// Messages is used to deliver messages that are received by the internal actor
	Messages() <-chan Message

	// Close will kill inbox actor and close the messages channel.
	Close()
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

type Pong struct {
	Address
}

func (a *Pong) SystemMessage() {}

type Failures interface {
	// TotalCount returns the total number of errors that have occurred across all instances.
	// For example, when an actor is restarted, a new instance is created. The count is reset, but the total count is not reset.
	TotalCount() int

	// LastFailureTime returns the last time a failure occurred for the actor
	LastFailureTime() time.Time
	// LastFailure is the last actor failure that occurred
	LastFailure() error

	// Count is the current failure count that is being tracked.
	Count() int

	// Reset count back to 0. The count is reset when a new instance is created.
	// It may also be be reset by a supervisor strategy, e.g., exponential back off strategy
	ResetCount()

	// ResetTime is the last time failures were reset
	ResetTime() time.Time

	// LastFailureSinceReset returns the last failure since being reset
	LastFailureSinceReset() error
}

type InstanceFactory func() Instance

type Instance interface {
	Receive(ctx MessageContext) error
}

type Message interface {
	encoding.BinaryMarshaler
	encoding.BinaryUnmarshaler
}

type MessageFactory interface {
	// NewMessage creates a new Message instance
	NewMessage() Message

	// Validate checks if the msg is compatible and valid for message types produced by this MessageFactory
	Validate(msg Message) error
}

// Envelope is a message envelope. Envelope is itself a message.
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

type Receive func(ctx MessageContext) error

type Settings interface {
	Factory() InstanceFactory
	SupervisorStrategy() SupervisorStrategy

	// Channels returns a map of ChannelSettings where the key is the channel name
	Channels() map[string]ChannelSettings
}

type ChannelMiddleware func(next Receive) Receive

type ChannelSettings interface {
	// Channel returns the channel name
	Channel() string

	// BufSize returns the channel buffer size
	BufSize() int

	// Middleware returns middleware that is applied to the actor's current behavior
	Middleware() []ChannelMiddleware

	// MessageFactory creates messages that the channel supports.
	// It is used to unmarshal incoming messages from remote actor refs.
	MessageFactory() MessageFactory
}
