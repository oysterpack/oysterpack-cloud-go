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

	"fmt"

	"time"

	"strings"

	"github.com/nats-io/nuid"
	"github.com/oysterpack/oysterpack.go/pkg/actor"
	"github.com/rs/zerolog/log"
	"gopkg.in/tomb.v2"
)

func props(channelSize uint) *actor.Props {
	return &actor.Props{
		MessageProcessor: &actor.MessageProcessorProps{
			Factory: func() actor.MessageProcessor {
				return actor.MessageHandlers{
					actor.PING_REQUEST.MessageType(): actor.PING_REQUEST_HANDLER,
					Spawn_MessageType: &actor.MessageHandler{
						Receive: func(ctx *actor.MessageContext) error {
							spawn, ok := ctx.Message.Message().(*Spawn)
							if !ok {
								return fmt.Errorf("Unsupported message Go type : %T", ctx.Message.Message())
							}
							if spawn.Name == "" {
								for i := 0; i < spawn.N; i++ {
									if _, err := spawn.Props.NewChild(ctx.Actor, ctx.NextUID()); err != nil {
										return err
									}
								}
								if ctx.Message.ReplyTo() != nil {
									replyToActor, ok := ctx.System().Actor(ctx.Message.ReplyTo().Address)
									if ok {
										replyToActor.Tell(replyToActor.Envelope(actor.PING_REQUEST, nil, nil))
									}
								}
							} else {
								if _, err := spawn.Props.NewChild(ctx.Actor, spawn.Name); err != nil {
									return err
								}
							}
							return nil
						},
						Unmarshal: actor.Unmarshaller(&Spawn{}),
					},
				}
			},
			ErrorHandler: func(ctx *actor.MessageContext, err error) {
				ctx.Logger().Error().Err(err).Msg("RESUME")
			},
			ChannelSize: channelSize,
		},
		Logger: log.Logger,
	}
}

const Spawn_MessageType = actor.MessageType(0xe7fae189cec540ff)

type Spawn struct {
	actor.EmptyMessage
	*actor.Props
	Name string
	N    int
}

func (a *Spawn) MessageType() actor.MessageType {
	return Spawn_MessageType
}

