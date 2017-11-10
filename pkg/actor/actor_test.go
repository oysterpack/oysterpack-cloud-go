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

package actor_test

import (
	"testing"

	"github.com/oysterpack/oysterpack.go/pkg/actor"
	"github.com/rs/zerolog/log"
	"gopkg.in/tomb.v2"
)

func TestActor(t *testing.T) {
	system := actor.NewSystem()
	systemTomb := tomb.Tomb{}
	defer system.KillRootActors(systemTomb)

	fooProps := &actor.Props{
		MessageProcessor: &actor.MessageProcessorProps{
			Factory: func() actor.MessageProcessor {
				return actor.MessageHandlers{
					actor.PING_REQUEST.MessageType(): actor.PING_REQUEST_HANDLER,
				}
			},
			ErrorHandler: func(ctx *actor.MessageContext, err error) {
				ctx.Logger().Error().Err(err).Msg("RESUME")
			},
			ChannelSize: 1,
		},
		Logger: log.Logger,
	}
	foo, err := fooProps.NewRootActor(system, "foo")
	if err != nil {
		t.Fatal(err)
	}
	if !foo.Alive() {
		t.Fatal("foo is not alive")
	}

	inbox := make(chan *actor.Envelope)
	barProps := &actor.Props{
		MessageProcessor: &actor.MessageProcessorProps{
			Factory: func() actor.MessageProcessor {
				return actor.MessageHandlers{
					actor.PING_RESPONSE.MessageType(): &actor.MessageHandler{actor.InboxMessageHandler(inbox), actor.Unmarshaller(actor.PING_RESPONSE)},
				}
			},
			ErrorHandler: func(ctx *actor.MessageContext, err error) {
				ctx.Logger().Error().Err(err).Msg("RESUME")
			},
			ChannelSize: 512,
		},
		Logger: log.Logger,
	}
	bar, err := barProps.NewRootActor(system, "bar")
	if err != nil {
		t.Fatal(err)
	}
	if !bar.Alive() {
		t.Fatal("bar is not alive")
	}

	if count := system.ActorCount(); count != 2 {
		t.Errorf("Actor count should be 2 : %d", count)
	}

	pingRequest := foo.NewEnvelope(actor.PING_REQUEST, foo.Address(), &actor.ReplyTo{bar.Address(), actor.PING_RESPONSE.MessageType()}, nil)
	foo.Tell(pingRequest)
	pingResponse := <-inbox
	t.Logf("PingResponse : %v", pingResponse)
	if pingRequest.Id() != *pingResponse.CorrelationId() {
		t.Error("Response CorrelationId should match request message id.")
	}

	for i := 0; i < 10; i++ {
		foo.Tell(foo.NewEnvelope(actor.PING_REQUEST, foo.Address(), nil, nil))
	}
}
