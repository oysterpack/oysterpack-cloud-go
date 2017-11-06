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

package actor2_test

import (
	"testing"

	"time"

	"github.com/oysterpack/oysterpack.go/pkg/actor2"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func TestNewSystem(t *testing.T) {
	system, err := actor2.NewSystem("oysterpack", log.Logger)
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

	log.Logger.Info().Msgf("address = %v, name = %s, path = %v, id = %s", system.Address(), system.Name(), system.Address().Path, system.Address().Id)

	if !system.Alive() {
		t.Error("system is not alive")
	}

	if system.Name() != "oysterpack" {
		t.Errorf("name did not match : %v", system.Name())
	}

	bar, err := system.Spawn("bar",
		barMessageProcessor,
		actor2.RESTART_ACTOR_STRATEGY,
		log.Logger.Level(zerolog.DebugLevel),
		nil)

	if err != nil {
		t.Fatalf("Failed to spawn the bar actor : %v", err)
	}

	foo, err := system.Spawn("foo",
		func() actor2.MessageProcessor {
			return fooMessageProcessor(10, bar.Address())
		},
		actor2.RESTART_ACTOR_STRATEGY,
		log.Logger,
		nil)

	if err != nil {
		t.Fatalf("Failed to spawn the foo actor : %v", err)
	}

	foo.Tell(system.NewEnvelope(actor2.CHANNEL_INBOX, actor2.LIFECYCLE_MSG_STARTED, actor2.STARTED, nil, ""))

	if err := foo.Wait(); err != nil {
		t.Errorf("foo actor died with an error : v", err)
	}

	foo, err = system.Spawn("foo2",
		func() actor2.MessageProcessor {
			return fooMessageProcessor(10, bar.Address())
		},
		actor2.RESTART_ACTOR_STRATEGY,
		log.Logger,
		nil)
	if err != nil {
		t.Fatalf("Failed to spawn the foo actor : %v", err)
	}
	log.Logger.Debug().Dict("foo2", zerolog.Dict().Strs("path", foo.Address().Path).Str("id", foo.Address().Id)).Msg("")
	for i := 0; i < 10; i++ {
		system.Send(system.NewEnvelope(actor2.CHANNEL_SYSTEM, actor2.SYS_MSG_HEARTBEAT_REQ, actor2.HEARTBEAT_REQ, nil, ""), foo.Address())
	}

	foo, err = system.Spawn("foo3",
		func() actor2.MessageProcessor {
			return &FooMessageProcessor{10, system.Address()}
		},
		actor2.RESTART_ACTOR_STRATEGY,
		log.Logger,
		nil)
	if err != nil {
		t.Fatalf("Failed to spawn the foo actor : %v", err)
	}
	log.Logger.Debug().Dict("foo", zerolog.Dict().Strs("path", foo.Address().Path).Str("id", foo.Address().Id)).Msg("")

	//for i := 0; i < 10; i++ {
	//	log.Logger.Debug().Int("b.N", i).Msg("system.Tell(HeartbeatRequest,foo.Address()) - switch based")
	//	msg := system.NewEnvelope(actor.CHANNEL_SYSTEM, actor.SYS_MSG_HEARTBEAT_REQ, actor.HEARTBEAT_REQ, nil, "")
	//	//system.Send(msg, foo.Address())
	//	foo.Tell(msg)
	//}

}

func fooMessageProcessor(count int, address *actor2.Address) actor2.MessageProcessor {
	return actor2.MessageHandlers{
		actor2.MessageChannelKey{actor2.CHANNEL_INBOX, actor2.SYS_MSG_HEARTBEAT_RESP}: actor2.MessageHandler{func(ctx *actor2.MessageContext) error {
			count--
			ctx.Logger().Debug().Int("count", count).Msg("")
			if count == 0 {
				ctx.Logger().Debug().Msg("foo is done")
				ctx.Kill(nil)
				return nil
			}

			// when a heartbeat response is received, send another heartbeat request
			replyTo, _ := ctx.ChannelAddress(actor2.CHANNEL_INBOX)
			heartbeatReq := ctx.NewEnvelope(actor2.CHANNEL_SYSTEM, actor2.SYS_MSG_HEARTBEAT_REQ, actor2.HEARTBEAT_REQ, replyTo, "")
			if err := ctx.Send(heartbeatReq, address); err != nil {
				ctx.Logger().Error().Err(err).Msg("")
				return err
			}

			return nil
		},
			actor2.UnmarshalHeartbeatResponse,
		},
		actor2.MessageChannelKey{actor2.CHANNEL_INBOX, actor2.LIFECYCLE_MSG_STARTED}: actor2.MessageHandler{
			func(ctx *actor2.MessageContext) error {
				ctx.Logger().Debug().Msg("STARTING")
				// when the actor is started, send the bar actor a heartbeat request to initiate the communication
				var ref *actor2.Actor
				for ; ref == nil; ref = ctx.System().LookupActor(address) {
					ctx.Logger().Debug().Msgf("DID NOT FIND : %v", address)
					time.Sleep(100 * time.Millisecond)

				}

				replyTo, _ := ctx.ChannelAddress(actor2.CHANNEL_INBOX)
				heartbeatReq := ctx.NewEnvelope(actor2.CHANNEL_SYSTEM, actor2.SYS_MSG_HEARTBEAT_REQ, actor2.PING_REQ, replyTo, ctx.Message.Id())
				return ref.Tell(heartbeatReq)
			},
			actor2.UnmarshalStartedMessage,
		},
		actor2.MessageChannelKey{actor2.CHANNEL_SYSTEM, actor2.SYS_MSG_HEARTBEAT_REQ}: actor2.MessageHandler{
			actor2.HandleHearbeatRequest,
			actor2.UnmarshalHeartbeatRequest,
		},
	}
}

func barMessageProcessor() actor2.MessageProcessor {
	return actor2.MessageHandlers{
		actor2.MessageChannelKey{actor2.CHANNEL_SYSTEM, actor2.SYS_MSG_HEARTBEAT_REQ}: actor2.MessageHandler{
			actor2.HandleHearbeatRequest,
			actor2.UnmarshalHeartbeatRequest,
		},
	}
}

// Benchmark results
// =================
//
// BenchmarkActorMessaging/foo_<--HeartBeat-->_bar_-_MessageChannelHandlers_based-8                  300000              4545 ns/op             352 B/op          7 allocs/op
// BenchmarkActorMessaging/foo_<--HeartBeat-->_bar_-_switch_based-8                                  300000              4175 ns/op             384 B/op          9 allocs/op
//
// BenchmarkActorMessaging/system.Tell(HeartbeatRequest)-8                                          1000000              2223 ns/op             160 B/op          3 allocs/op
// BenchmarkActorMessaging/system.Send(HeartbeatRequest)-8                                          1000000              2556 ns/op             160 B/op          3 allocs/op
// BenchmarkActorMessaging/bar.Tell(HeartbeatRequest)-8                                              500000              2606 ns/op             160 B/op          3 allocs/op

// BenchmarkActorMessaging/foo_<--HeartBeat-->_bar_-_buffered_chan-8                                 300000              4613 ns/op             352 B/op          7 allocs/op
// BenchmarkActorMessaging/bar.Tell(HeartbeatRequest)_-_buffered_chan-8                             1000000              1290 ns/op             160 B/op          2 allocs/op
// BenchmarkActorMessaging/bar.Send(HeartbeatRequest)_-_buffered_chan-8                             1000000              1186 ns/op             160 B/op          2 allocs/op
//
// Take aways
// 	1. go maps are super fast - Send performs an actor lookup, and then sends the message via Actor.Tell(). The map lookup added no noticeable overhead
//  2. Buffered channels boost throughput (in this case).
// 	   - NOTE: if the message consumer is much slower than the message producer, then buffering would add no value.
// 		 The value added by buffering channels must be proven through benchmarking and performance testing.
func BenchmarkActorMessaging(b *testing.B) {
	system, err := actor2.NewSystem("oysterpack", log.Logger)
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

	bar, err := system.Spawn("bar",
		barMessageProcessor,
		actor2.RESTART_ACTOR_STRATEGY,
		log.Logger,
		nil)

	if err != nil {
		b.Fatalf("Failed to spawn the bar actor : %v", err)
	}

	log.Logger.Debug().Dict("bar", zerolog.Dict().Strs("path", bar.Address().Path).Str("id", bar.Address().Id)).Msg("")

	b.Run("foo <--HeartBeat--> bar - MessageChannelHandlers based", func(b *testing.B) {
		foo, err := system.Spawn("foo",
			func() actor2.MessageProcessor {
				return fooMessageProcessor(b.N, bar.Address())
			},
			actor2.RESTART_ACTOR_STRATEGY,
			log.Logger,
			nil)
		if err != nil {
			b.Fatalf("Failed to spawn the foo actor : %v", err)
		}
		log.Logger.Debug().Dict("foo", zerolog.Dict().Strs("path", foo.Address().Path).Str("id", foo.Address().Id)).Msg("")
		b.ResetTimer()
		foo.Tell(system.NewEnvelope(actor2.CHANNEL_INBOX, actor2.LIFECYCLE_MSG_STARTED, actor2.STARTED, nil, ""))
		foo.Wait()
	})

	b.Run("foo <--HeartBeat--> bar - switch based", func(b *testing.B) {
		foo, err := system.Spawn("foo",
			func() actor2.MessageProcessor {
				return &FooMessageProcessor{b.N, system.Address()}
			},
			actor2.RESTART_ACTOR_STRATEGY,
			log.Logger,
			nil)
		if err != nil {
			b.Fatalf("Failed to spawn the foo actor : %v", err)
		}
		log.Logger.Debug().Dict("foo", zerolog.Dict().Strs("path", foo.Address().Path).Str("id", foo.Address().Id)).Msg("")
		b.ResetTimer()
		foo.Tell(system.NewEnvelope(actor2.CHANNEL_INBOX, actor2.LIFECYCLE_MSG_STARTED, actor2.STARTED, nil, ""))
		foo.Wait()
	})

	b.Run("system.Tell(HeartbeatRequest)", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			system.Tell(system.NewEnvelope(actor2.CHANNEL_SYSTEM, actor2.SYS_MSG_HEARTBEAT_REQ, actor2.HEARTBEAT_REQ, nil, ""))
		}
	})

	b.Run("system.Send(HeartbeatRequest,system.Address())", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			system.Send(system.NewEnvelope(actor2.CHANNEL_SYSTEM, actor2.SYS_MSG_HEARTBEAT_REQ, actor2.HEARTBEAT_REQ, nil, ""), system.Address())
		}
	})

	b.Run("bar.Tell(HeartbeatRequest)", func(b *testing.B) {
		bar := system.LookupActor(&actor2.Address{Path: append(system.Address().Path, "bar")})
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			bar.Tell(system.NewEnvelope(actor2.CHANNEL_SYSTEM, actor2.SYS_MSG_HEARTBEAT_REQ, actor2.HEARTBEAT_REQ, nil, ""))
		}
	})

	b.Run("foo.Tell(HeartbeatRequest,foo.Address()) - switch based", func(b *testing.B) {
		//b.Skip("hanging on sending a message")
		foo, err := system.Spawn("foo2",
			func() actor2.MessageProcessor {
				return &FooMessageProcessor{b.N, system.Address()}
			},
			actor2.RESTART_ACTOR_STRATEGY,
			log.Logger,
			nil)
		if err != nil {
			b.Fatalf("Failed to spawn the foo actor : %v", err)
		}
		log.Logger.Debug().Dict("foo", zerolog.Dict().Strs("path", foo.Address().Path).Str("id", foo.Address().Id)).Msg("")
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			log.Logger.Debug().Int("b.N", b.N).Msg("system.Tell(HeartbeatRequest,foo.Address()) - switch based")
			msg := system.NewEnvelope(actor2.CHANNEL_SYSTEM, actor2.SYS_MSG_HEARTBEAT_REQ, actor2.HEARTBEAT_REQ, nil, "")
			//system.Send(msg, foo.Address())
			foo.Tell(msg)
		}
	})

	const chanBufSize = 128

	b.Run("foo <--HeartBeat--> bar - buffered chan", func(b *testing.B) {
		bar.Kill(nil)
		bar.Wait()

		bar, err = system.Spawn("bar",
			barMessageProcessor,
			actor2.RESTART_ACTOR_STRATEGY,
			log.Logger,
			map[actor2.Channel]*actor2.ChannelSettings{actor2.CHANNEL_SYSTEM: &actor2.ChannelSettings{BufSize: chanBufSize}})
		if err != nil {
			b.Fatalf("Failed to spawn the foo actor : %v", err)
		}
		log.Logger.Debug().Dict("bar", zerolog.Dict().Strs("path", bar.Address().Path).Str("id", bar.Address().Id)).Msg("")

		foo, err := system.Spawn("foo",
			func() actor2.MessageProcessor {
				return fooMessageProcessor(b.N, bar.Address())
			},
			actor2.RESTART_ACTOR_STRATEGY,
			log.Logger,
			map[actor2.Channel]*actor2.ChannelSettings{actor2.CHANNEL_INBOX: &actor2.ChannelSettings{BufSize: chanBufSize}})
		if err != nil {
			b.Fatalf("Failed to spawn the foo actor : %v", err)
		}
		log.Logger.Debug().Dict("foo", zerolog.Dict().Strs("path", foo.Address().Path).Str("id", foo.Address().Id)).Msg("")

		b.ResetTimer()
		foo.Tell(system.NewEnvelope(actor2.CHANNEL_INBOX, actor2.LIFECYCLE_MSG_STARTED, actor2.STARTED, nil, ""))
		foo.Wait()
	})

	b.Run("bar.Tell(HeartbeatRequest) - buffered chan", func(b *testing.B) {
		bar.Kill(nil)
		bar.Wait()

		bar, err = system.Spawn("bar",
			barMessageProcessor,
			actor2.RESTART_ACTOR_STRATEGY,
			log.Logger,
			map[actor2.Channel]*actor2.ChannelSettings{actor2.CHANNEL_SYSTEM: &actor2.ChannelSettings{BufSize: chanBufSize}})
		if err != nil {
			b.Fatalf("Failed to spawn the foo actor : %v", err)
		}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			bar.Tell(bar.NewEnvelope(actor2.CHANNEL_SYSTEM, actor2.SYS_MSG_HEARTBEAT_REQ, actor2.HEARTBEAT_REQ, nil, ""))
		}
	})

	b.Run("bar.Send(HeartbeatRequest) - buffered chan", func(b *testing.B) {
		bar.Kill(nil)
		bar.Wait()

		bar, err = system.Spawn("bar",
			barMessageProcessor,
			actor2.RESTART_ACTOR_STRATEGY,
			log.Logger,
			map[actor2.Channel]*actor2.ChannelSettings{actor2.CHANNEL_SYSTEM: &actor2.ChannelSettings{BufSize: chanBufSize}})
		if err != nil {
			b.Fatalf("Failed to spawn the foo actor : %v", err)
		}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			system.Send(bar.NewEnvelope(actor2.CHANNEL_SYSTEM, actor2.SYS_MSG_HEARTBEAT_REQ, actor2.HEARTBEAT_REQ, nil, ""), bar.Address())
		}
	})

}

