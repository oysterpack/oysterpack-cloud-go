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

	"time"

	"github.com/oysterpack/oysterpack.go/pkg/actor"
	"github.com/rs/zerolog/log"
)

func TestNewSystem(t *testing.T) {
	system, err := actor.NewSystem("oysterpack", log.Logger)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		log.Logger.Info().Msg("Killing system ...")
		system.Kill(nil)
		log.Logger.Info().Msg("Kill signalled. Waiting for system to terminate ...")
		if err := system.Wait(); err != nil {
			t.Error(err)
		}
		log.Logger.Info().Msg("System terminated")
	}()

	log.Logger.Info().Msgf("address = %v, name = %s, path = %v, id = %s", system.Address(), system.Name(), system.Path(), system.Id())

	if !system.Alive() {
		t.Error("system is not alive")
	}

	if system.Name() != "oysterpack" {
		t.Errorf("name did not match : %v", system.Name())
	}

	_, err = system.Spawn("bar",
		barMessageProcessor,
		[]*actor.ChannelSettings{
			{
				Channel: actor.CHANNEL_SYSTEM,
				ChannelMessageFactory: actor.ChannelMessageFactory{
					actor.SYS_MSG_HEARTBEAT_REQ: func() actor.Message {
						return actor.HEARTBEAT_REQ
					},
				},
			},
		},
		actor.RESTART_ACTOR_STRATEGY,
		log.Logger)

	if err != nil {
		t.Fatalf("Failed to spawn the bar actor : %v", err)
	}

	foo, err := system.Spawn("foo",
		func() actor.MessageProcessor {
			return fooMessageProcessor(10)
		},
		[]*actor.ChannelSettings{
			{
				Channel: actor.CHANNEL_INBOX,
				ChannelMessageFactory: actor.ChannelMessageFactory{
					actor.SYS_MSG_HEARTBEAT_RESP: func() actor.Message {
						return actor.HEARTBEAT_RESP
					},
				},
			},
			{
				Channel: actor.CHANNEL_LIFECYCLE,
				ChannelMessageFactory: actor.ChannelMessageFactory{
					actor.LIFECYCLE_MSG_STARTED: func() actor.Message {
						return actor.STARTED
					},
				},
			},
		},
		actor.RESTART_ACTOR_STRATEGY,
		log.Logger)

	if err != nil {
		t.Fatalf("Failed to spawn the foo actor : %v", err)
	}

	if err := foo.Wait(); err != nil {
		t.Errorf("foo actor died with an error : v", err)
	}

}

func fooMessageProcessor(count int) actor.MessageProcessor {
	return actor.MessageChannelHandlers{
		actor.CHANNEL_INBOX: actor.MessageTypeHandlers{
			// when a heartbeat response is received, send another heartbeat request
			actor.SYS_MSG_HEARTBEAT_RESP: func(ctx *actor.MessageContext) error {
				heartbeatReq := ctx.RequestEnvelope(ctx.Envelope.ReplyTo().Channel, actor.SYS_MSG_HEARTBEAT_REQ, actor.PING_REQ, actor.CHANNEL_INBOX)
				if err := ctx.Send(heartbeatReq, ctx.Envelope.ReplyTo().Address); err != nil {
					return err
				}
				count--
				ctx.Logger().Debug().Int("count", count).Msg("")
				if count == 0 {
					ctx.Logger().Debug().Msg("foo is done")
					ctx.Kill(nil)
				}
				return nil
			},
		},
		actor.CHANNEL_LIFECYCLE: actor.MessageTypeHandlers{
			actor.LIFECYCLE_MSG_STARTED: func(ctx *actor.MessageContext) error {
				// when the actor is started, send the bar actor a heartbeat request to initiate the communication
				var ref *actor.Actor
				for ; ref == nil; ref = ctx.System().LookupActor(&actor.Address{Path: append(ctx.System().Path(), "bar")}) {
					time.Sleep(100 * time.Millisecond)
				}
				heartbeatReq := ctx.RequestEnvelope(actor.CHANNEL_SYSTEM, actor.SYS_MSG_HEARTBEAT_REQ, actor.PING_REQ, actor.CHANNEL_INBOX)
				ref.Tell(heartbeatReq)
				return nil
			},
		},
	}
}

func barMessageProcessor() actor.MessageProcessor {
	return actor.MessageChannelHandlers{
		actor.CHANNEL_SYSTEM: actor.MessageTypeHandlers{
			actor.SYS_MSG_HEARTBEAT_REQ: actor.HandleHearbeat,
		},
	}
}

func BenchmarkActors(b *testing.B) {
	system, err := actor.NewSystem("oysterpack", log.Logger)
	if err != nil {
		b.Fatal(err)
	}
	defer func() {
		log.Logger.Info().Msg("Killing system ...")
		system.Kill(nil)
		log.Logger.Info().Msg("Kill signalled. Waiting for system to terminate ...")
		if err := system.Wait(); err != nil {
			b.Error(err)
		}
		log.Logger.Info().Msg("System terminated")
	}()

	_, err = system.Spawn("bar",
		barMessageProcessor,
		[]*actor.ChannelSettings{
			{
				Channel: actor.CHANNEL_SYSTEM,
				ChannelMessageFactory: actor.ChannelMessageFactory{
					actor.SYS_MSG_HEARTBEAT_REQ: func() actor.Message {
						return actor.HEARTBEAT_REQ
					},
				},
			},
		},
		actor.RESTART_ACTOR_STRATEGY,
		log.Logger)

	if err != nil {
		b.Fatalf("Failed to spawn the bar actor : %v", err)
	}

	b.Run("Heartbeat Messaging", func(b *testing.B) {
		foo, err := system.Spawn("foo",
			func() actor.MessageProcessor {
				return fooMessageProcessor(b.N)
			},
			[]*actor.ChannelSettings{
				{
					Channel: actor.CHANNEL_INBOX,
					ChannelMessageFactory: actor.ChannelMessageFactory{
						actor.SYS_MSG_HEARTBEAT_RESP: func() actor.Message {
							return actor.HEARTBEAT_RESP
						},
					},
				},
				{
					Channel: actor.CHANNEL_LIFECYCLE,
					ChannelMessageFactory: actor.ChannelMessageFactory{
						actor.LIFECYCLE_MSG_STARTED: func() actor.Message {
							return actor.STARTED
						},
					},
				},
			},
			actor.RESTART_ACTOR_STRATEGY,
			log.Logger)
		if err != nil {
			b.Fatalf("Failed to spawn the foo actor : %v", err)
		}
		b.ResetTimer()
		foo.Wait()
	})
}
