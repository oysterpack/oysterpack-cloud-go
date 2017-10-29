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

	"github.com/nats-io/nuid"
	"github.com/rs/zerolog"
	"gopkg.in/tomb.v2"
)

type Receive func(ctx *MessageContext) error

type MessageContext struct {
	*Actor
	Envelope *Envelope
}

// system channels
const (
	CHANNEL_SYSTEM    = Channel("0")
	CHANNEL_LIFECYCLE = Channel("1")
)

// Actor represents an actor in the Actor Model. Actors are objects which encapsulate state and behavior, they communicate exclusively by exchanging messages.
//
// The actors model provide a high level abstraction for writing concurrent and distributed systems. This approach simplifies
// the burden imposed on engineers, such as explicit locks and concurrent access to shared state, as actors receive messages synchronously.
//
// Actors communicate via messages over channels.
// All messages must be capnp messages to support distributed messaging.
// Actors are defined by an actor path. The actor path must be unique within a local process.
type Actor struct {
	// immutable
	system  System
	address *Address
	name    string
	created time.Time
	parent  *Actor

	children map[string]*Actor

	// - used when starting the actor to create the initial MessageProcessor
	// - used when restarting the actor to create a new MessageProcessor instance
	messageProcessorFactory MessageProcessorFactory
	msgProcessor            MessageProcessor
	msgProcessorEngine      MessageProcessorEngine

	supervisorStrategy SupervisorStrategy
	failures           *Failures

	t      tomb.Tomb
	uid    *nuid.NUID
	logger zerolog.Logger
}

func (a *Actor) MessageEnvelope(channel Channel, msg Message) *Envelope {
	return NewEnvelope(a.UID, channel, msg, nil)
}

func (a *Actor) RequestEnvelope(channel Channel, msg Message, replyToChannel Channel) *Envelope {
	return NewEnvelope(a.UID, channel, msg, &ChannelAddress{Channel: replyToChannel, Address: a.address})
}

func (a *Actor) start() error {
	a.msgProcessor = a.messageProcessorFactory()

	return nil
}

func (a *Actor) handlePing(msg *Envelope) {
	if msg.replyTo != nil {
		pong := PONG_FACTORY.NewMessage().(*Pong)
		pong.Address = a.address
		a.MessageEnvelope(msg.replyTo.Channel, pong)
		// TODO: send message to replyTo address
	}
}

func (a *Actor) System() System {
	return a.system
}

func (a *Actor) Name() string {
	return a.name
}

func (a *Actor) Path() []string {
	return a.address.Path
}

func (a *Actor) Id() string {
	return a.address.Id
}

// UID returns a new unique id.
// It is used to generate message envelope ids.
func (a *Actor) UID() string {
	return a.uid.Next()
}

// Alive indicates if the actor is still alive.
func (a *Actor) Alive() bool {
	return a.t.Alive()
}

// Dead returns the channel that can be used to wait until the actor is dead.
// The cause of death is returned on the channel.
func (a *Actor) Dead() <-chan struct{} {
	return a.t.Dead()
}

// Dying returns the channel that can be used to wait until the actor is dying.
func (a *Actor) Dying() <-chan struct{} {
	return a.t.Dying()
}

// Kill puts the actor in a dying state for the given reason, closes the Dying channel, and sets Alive to false.
//
// Although Kill may be called multiple times, only the first non-nil error is recorded as the death reason.
// If reason is ErrDying, the previous reason isn'this replaced even if nil. It's a runtime error to call Kill with
// ErrDying if this is not in a dying state.
func (a *Actor) Kill(reason error) {
	a.t.Kill(reason)
}

// Err returns the death reason, or ErrStillAlive if the actor is not in a dying or dead state.
func (a *Actor) Err() error {
	err := a.t.Err()
	if err == tomb.ErrStillAlive {
		return ErrStillAlive
	}
	return err
}

// Wait blocks until the actor has finished running, and then returns the reason for its death.
func (a *Actor) Wait() error {
	return a.t.Wait()
}

// A restart only swaps the Actor msgProcessor defined by the Props but the incarnation and hence the UID remains the same.
// As long as the incarnation is the same, you can keep using the same ActorRef.
func (a *Actor) Restart() {
	//TODO
}

// Pause will pause message processing.
// stash indicates the max number of messages that are received while paused to stash.
// When the actor is resumed, the messages will be unstashed.
// If stash <= 0, then messages will simply be dropped.
func (a *Actor) Pause(stash int) {
	//TODO
}

// Resume will resume message processing
func (a *Actor) Resume() {
	//TODO
}
