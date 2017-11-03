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

	"errors"
	"fmt"

	"github.com/rs/zerolog"
	"gopkg.in/tomb.v2"
)

func NewSystem(name string, logger zerolog.Logger) (*System, error) {
	if name != strings.TrimSpace(name) {
		return nil, fmt.Errorf("name cannot have whitespace passing %q", name)
	}

	if len(name) == 0 {
		return nil, errors.New("name cannot be blank")
	}

	system := &System{
		actors: make(map[string]*Actor),
		Actor: &Actor{
			path: []string{name},

			messageProcessorFactory: systemMessageProcessor,
			channelSettings:         systemChannelSettings,
			logger:                  logger,
		},
	}
	system.system = system

	if err := system.start(); err != nil {
		return nil, err
	}

	system.registerActor(system.Actor)

	// watch the system
	system.gaurdian.Go(func() error {
		for {
			// NOTE: the child may have been restarted. Thus, we always want to get the current MessageProcessorEngine.
			// MessageProcessorEngine access is protected by a RWMutex to enable safe concurrent access.
			msgProcessorEngine := system.messageProcessorEngine()
			select {
			case <-system.Dying():
				return nil
			case <-msgProcessorEngine.Dead():
				if err := msgProcessorEngine.Err(); err != nil {
					system.failures.failure(err)
					RESTART_ACTOR_STRATEGY(system.Actor, err)
				} else {
					return nil
				}
			}
		}
	})

	return system, nil
}

// System is an actor hierarchy.
type System struct {
	*Actor

	// all actors in the hierarchy stored - actor path is used as the key
	actors     map[string]*Actor
	actorsLock sync.RWMutex

	gaurdian tomb.Tomb
}

func (a *System) GaurdianAlive() bool {
	return a.gaurdian.Alive()
}

func (a *System) registerActor(actor *Actor) {
	a.actorsLock.Lock()
	defer a.actorsLock.Unlock()
	a.actors[actor.id] = actor
	a.actors[a.addressPathKey(actor.path)] = actor
	LOG_EVENT_REGISTERED.Log(a.logger.Debug()).
		Dict(LOG_FIELD_ACTOR, zerolog.Dict().Strs(LOG_FIELD_ACTOR_PATH, actor.path).Str(LOG_FIELD_ACTOR_ID, actor.id)).
		Msg("")
}

func (a *System) unregisterActor(path []string, id string) {
	a.actorsLock.Lock()
	defer a.actorsLock.Unlock()
	delete(a.actors, id)
	delete(a.actors, a.addressPathKey(path))
	LOG_EVENT_UNREGISTERED.Log(a.logger.Debug()).
		Dict(LOG_FIELD_ACTOR, zerolog.Dict().Strs(LOG_FIELD_ACTOR_PATH, path).Str(LOG_FIELD_ACTOR_ID, id)).
		Msg("")
}

func (a *System) addressPathKey(path []string) string { return strings.Join(path, "") }

func (a *System) LookupActor(address *Address) *Actor {
	a.actorsLock.RLock()
	defer a.actorsLock.RUnlock()
	actor := a.actors[a.addressPathKey(address.Path)]
	if actor == nil {
		return nil
	}
	if actor.id == address.Id || address.Id == "" {
		return actor
	}
	return nil
}

func systemMessageProcessor() MessageProcessor {
	return MessageChannelHandlers{
		CHANNEL_SYSTEM: MessageTypeHandlers{
			SYS_MSG_HEARTBEAT_REQ: HandleHearbeat,
			SYS_MSG_PING_REQ:      HandlePingRequest,
		},
	}
}

var (
	systemChannelSettings = map[Channel]*ChannelSettings{
		CHANNEL_SYSTEM: &ChannelSettings{
			Channel: CHANNEL_SYSTEM,
			BufSize: 0,
			ChannelMessageFactory: map[MessageType]func() Message{
				SYS_MSG_HEARTBEAT_REQ: func() Message { return HEARTBEAT_REQ },
				SYS_MSG_PING_REQ:      func() Message { return PING_REQ },
			},
		},
	}
)

// HandlePingRequest returns a PingResponse if a replyTo address was specified on the request
func HandlePingRequest(ctx *MessageContext) error {
	replyTo := ctx.Envelope.replyTo
	if replyTo == nil {
		return nil
	}
	return ctx.Send(ctx.MessageEnvelope(replyTo.Channel, SYS_MSG_PING_RESP, &PingResponse{ctx.address}), replyTo.Address)
}

// HandleHearbeat will send back a HeartbeatResponse if a replyTo address was specified
func HandleHearbeat(ctx *MessageContext) error {
	replyTo := ctx.Envelope.replyTo
	if replyTo == nil {
		return nil
	}
	return ctx.Send(ctx.RequestEnvelope(replyTo.Channel, SYS_MSG_HEARTBEAT_RESP, HEARTBEAT_RESP, CHANNEL_SYSTEM), replyTo.Address)
}