func TestSystem(t *testing.T) {

	t.Run("Empty System", func(t *testing.T) {
		systemTomb := tomb.Tomb{}

		// Given a system with no actors
		system := actor.NewSystem()
		if system.ActorCount() != 0 {
			t.Error("The actor system should be empty")
		}

		// When root actors are retrieved
		// Then the channel is immediately closed, and no actors are returned
		if _, ok := <-system.RootActors(systemTomb); ok {
			t.Error("Nothing should have been returned")
		}

		// When no actor lives at the specified address
		// Then no actor is returned
		if actor, ok := system.Actor(&actor.Address{Path: "foo"}); ok || actor != nil {
			t.Error("No actor should have been returned")
		}

		// When invalid addresses are used
		// Then no actor is returned
		if actor, ok := system.Actor(&actor.Address{}); ok || actor != nil {
			t.Error("No actor should have been returned")
		}
		if actor, ok := system.Actor(actor.NewAddress("foo", "id")); ok || actor != nil {
			t.Error("No actor should have been returned")
		}

		// When root actors are killed
		// Then nothing happens because there are no root actors
		system.KillRootActors(systemTomb)
	})

	t.Run("Actor hierarchy", func(t *testing.T) {
		systemTomb := tomb.Tomb{}
		system := actor.NewSystem()
		defer system.KillRootActors(systemTomb)

		// Given an actor hierarchy :
		// 			a
		//    b-1      b-2
		//    c-1      c-2
		root, err := props(1).NewRootActor(system, "a")
		if err != nil {
			t.Fatal(err)
		}

		// When trying to create a new root actor using a path that already exists
		// Then creating the new root actor should fail
		_, err = props(1).NewRootActor(system, "a")
		switch err := err.(type) {
		case nil:
			t.Fatal("Creating a new root actor with a path that is already in use should fail")
		case actor.ActorAlreadyRegisteredError:
		default:
			t.Fatal("failed with a different error type : %T : %v", err, err)
		}

		if err = root.TellBlocking(root.NewEnvelope(&Spawn{Name: "b-1", Props: props(1)}, root.Address(), nil, nil)); err != nil {
			t.Fatal(err)
		}
		if err = root.TellBlocking(root.NewEnvelope(&Spawn{Name: "b-2", Props: props(1)}, root.Address(), nil, nil)); err != nil {
			t.Fatal(err)
		}

		for _, ok := root.Child("b-1"); !ok; {
			time.Sleep(time.Millisecond)
			_, ok = root.Child("b-1")
		}
		for _, ok := root.Child("b-2"); !ok; {
			time.Sleep(time.Millisecond)
			_, ok = root.Child("b-2")
		}

		actor_b1, _ := root.Child("b-1")
		if err = actor_b1.Tell(actor_b1.NewEnvelope(&Spawn{Name: "c-1", Props: props(1)}, root.Address(), nil, nil)); err != nil {
			t.Fatal(err)
		}

		actor_b2, _ := root.Child("b-2")
		if err = actor_b2.Tell(actor_b2.NewEnvelope(&Spawn{Name: "c-2", Props: props(1)}, root.Address(), nil, nil)); err != nil {
			t.Fatal(err)
		}

		for _, ok := actor_b1.Child("c-1"); !ok; {
			time.Sleep(time.Millisecond)
			_, ok = actor_b1.Child("c-1")
		}
		for _, ok := actor_b2.Child("c-2"); !ok; {
			time.Sleep(time.Millisecond)
			_, ok = actor_b2.Child("c-2")
		}

		actor_c1, _ := actor_b1.Child("c-1")

		actor_c2, _ := actor_b2.Child("c-2")
		// When each Actor is looked up
		// Then it is found and returned
		actors := []*actor.Actor{root, actor_b1, actor_b2, actor_c1, actor_c2}
		for _, a := range actors {
			b, ok := system.Actor(a.Address())
			if !ok {
				t.Errorf("Actor was not found: %v", a.Address())
			}
			if b.ID() != a.ID() {
				t.Error("wrong actor was returned")
			}

			b, ok = system.Actor(&actor.Address{Path: a.Address().Path})
			if !ok {
				t.Errorf("Actor was not found: %v", a.Address())
			}
			if b.ID() != a.ID() {
				t.Error("wrong actor was returned")
			}
		}

		// And each actor is linked with the right parents and children within the hierarchy
		children := root.Children()
		if len(children) != 2 {
			t.Error("root has the wrong number of children")
		}
		for _, child := range children {
			if child.Parent().ID() != root.ID() {
				t.Errorf("parent is wrong : %s -> %s", child.Path(), child.Parent().ID())
			}
			if !child.Created().After(root.Created()) {
				t.Error("the child was created after the parent : %v : %v", root.Created(), child.Created())
			}
			if child.Path() != strings.Join([]string{root.Name(), child.Name()}, "/") {
				t.Error("child path is wrong : %v : %v", child.Path(), child.Name())
			}
		}
		// And ChildreNames should return the proper names
		childrenNames := root.ChildrenNames()
		if len(childrenNames) != 2 {
			t.Error("root has the wrong number of children")
		}
		for _, name := range childrenNames {
			if child, ok := root.Child(name); !ok {
				t.Error("child was not found")
			} else if child.Name() != name {
				t.Errorf("wrong child was returned : %v : %v", name, child.Name())
			}
		}

	})

}