type FooMessageProcessor struct {
	count   int
	address *actor2.Address
}

// ChannelNames returns the channels that are supported by this message processor
func (a *FooMessageProcessor) ChannelNames() []actor2.Channel {
	return []actor2.Channel{actor2.CHANNEL_INBOX, actor2.CHANNEL_LIFECYCLE}
}

// MessageTypes returns the message types that are supported for the specified channel
func (a *FooMessageProcessor) MessageTypes(channel actor2.Channel) []actor2.MessageType {
	switch channel {
	case actor2.CHANNEL_INBOX:
		return []actor2.MessageType{actor2.SYS_MSG_HEARTBEAT_RESP}
	case actor2.CHANNEL_LIFECYCLE:
		return []actor2.MessageType{actor2.LIFECYCLE_MSG_STARTED}
	case actor2.CHANNEL_SYSTEM:
		return []actor2.MessageType{actor2.SYS_MSG_HEARTBEAT_REQ}
	default:
		return nil
	}
}

// Handler returns a Receive function that will be used to handle messages on the specified channel for the
// specified message type
func (a *FooMessageProcessor) Handler(channel actor2.Channel, msgType actor2.MessageType) actor2.Receive {
	switch channel {
	case actor2.CHANNEL_INBOX:
		switch msgType {
		case actor2.SYS_MSG_HEARTBEAT_RESP:
			return a.handleHeartbeatResponse
		case actor2.LIFECYCLE_MSG_STARTED:
			return a.started
		default:
			return nil
		}
	case actor2.CHANNEL_SYSTEM:
		switch msgType {
		case actor2.SYS_MSG_HEARTBEAT_REQ:
			return actor2.HandleHearbeatRequest
		default:
			return nil
		}
	default:
		return nil
	}
}

