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
	"github.com/oysterpack/oysterpack.go/pkg/commons"
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
	CHANNEL_LIFECYCLE = Channel("1")
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
	children map[string]*Actor // keys are the actor names

	messageProcessorFactory MessageProcessorFactory
	msgProcessorEngine      *MessageProcessorEngine
	channels                map[Channel]chan *Envelope
	channelSettings         map[Channel]*ChannelSettings

	failures Failures

	tomb.Tomb
	lock sync.RWMutex

	uid    *nuid.NUID
	logger zerolog.Logger

	restarts        int
	lastRestartTime time.Time
}

// ChannelSettings is actor message channel related configurations
type ChannelSettings struct {
	Channel
	BufSize int
	// used to create new empty Envelope(s)
	ChannelMessageFactory ChannelMessageFactory
	MessageValidator      func(msg Message) error
}

func (a *Actor) Spawn(name string, messageProcessorFactory MessageProcessorFactory, channelSettings []*ChannelSettings, logger zerolog.Logger, supervise SupervisorStrategy) (*Actor, error) {
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

	if len(channelSettings) == 0 {
		return nil, errors.New("At least 1 channel is required")
	}

	chanSettingsMap := make(map[Channel]*ChannelSettings, len(channelSettings))
	for _, settings := range channelSettings {
		if _, exists := chanSettingsMap[settings.Channel]; exists {
			return nil, fmt.Errorf("Duplcate channel name : %s", settings.Channel)
		}
		chanSettingsMap[settings.Channel] = settings
		if err := settings.Channel.Validate(); err != nil {
			return nil, err
		}
	}

	child := &Actor{
		system: a.system,
		path:   append(a.path, name),
		parent: a,

		messageProcessorFactory: messageProcessorFactory,
		channelSettings:         chanSettingsMap,
		logger:                  logger,
	}

	if err := child.start(); err != nil {
		return nil, err
	}

	a.children[name] = child
	a.system.registerActor(child)

	// watch the child
	go func() {
		for {
			// NOTE: the child may have been restarted. Thus, we always want to get the current MessageProcessorEngine.
			// MessageProcessorEngine access is protected by a RWMutex to enable safe concurrent access.
			msgProcessorEngine := child.messageProcessorEngine()
			select {
			case <-a.Dead():
				LOG_EVENT_DEAD.Log(child.logger.Info()).Msg("")
				return
			case <-msgProcessorEngine.Dead():
				if err := msgProcessorEngine.Err(); err != nil {
					a.failures.failure(err)
					supervise(child, err)
				}
			case <-child.Dead():
				a.deleteChild(name)
				a.system.unregisterActor(child.path, child.id)
			}
		}
	}()

	return child, nil
}

func (a *Actor) deleteChild(name string) {
	a.lock.Lock()
	defer a.lock.Unlock()
	delete(a.children, name)
}

func (a *Actor) messageProcessorEngine() *MessageProcessorEngine {
	a.lock.RLock()
	defer a.lock.RUnlock()
	return a.msgProcessorEngine
}

func (a *Actor) UnmarshalEnvelope(channel Channel, msgType MessageType, msg []byte) (*Envelope, error) {
	if settings := a.channelSettings[channel]; settings != nil {
		newMessage := settings.ChannelMessageFactory[msgType]
		if newMessage == nil {
			return nil, &ChannelMessageTypeNotSupportedError{channel, msgType}
		}
		envelope := &Envelope{message: newMessage()}
		if err := envelope.UnmarshalBinary(msg); err != nil {
			return nil, err
		}
		return envelope, nil
	}
	return nil, &ChannelNotSupportedError{channel}
}

// MessageEnvelope wraps the message in an envelope.
// Panics if channel is unknown for this actor.
func (a *Actor) MessageEnvelope(channel Channel, msg Message) *Envelope {
	if _, exists := a.channelSettings[channel]; !exists {
		LOG_EVENT_UNKNOWN_CHANNEL.Log(a.logger.Panic()).Str(LOG_FIELD_CHANNEL, channel.String()).Msg("")
	}
	return NewEnvelope(a.UID, channel, msg, nil)
}