// go test -run=NONE -bench=. -benchmem -v -args -loglevel ERROR
// set loglevel to ERROR because the PingRequest handler logs an INFO level message when a PingRequest message is received
func BenchmarkActor(b *testing.B) {
	system := actor.NewSystem()
	systemTomb := tomb.Tomb{}
	defer system.KillRootActors(systemTomb)

	runCount := 0
	channelBlockedCount := 1
	var channelSize uint = 1

	// The goal of these benchmarks is to find the optimal channel size, which resulted in no back pressure.
	// in local benchmarks a channel size of 512 was the sweet spot
	//
	//BenchmarkActor/Tell_channelSize2-8               3000000               476 ns/op             144 B/op          2 allocs/op
	//BenchmarkActor/Tell_channelSize4-8               3000000               533 ns/op             144 B/op          2 allocs/op
	//BenchmarkActor/Tell_channelSize8-8               3000000               545 ns/op             144 B/op          2 allocs/op
	//BenchmarkActor/Tell_channelSize16-8              3000000               459 ns/op             144 B/op          2 allocs/op
	//BenchmarkActor/Tell_channelSize32-8              3000000               486 ns/op             144 B/op          2 allocs/op
	//BenchmarkActor/Tell_channelSize64-8              3000000               484 ns/op             144 B/op          2 allocs/op
	//BenchmarkActor/Tell_channelSize128-8             3000000               523 ns/op             144 B/op          2 allocs/op
	//BenchmarkActor/Tell_channelSize256-8             3000000               480 ns/op             144 B/op          2 allocs/op
	//BenchmarkActor/Tell_channelSize512-8             3000000               511 ns/op             144 B/op          2 allocs/op
	//--- BENCH: BenchmarkActor
	//system_test.go:109: channelSize = 2 : channelBlockedCount/runCount = 3497188/4010101 = 87.20947427508683
	//system_test.go:109: channelSize = 4 : channelBlockedCount/runCount = 3471457/4010101 = 86.56781961352095
	//system_test.go:109: channelSize = 8 : channelBlockedCount/runCount = 2853331/4010101 = 71.15359438577732
	//system_test.go:109: channelSize = 16 : channelBlockedCount/runCount = 2407665/4010101 = 60.040008967355185
	//system_test.go:109: channelSize = 32 : channelBlockedCount/runCount = 1892764/4010101 = 47.19990843123403
	//system_test.go:109: channelSize = 64 : channelBlockedCount/runCount = 1359998/4010101 = 33.914307893990696
	//system_test.go:109: channelSize = 128 : channelBlockedCount/runCount = 276657/4010101 = 6.899003291937037
	//system_test.go:109: channelSize = 256 : channelBlockedCount/runCount = 450/4010101 = 0.011221662496780006
	//system_test.go:109: channelSize = 512 : channelBlockedCount/runCount = 0/4010101 = 0
	//
	for channelBlockedCount > 0 {
		runCount = 0
		channelBlockedCount = 0
		channelSize = channelSize * 2
		b.Run(fmt.Sprint("Tell channelSize", channelSize), ActorTellBenchmark(system, channelSize, &runCount, &channelBlockedCount))
		b.Logf("channelSize = %d : channelBlockedCount/runCount = %d/%d = %v", channelSize, channelBlockedCount, runCount, float64(channelBlockedCount)/float64(runCount)*100)
	}

	channelBlockedCount = 1
	channelSize = 1
	for channelBlockedCount > 0 {
		runCount = 0
		channelBlockedCount = 0
		channelSize = channelSize * 2
		b.Run(fmt.Sprint("Tell Blocking channelSize", channelSize), ActorTellBlockingBenchmark(system, 1, &runCount, &channelBlockedCount))
		b.Logf("channelSize = %d : channelBlockedCount/runCount = %d/%d = %v", channelSize, channelBlockedCount, runCount, float64(channelBlockedCount)/float64(runCount)*100)
	}

	// #1 starting Actor MessageProcessor goroutine directly via go
	// BenchmarkActor/Spawn-8            200000             10219 ns/op            6789 B/op         24 allocs/op
	//
	// #2 handing off the Actor MessageProcessor goroutine to the gofuncs channel to be started by a single goroutine
	// BenchmarkActor/Spawn-8            200000              8824 ns/op            6804 B/op         25 allocs/op
	//
	// #2 showed a significant ~15% improvement
	b.Run("Props.NewChild", PropsNewChildBenchmark(system))

	time.Sleep(time.Millisecond * 500)
	if inboxes, ok := system.Actor(&actor.Address{Path: "inboxes"}); ok {
		b.Logf("inbox count : %d", inboxes.ChildrenCount())
		inboxes.Kill()
		<-inboxes.Dead()
		b.Logf("after dead : inbox count : %d", inboxes.ChildrenCount())
	}

}

func ActorTellBenchmark(system *actor.System, channelSize uint, runCount *int, channelBlockedCount *int) func(b *testing.B) {
	return func(b *testing.B) {
		if foo, ok := system.Actor(&actor.Address{Path: "foo"}); ok {
			foo.Kill()
			<-foo.Dead()
		}

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
				ChannelSize: channelSize,
			},
			Logger: log.Logger,
		}
		foo, err := fooProps.NewRootActor(system, "foo")
		if err != nil {
			b.Fatal(err)
		}
		if !foo.Alive() {
			b.Fatal("foo is not alive")
		}

		*runCount += b.N
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			if err := foo.Tell(foo.NewEnvelope(actor.PING_REQUEST, foo.Address(), nil, nil)); err != nil {
				if err == actor.ErrChannelBlocked {
					*channelBlockedCount++
				}
			}
		}
	}
}

