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

// system channels
const (
	CHANNEL_SYSTEM    = "0"
	CHANNEL_LIFECYCLE = "1"
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

	uid *nuid.NUID

	this        tomb.Tomb
	instance    Instance
	receive     Receive
	receiveTomb tomb.Tomb

	receiveChan chan *Envelope
	sysChan     chan *Envelope

	failures *Failures

	deathChan chan error

	logger zerolog.Logger

	settings *Settings
}

func (a *Actor) MessageEnvelope(channel string, msgType MessageType, msg Message) *Envelope {
	return NewEnvelope(a.UID, channel, msgType, msg, nil)
}

func (a *Actor) RequestEnvelope(channel string, msgType MessageType, msg Message, replyToChannel string) *Envelope {
	return NewEnvelope(a.UID, channel, msgType, msg, &ChannelAddress{Channel: replyToChannel, Address: a.address})
}

func (a *Actor) start() error {
	a.instance = a.settings.InstanceFactory()
	a.receive = a.instance.Receive

	go func() {
		err := a.this.Wait()
		a.receive(&MessageContext{Actor: a, Envelope: a.MessageEnvelope(CHANNEL_LIFECYCLE, MESSAGE_TYPE_STOPPED, STOPPED)})
		a.deathChan <- err
		close(a.deathChan)
	}()

	a.this.Go(func() error {
		for {
			select {
			case <-a.this.Dying():
				return a.receive(&MessageContext{Actor: a, Envelope: a.MessageEnvelope(CHANNEL_LIFECYCLE, MESSAGE_TYPE_STOPPING, STOPPING)})
			case msg := <-a.sysChan:
				a.handleSystemMessage(msg)
			}
		}
	})

	return nil
}

func (a *Actor) handleSystemMessage(msg *Envelope) {
	switch msg.msgType {
	case MESSAGE_TYPE_PING:
		if msg.replyTo != nil {
			pong := PONG_FACTORY.NewMessage().(*Pong)
			pong.Address = a.address
			a.MessageEnvelope(msg.replyTo.Channel, MESSAGE_TYPE_PONG, pong)
			// TODO: send message to replyTo address
		}

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
	return a.this.Alive()
}

// Dead returns the channel that can be used to wait until the actor is dead.
// The cause of death is returned on the channel.
func (a *Actor) Dead() <-chan error {
	return a.deathChan
}

// Dying returns the channel that can be used to wait until the actor is dying.
func (a *Actor) Dying() <-chan struct{} {
	return a.this.Dying()
}

// Kill puts the actor in a dying state for the given reason, closes the Dying channel, and sets Alive to false.
//
// Although Kill may be called multiple times, only the first non-nil error is recorded as the death reason.
// If reason is ErrDying, the previous reason isn'this replaced even if nil. It's a runtime error to call Kill with
// ErrDying if this is not in a dying state.
func (a *Actor) Kill(reason error) {
	a.this.Kill(reason)
}

// A restart only swaps the Actor instance defined by the Props but the incarnation and hence the UID remains the same.
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
