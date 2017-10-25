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

	"sync"

	"fmt"

	"gopkg.in/tomb.v2"
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
	// each new actor instance is assigned a unique id
	address Address
	created time.Time

	system *ActorSystem

	parent   *Actor
	children map[string]*Actor

	// user messages are forwarded to the actor instance
	userMsgs chan interface{}
	// system messages are handled internally, i.e., by the Actor
	sysMsgs chan interface{}

	instance Instance
	behavior func(ctx *Context)

	tomb.Tomb

	msgSeqLock sync.Mutex
	msgSeq     uint64
}

type Instance interface {
	// Receive is the actor instance initial behavior
	Receive(ctx *Context)
}

// Id is the unique actor id within the actor system. It can be used to send a message directly to this instance.
// In a distributed environment, there may be multiple instances of the actor running - each running in their own separate process.
// Withing a local actor system, there is 1 actor instance per path. However, in a distributed actor system, there may be multiple
// actor instances per path - this is how the actor is scaled.
func (a *Actor) Id() string {
	return a.address.Id
}

// Created is when the actor was created.
func (a *Actor) Created() time.Time {
	return a.Created()
}

// System is the actor system that the actor belongs to
func (a *Actor) System() *ActorSystem {
	return a.system
}

// Path is the actor path within the actor system
func (a *Actor) Path() []string {
	return a.address.Path
}

// Name corresponds to the actor's name relative to its parent, i.e., it corresponds to the last path element.
func (a *Actor) Name() string {
	return a.address.Path[len(a.address.Path)-1]
}

// Parent actor
func (a *Actor) Parent() *Actor {
	return a.parent
}

// Children actors
func (a *Actor) Children() []*Actor {
	size := len(a.children)
	children := make([]*Actor, size)
	i := 0
	for _, child := range a.children {
		children[i] = child
		i++
	}
	return children
}

// ChildrenNames returns the names of the child actors
func (a *Actor) ChildrenNames() []string {
	size := len(a.children)
	children := make([]string, size)
	i := 0
	for name := range a.children {
		children[i] = name
		i++
	}
	return children
}

//Tell sends a message to the given address
func (a *Actor) Tell(to *Address, msg interface{}) error {
	return a.system.Tell(to, a.NewMessage(msg))
}

func (a *Actor) NewMessage(data interface{}) *Message {
	return &Message{
		&Header{Id: a.nextMsgId(), Created: time.Now(), Sender: a.address},
		data,
	}
}

func (a *Actor) nextMsgSequence() uint64 {
	a.msgSeqLock.Lock()
	defer a.msgSeqLock.Unlock()
	a.msgSeq++
	return a.msgSeq
}

func (a *Actor) nextMsgId() string {
	return fmt.Sprintf("%s%d", a.address.Id, a.nextMsgSequence())
}