func ActorTellBlockingBenchmark(system *actor.System, channelSize uint, runCount *int, channelBlockedCount *int) func(b *testing.B) {
	return func(b *testing.B) {
		if foo, ok := system.Actor(&actor.Address{Path: "foo"}); ok {
			foo.Kill()
			<-foo.Dead()
		}

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
				ChannelSize: channelSize,
			},
			Logger: log.Logger,
		}
		foo, err := fooProps.NewRootActor(system, "foo")
		if err != nil {
			b.Fatal(err)
		}
		if !foo.Alive() {
			b.Fatal("foo is not alive")
		}

		*runCount += b.N
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			if err := foo.TellBlocking(foo.NewEnvelope(actor.PING_REQUEST, foo.Address(), nil, nil)); err != nil {
				if err == actor.ErrChannelBlocked {
					*channelBlockedCount++
				}
			}
		}
	}
}

func PropsNewChildBenchmark(system *actor.System) func(b *testing.B) {
	return func(b *testing.B) {
		if inboxes, ok := system.Actor(&actor.Address{Path: "inboxes"}); ok {
			b.Logf("inbox count : %d", inboxes.ChildrenCount())
			inboxes.Kill()
			<-inboxes.Dead()
			b.Logf("after dead : inbox count : %d", inboxes.ChildrenCount())
		}
		inboxes, err := props(1).NewRootActor(system, "inboxes")
		if err != nil {
			b.Fatal(err)
		}
		msgs := make([]*actor.Envelope, b.N)
		childProps := props(1)
		for i := 0; i < b.N; i++ {
			msgs[i] = inboxes.Envelope(&Spawn{Name: nuid.Next(), Props: childProps}, nil, nil)
		}
		b.Logf("b.N = %d", b.N)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			//if err := inboxes.TellBlocking(inboxes.Envelope(&Spawn{Name: nuid.Next(), Props: childProps}, nil, nil)); err != nil {
			if err := inboxes.TellBlocking(msgs[i]); err != nil {
				b.Fatal(err)
			}
		}
	}
}

//func PropsNewChildBenchmark(system *actor.System) func(b *testing.B) {
//	inbox := make(chan *actor.Envelope)
//	inboxProps := &actor.Props{
//		MessageProcessor: actor.MessageProcessorProps{
//			Factory: func() actor.MessageProcessor {
//				return actor.MessageHandlers{
//					actor.PING_REQUEST.MessageType(): actor.MessageHandler{actor.InboxMessageHandler(inbox), actor.Unmarshaller(actor.PING_REQUEST)},
//				}
//			},
//			ErrorHandler: func(ctx *actor.MessageContext, err error) {
//				ctx.Logger().Error().Err(err).Msg("RESUME")
//			},
//			ChannelSize: 512,
//		},
//		Logger: log.Logger,
//	}
//
//	inboxActor, _ := inboxProps.NewRootActor(system, nuid.Next())
//
//	return func(b *testing.B) {
//		if inboxes, ok := system.Actor(actor.Address{Path: "inboxes"}); ok {
//			b.Logf("inbox count : %d", inboxes.ChildrenCount())
//			inboxes.Kill()
//			<-inboxes.Dead()
//			b.Logf("after dead : inbox count : %d", inboxes.ChildrenCount())
//		}
//		inboxes, err := props(1).NewRootActor(system, "inboxes")
//		if err != nil {
//			b.Fatal(err)
//		}
//		b.Logf("b.N = %d", b.N)
//		b.ResetTimer()
//		if err := inboxes.TellBlocking(inboxes.Envelope(&Spawn{N: b.N, Props: props(1)}, &actor.ReplyTo{Address: inboxActor.Address(), MessageType: actor.PING_REQUEST.MessageType()}, nil)); err != nil {
//			b.Fatal(err)
//		}
//		<-inbox
//	}
//}
