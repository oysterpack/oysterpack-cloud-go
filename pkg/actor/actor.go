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

	"errors"

	"sync"

	"fmt"

	"strings"

	"github.com/nats-io/nuid"
	"github.com/oysterpack/oysterpack.go/pkg/logging"
	"github.com/rs/zerolog"
	"gopkg.in/tomb.v2"
)

type Receive func(ctx *MessageContext) error

type MessageContext struct {
	*Actor
	Envelope *Envelope
}

// system channels
const (
	CHANNEL_SYSTEM     = Channel("0")
	CHANNEL_LIFECYCLE  = Channel("1")
	CHANNEL_SUPERVISOR = Channel("2")
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
	system *System

	path    []string
	id      string
	address *Address

	created time.Time

	parent   *Actor
	children map[string]*Actor

	messageProcessorFactory MessageProcessorFactory
	msgProcessorEngine      *MessageProcessorEngine
	channels                map[Channel]chan *Envelope
	channelBufSizes         map[Channel]int

	failures Failures

	tomb.Tomb
	lock sync.RWMutex

	uid    *nuid.NUID
	logger zerolog.Logger

	restarts        int
	lastRestartTime time.Time
}

func (a *Actor) Spawn(name string, messageProcessorFactory MessageProcessorFactory, channelBufSizes map[Channel]int, logger zerolog.Logger, supervisor SupervisorStrategy) (*Actor, error) {
	a.lock.Lock()
	defer a.lock.Unlock()

	if name != strings.TrimSpace(name) {
		return nil, fmt.Errorf("name cannot have whitespace passing %q", name)
	}

	if len(name) == 0 {
		return nil, errors.New("name cannot be blank")
	}

	if a.children[name] != nil {
		return nil, fmt.Errorf("Child already exists with the same name : %s", name)
	}

	if messageProcessorFactory == nil {
		return nil, errors.New("MessageProcessorFactory is required")
	}

	child := &Actor{
		system: a.system,
		path:   append(a.path, name),
		parent: a,

		messageProcessorFactory: messageProcessorFactory,
		logger:                  logger,
	}

	if err := child.start(); err != nil {
		return nil, err
	}

	a.children[name] = child

	// watch the child
	a.Go(func() error {
		for {
			msgProcessorEngine := child.messageProcessorEngine()
			select {
			case <-a.Dying():
				return nil
			case <-msgProcessorEngine.Dead():
				if err := msgProcessorEngine.Err(); err != nil {
					a.failures.failure(err)
					if err, ok := err.(*MessageProcessingError); ok {
						supervisor(child, err)
					} else {
						MESSAGE_PROCESSOR_FAILURE.Log(child.logger.Error()).Err(err).Str(logging.TYPE, fmt.Sprintf("%T", err)).Msg("")
						if err := child.restart(RESTART_ACTOR); err != nil {

						}
					}
				}
			case <-child.Dead():
				if err := child.Err(); err != nil {
					a.failures.failure(err)
					if err, ok := err.(*MessageProcessingError); ok {
						supervisor(child, err)
					} else {
						MESSAGE_PROCESSOR_FAILURE.Log(child.logger.Error()).Err(err).Str(logging.TYPE, fmt.Sprintf("%T", err)).Msg("")
						child.restart(RESTART_ACTOR)
					}
				} else {
					delete(a.children, name)
				}
				return nil
			}
		}
		return nil
	})

	return child, nil
}

func (a *Actor) messageProcessorEngine() *MessageProcessorEngine {
	a.lock.RLock()
	defer a.lock.RUnlock()
	return a.msgProcessorEngine
}

func (a *Actor) Message(channel Channel, msg Message) *Envelope {
	return NewEnvelope(a.UID, channel, msg, nil)
}

func (a *Actor) Request(channel Channel, msg Message, replyToChannel Channel) *Envelope {
	return NewEnvelope(a.UID, channel, msg, &ChannelAddress{Channel: replyToChannel, Address: a.address})
}

func (a *Actor) MessageContext(msg *Envelope) *MessageContext {
	return &MessageContext{a, msg}
}

func (a *Actor) messageProcessorEngineChannel(channel Channel) chan<- *MessageContext {
	a.lock.RLock()
	defer a.lock.RUnlock()
	return a.msgProcessorEngine.Channel(channel)
}

func (a *Actor) start() (err error) {
	a.lock.Lock()
	defer a.lock.Unlock()
	defer func() {
		if err != nil {
			a.Kill(err)
		}
	}()

	if err = a.init(); err != nil {
		return
	}
	if err = a.startMessageProcessorEngine(); err != nil {
		return
	}

	// when this actor is killed, tear down the actor hierarchy
	a.Go(func() error {
		<-a.Dying()
		a.lock.RLock()
		defer a.lock.RUnlock()
		for _, child := range a.children {
			child.Kill(nil)
		}

		for _, child := range a.children {
			child.Wait()
		}
		a.msgProcessorEngine.Kill(nil)
		a.msgProcessorEngine.Wait()
		return nil
	})
	return
}

