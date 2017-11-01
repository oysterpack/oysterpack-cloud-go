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

	"github.com/rs/zerolog"
)

// System is an actor hierarchy.
type System struct {
	*Actor

	// all actors in the hierarchy stored - actor path is used as the key
	actors     map[string]*Actor
	actorsLock sync.RWMutex
}

func (a *System) registerActor(actor *Actor) {
	a.actorsLock.Lock()
	defer a.actorsLock.Unlock()
	a.actors[actor.id] = actor
	a.actors[a.addressPathKey(actor.path)] = actor
	LOG_EVENT_REGISTERED.Log(a.logger.Info()).
		Dict(LOG_FIELD_ACTOR, zerolog.Dict().Strs(LOG_FIELD_ACTOR_PATH, actor.path).Str(LOG_FIELD_ACTOR_ID, actor.id)).
		Msg("")
}

func (a *System) unregisterActor(path []string, id string) {
	a.actorsLock.Lock()
	defer a.actorsLock.Unlock()
	delete(a.actors, id)
	delete(a.actors, a.addressPathKey(path))
	LOG_EVENT_UNREGISTERED.Log(a.logger.Info()).
		Dict(LOG_FIELD_ACTOR, zerolog.Dict().Strs(LOG_FIELD_ACTOR_PATH, path).Str(LOG_FIELD_ACTOR_ID, id)).
		Msg("")
}

func (a *System) addressPathKey(path []string) string { return strings.Join(path, "") }

func (a *System) LookupActor(address *Address) *Actor {
	a.actorsLock.RLock()
	defer a.actorsLock.RUnlock()
	actor := a.actors[a.addressPathKey(address.Path)]
	if actor.id == address.Id || address.Id == "" {
		return actor
	}
	return nil
}

// LookupActorRef returns an actor reference that can be used to send messages to the actor
func (a *System) LookupActorRef(address *Address) Ref {
	return a.LookupActor(address)
}

func SystemMessageProcessor() MessageProcessor {
	return MessageChannelHandlers{
		CHANNEL_SYSTEM: MessageTypeHandlers{
			SYSTEM_MESSAGE_HEARTBEAT: HandleHearbeat,
			SYSTEM_MESSAGE_ECHO:      HandleEcho,
			SYSTEM_MESSAGE_PING_REQ:  HandlePingRequest,
		},
	}
}

func HandlePingRequest(ctx *MessageContext) error {
	replyTo := ctx.Envelope.replyTo
	if replyTo == nil {
		return nil
	}
	return ctx.Send(ctx.MessageEnvelope(replyTo.Channel, &PingResponse{ctx.address}), replyTo.Address)
}

func HandleHearbeat(ctx *MessageContext) error {
	replyTo := ctx.Envelope.replyTo
	if replyTo == nil {
		return nil
	}
	return ctx.Send(ctx.MessageEnvelope(replyTo.Channel, HEARTBEAT), replyTo.Address)
}

func HandleEcho(ctx *MessageContext) error {
	replyTo := ctx.Envelope.replyTo
	if replyTo == nil {
		return nil
	}
	return ctx.Send(ctx.MessageEnvelope(replyTo.Channel, HEARTBEAT), replyTo.Address)
}
