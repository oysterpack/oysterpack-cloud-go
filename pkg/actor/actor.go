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
	"gopkg.in/tomb.v2"
)

type Receive func(ctx *MessageContext) error

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
	path    []string
	name    string
	id      string
	created time.Time
	parent  *Actor

	children map[string]*Actor

	uid *nuid.NUID

	this        tomb.Tomb
	instance    Instance
	behavior    Receive
	receiveTomb tomb.Tomb

	failures *Failures

	deathChan chan error
}

func (a *Actor) run() error {
	go func() {
		<-a.this.Dead()
		a.deathChan <- a.this.Err()
	}()

	for {
		select {
		case <-a.this.Dying():

		}
	}

	return nil
}

func (a *Actor) System() System {
	return a.system
}

func (a *Actor) Name() string {
	return a.name
}

func (a *Actor) Path() []string {
	return a.path
}

func (a *Actor) Id() string {
	return a.id
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