func (a *Actor) init() error {
	a.uid = nuid.New()
	a.created = time.Now()
	a.id = a.uid.Next()
	a.address = &Address{a.path, a.id}

	if err := a.validate(); err != nil {
		return err
	}

	a.logger = a.logger.With().Strs(ACTOR_PATH, a.path).Logger()
	return nil
}

func (a *Actor) startMessageProcessorEngine() (err error) {
	if a.msgProcessorEngine, err = StartMessageProcessorEngine(a.messageProcessorFactory(), a.logger); err != nil {
		return
	}
	msgProcessorEngine := a.msgProcessorEngine
	a.channels = make(map[Channel]chan *Envelope)
	for _, channel := range msgProcessorEngine.ChannelNames() {
		c := make(chan *Envelope, a.channelBufSizes[channel])
		a.channels[channel] = c
		messageProcessorEngineChannel := a.messageProcessorEngineChannel(channel)

		// these goroutines will die when either the actor or the message processor engine dies
		a.Go(func() error {
			for {
				select {
				case msg := <-c:
					select {
					case messageProcessorEngineChannel <- a.MessageContext(msg):
					case <-a.Dying():
						return nil
					case <-msgProcessorEngine.Dead():
						return nil
					}
				case <-a.Dying():
					return nil
				case <-msgProcessorEngine.Dead():
					return nil
				}
			}
		})
	}
	a.Tell(a.Message(CHANNEL_LIFECYCLE, STARTED))
	ACTOR_STARTED.Log(a.logger.Info()).Msg("")

	// invoke MessageProcessor.Stopped() after it is dead to enable it to release resources and cleanup.
	a.Go(func() error {
		defer msgProcessorEngine.Stopped()
		select {
		case <-a.Dying():
		case <-msgProcessorEngine.Dead():
		}
		return nil
	})

	return
}

func (a *Actor) restart(mode RestartMode) (err error) {
	a.lock.Lock()
	defer a.lock.Unlock()
	defer func() {
		if err != nil {
			a.Kill(err)
		}
	}()

	a.msgProcessorEngine.Kill(nil)
	a.msgProcessorEngine.Wait()
	a.msgProcessorEngine.closeChannels()

	switch mode {
	case RESTART_ACTOR:
	case RESTART_ACTOR_HIERARCHY:
		for _, child := range a.children {
			child.restart(mode)
		}
	case RESTART_ACTOR_KILL_CHILDREN:
		for _, child := range a.children {
			child.Kill(nil)
		}
		for _, child := range a.children {
			child.Wait()
		}
	default:
		return fmt.Errorf("Unknown RestartMode : %d", mode)
	}

	if err = a.restartMessageProcessorEngine(); err != nil {
		return err
	}

	ACTOR_RESTARTED.Log(a.logger.Info()).Msg("")

	a.restarts++
	a.lastRestartTime = time.Now()

	return nil
}

func (a *Actor) restartMessageProcessorEngine() (err error) {
	if a.msgProcessorEngine, err = StartMessageProcessorEngine(a.messageProcessorFactory(), a.logger); err != nil {
		return
	}
	msgProcessorEngine := a.msgProcessorEngine
	for channel, c := range a.channels {
		messageProcessorEngineChannel := a.messageProcessorEngineChannel(channel)

		// these goroutines will die when either the actor or the message processor engine dies
		a.Go(func() error {
			for {
				select {
				case msg := <-c:
					select {
					case messageProcessorEngineChannel <- a.MessageContext(msg):
					case <-a.Dying():
						return nil
					case <-msgProcessorEngine.Dead():
						return nil
					}
				case <-a.Dying():
					return nil
				case <-msgProcessorEngine.Dead():
					return nil
				}
			}
		})
	}
	a.Tell(a.Message(CHANNEL_LIFECYCLE, STARTED))
	return
}

func (a *Actor) validate() error {
	if len(a.path) == 0 {
		return errors.New("Path is required")
	}

	if err := a.address.Validate(); err != nil {
		return err
	}

	if len(a.path) > 1 {
		if a.system == nil {
			return errors.New("An actor must belong to a system")
		}

		if a.parent == nil {
			return errors.New("The actor must have a parent")
		}
	} // else this is the root actor of the system

	if a.messageProcessorFactory == nil {
		return errors.New("MessageProcessorFactory is required")
	}

	return nil
}

func (a *Actor) System() *System {
	return a.system
}

func (a *Actor) Name() string {
	return a.path[len(a.path)-1]
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

// Err returns the death reason, or ErrStillAlive if the actor is not in a dying or dead state.
func (a *Actor) Err() error {
	err := a.Tomb.Err()
	if err == tomb.ErrStillAlive {
		return ErrStillAlive
	}
	return err
}

// Tell sends a message with fire and forget semantics.
func (a *Actor) Tell(msg *Envelope) error {
	c := a.channels[msg.Channel()]
	if c == nil {
		return &InvalidChannelError{msg.Channel()}
	}

	select {
	case c <- msg:
	default:
		a.Go(func() error {
			select {
			case <-a.Dying():
			case c <- msg:
			}
			return nil
		})
	}
	return nil
}

func (a *Actor) Channels() []Channel {
	channels := make([]Channel, len(a.channels))
	i := 0
	for channel := range a.channels {
		channels[i] = channel
		i++
	}
	return channels
}