func (a *FooMessageProcessor) started(ctx *actor2.MessageContext) error {
	// when the actor is started, send the bar actor a heartbeat request to initiate the communication
	var ref *actor2.Actor
	for ; ref == nil; ref = ctx.System().LookupActor(a.address) {
		time.Sleep(100 * time.Millisecond)
	}
	replyTo, _ := ctx.ChannelAddress(actor2.CHANNEL_INBOX)
	heartbeatReq := ctx.NewEnvelope(actor2.CHANNEL_SYSTEM, actor2.SYS_MSG_HEARTBEAT_REQ, actor2.HEARTBEAT_REQ, replyTo, "")
	ref.Tell(heartbeatReq)
	return nil
}

func (a *FooMessageProcessor) handleHeartbeatResponse(ctx *actor2.MessageContext) error {
	a.count--
	ctx.Logger().Debug().Int("count", a.count).Msg("")
	if a.count == 0 {
		ctx.Logger().Debug().Msg("foo is done")
		ctx.Kill(nil)
		return nil
	}

	replyTo, _ := ctx.ChannelAddress(actor2.CHANNEL_INBOX)
	heartbeatReq := ctx.NewEnvelope(actor2.CHANNEL_SYSTEM, actor2.SYS_MSG_HEARTBEAT_REQ, actor2.PING_REQ, replyTo, "")
	if err := ctx.Send(heartbeatReq, a.address); err != nil {
		return err
	}

	return nil
}

func (a *FooMessageProcessor) UnmarshalMessage(channel actor2.Channel, msgType actor2.MessageType, msg []byte) (*actor2.Envelope, error) {
	switch channel {
	case actor2.CHANNEL_INBOX:
		switch msgType {
		case actor2.SYS_MSG_HEARTBEAT_RESP:
			return actor2.UnmarshalHeartbeatResponse(msg)
		default:
			return nil, &actor2.ChannelMessageTypeNotSupportedError{channel, msgType}
		}
	default:
		return nil, &actor2.ChannelNotSupportedError{channel}
	}
}

// Stopped is invoked after the message processor is stopped. The message processor should perform any cleanup here.
func (a *FooMessageProcessor) Stopped() {}
