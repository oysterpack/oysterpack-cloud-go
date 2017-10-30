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
	"strings"
	"sync"
)

// System is an actor hierarchy.
type System struct {
	*Actor

	// all actors in the hierarchy stored by their id and by their path, i.e., the actor is stored twice in this map
	actors     map[string]*Actor
	actorsLock sync.RWMutex
}

func (a *System) registerActor(actor *Actor) {
	a.actorsLock.Lock()
	defer a.actorsLock.Unlock()
	a.actors[actor.id] = actor
	a.actors[a.addressPathKey(actor.path)] = actor
}

func (a *System) unregisterActor(path []string, id string) {
	a.actorsLock.Lock()
	defer a.actorsLock.Unlock()
	delete(a.actors, id)
	delete(a.actors, a.addressPathKey(path))
}

func (a *System) addressPathKey(path []string) string { return strings.Join(path, "") }

func (a *System) FindActor(address *Address) *Actor {
	a.actorsLock.RLock()
	defer a.actorsLock.RUnlock()
	if address.Id != "" {
		actor := a.actors[address.Id]
		if actor != nil {
			return actor
		}
		return a.actors[a.addressPathKey(address.Path)]
	}
	return a.actors[a.addressPathKey(address.Path)]
}

func SystemMessageProcessor() MessageProcessor {

	return MessageChannelHandlers{}
}

func HandlePing(ctx *MessageContext) error {
	replyTo := ctx.Envelope.replyTo
	if replyTo == nil {
		// TODO: log the ping
		return nil
	}

	actor := ctx.system.FindActor(replyTo.Address)
	if actor != nil {
		actor.Tell(ctx.MessageEnvelope(replyTo.Channel, &Pong{ctx.address}))
	} else {
		// TODO: log that the reply to actor was not found
	}

	return nil
}