// RequestEnvelope wraps the message request in an envelope
// Panics if channel is unknown for this actor.
func (a *Actor) RequestEnvelope(channel Channel, msg Message, replyToChannel Channel) *Envelope {
	if _, exists := a.channelSettings[channel]; !exists {
		LOG_EVENT_UNKNOWN_CHANNEL.Log(a.logger.Panic()).Str(LOG_FIELD_CHANNEL, channel.String()).Msg("")
	}
	return NewEnvelope(a.UID, channel, msg, &ChannelAddress{Channel: replyToChannel, Address: a.address})
}

func (a *Actor) messageContext(msg *Envelope) *MessageContext {
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
		LOG_EVENT_DYING.Log(a.logger.Info()).Msg("")
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

	a.logger = a.logger.With().Strs(LOG_FIELD_ACTOR_PATH, a.path).Logger()
	return nil
}

func (a *Actor) startMessageProcessorEngine() (err error) {
	if a.msgProcessorEngine, err = StartMessageProcessorEngine(a.messageProcessorFactory(), a.logger); err != nil {
		return
	}
	msgProcessorEngine := a.msgProcessorEngine
	a.channels = make(map[Channel]chan *Envelope)
	for _, channel := range msgProcessorEngine.ChannelNames() {
		c := make(chan *Envelope, a.channelSettings[channel].BufSize)
		a.channels[channel] = c
		messageProcessorEngineChannel := a.messageProcessorEngineChannel(channel)

		// These goroutines will die when either the actor or the message processor engine dies.
		// When the message processor engine is dead, the parent actor is notified (via the tomb). If the message processor
		// engine died with an error, then based on the supervisor strategy for this actor, the actor may be restarted.
		a.Go(func() error {
			for {
				select {
				case msg := <-c:
					select {
					case messageProcessorEngineChannel <- a.messageContext(msg):
					case <-a.Dying():
						return nil
					case <-msgProcessorEngine.Dying():
						return nil
					}
				case <-a.Dying():
					return nil
				case <-msgProcessorEngine.Dying():
					return nil
				}
			}
		})
	}
	a.Tell(a.MessageEnvelope(CHANNEL_LIFECYCLE, STARTED))
	LOG_EVENT_STARTED.Log(a.logger.Info()).Msg("")
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

	LOG_EVENT_RESTARTED.Log(a.logger.Info()).Msg("")

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

		// these goroutines will die when either the actor or the message processor engine is dying
		a.Go(func() error {
			for {
				select {
				case msg := <-c:
					select {
					case messageProcessorEngineChannel <- a.messageContext(msg):
					case <-a.Dying():
						return nil
					case <-msgProcessorEngine.Dying():
						return nil
					}
				case <-a.Dying():
					return nil
				case <-msgProcessorEngine.Dying():
					return nil
				}
			}
		})
	}
	a.Tell(a.MessageEnvelope(CHANNEL_LIFECYCLE, STARTED))
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

// Path is where the Actor lives
// NOTE: the first element of the path corresponds to the actor system name
func (a *Actor) Path() []string {
	return a.path
}

func (a *Actor) Id() string {
	return a.id
}

func (a *Actor) Address() *Address {
	return a.address
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
//
// Types of errors :
//  - *ChannelNotSupportedError
func (a *Actor) Tell(msg *Envelope) (err error) {
	c := a.channels[msg.Channel()]
	if c == nil {
		return &ChannelNotSupportedError{msg.Channel()}
	}

	select {
	case <-a.Dying():
		return ErrNotAlive
	case c <- msg:
	default:
		a.Go(func() error {
			defer commons.IgnorePanic()
			select {
			case <-a.Dying():
			case c <- msg:
			}
			return nil
		})
	}
	return nil
}

// Send a message to the specified address.
//
// If sending via the address path, the path is expected to be relative to this actor. The message will be sent to any
// actor instance that lives at that address.
//
// If the actor id is specified in the address, then the message will only be delivered if the same actor instance is still
// living with that id. In other words, if an actor is living at the specified address, but has a different id, then the message
// will not be delivered to that actor because it was intended for another actor instance.
func (a *Actor) Send(msg *Envelope, address *Address) error {
	actor := a.system.LookupActorRef(address)
	if actor == nil {
		return &ActorNotFoundError{address}
	}

	return actor.Tell(msg)
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
