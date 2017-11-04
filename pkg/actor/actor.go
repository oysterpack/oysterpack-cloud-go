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
	"github.com/rs/zerolog"
	"gopkg.in/tomb.v2"
)

// standard channels
const (
	CHANNEL_SYSTEM    = Channel("0")
	CHANNEL_LIFECYCLE = Channel("1")
	CHANNEL_INBOX     = Channel("2")
)

type Receive func(ctx *MessageContext) error

type MessageContext struct {
	*Actor
	Message *Envelope
}

// Actor represents an actor in the Actor Model. Actors are objects which encapsulate state and behavior, they communicate exclusively by exchanging messages.
//
// The actors model provide a high level abstraction for writing concurrent and distributed systems. This approach simplifies
// the burden imposed on engineers, such as explicit locks and concurrent access to shared state, as actors receive messages synchronously.
//
// Actors communicate via messages over channels.
// All messages must implement the Message interface, and thus support binary marshalling which in turn supports distributed messaging.
// Actors are defined by an actor path. The actor path must be unique within a local process.
//
// An actor sets up a message channel pipeline :
//
//										 |   |   |   |
//										 V   V   V   V
//       Actor channels  	            [ ]	[ ]	[ ]	[ ]		- may be buffered
//										 |   |   |   |
//										 V   V   V   V
//   Actor channel goroutines			 o   o   o   o		- fan in messages from the Actor channels into the MessageProcessorEngine channel
//										 |___|___|___|
//											   |
//											   V
//   MessageProcessorEngine channel			  [ ]
//											   |
//											   V
//	 MessageProcessor goroutine				   o
//
//   MessageProcessor watcher				   o			- invokes MessageProcessor.Stopped() after MessageProcessor is dead
//
//   Actor watcher							   o			- if Actor MessageProcessor engine dies with an error, then executes supervision strategy
//															- when Actor is killed, it deletes itself from its parent and shuts down the actor hierarchy
// 	where
//		[ ] = go channel
//		 o  = goroutine
//
// For each MessageProcessor Channel, the actor will create a channel. Each of the actor channels can be buffered.
// Messages arriving on the incoming actor channels are fanned into the core MessageProcessorEngine channel.
//
// An actor will have 3 core goroutines, plus 1 for each message channel:
//
// 	A		 B	   1	 2     N      3
//  |-start->|
//  |		 |-go->|
//	|  		 |-----|-go->|
//	|		 |     |     |
//  |--------|-----|-----|-go->|
//  |--------|-----|-----|-----|--go->|
//
//	where
// 		A = Actor
//		B = MessageProcessorEngine
//		1 = MessageProcessorEngine watcher
//		2 = MessageProcessor
//		N = N message channels, i.e., 1 goroutine per message channel supported by the actor
//		3 = Actor watcher

type Actor struct {
	// immutable
	system *System

	path    []string
	id      string
	address *Address

	created time.Time

	parent *Actor

	childrenLock sync.RWMutex
	children     map[string]*Actor

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

	supervisor SupervisorStrategy
}

func (a *Actor) putChild(actor *Actor) {
	if actor == nil {
		return
	}
	a.childrenLock.Lock()
	defer a.childrenLock.Unlock()
	if a.children == nil {
		a.children = make(map[string]*Actor)
	}
	a.children[actor.Name()] = actor
}

func (a *Actor) getChild(name string) *Actor {
	a.childrenLock.RLock()
	defer a.childrenLock.RUnlock()
	return a.children[name]
}

func (a *Actor) deleteChild(name string) {
	a.childrenLock.Lock()
	defer a.childrenLock.Unlock()
	if a.children != nil {
		delete(a.children, name)
	}
}

func (a *Actor) getChildren() []*Actor {
	a.childrenLock.RLock()
	defer a.childrenLock.RUnlock()
	temp := make([]*Actor, len(a.children))
	i := 0
	for _, child := range a.children {
		temp[i] = child
	}
	return temp
}

//func (a *actorMap) addressPathKey(path []string) string { return strings.Join(path, "") }

// ChannelSettings is actor message channel related configurations
type ChannelSettings struct {
	Channel
	BufSize int
}

// Spawn creates a new child actor
//
// logger is augmented
func (a *Actor) Spawn(name string, messageProcessorFactory MessageProcessorFactory, supervisor SupervisorStrategy, logger zerolog.Logger, channelSettings ...*ChannelSettings) (*Actor, error) {
	if err := a.validateSpawnParams(name, messageProcessorFactory, channelSettings, supervisor, logger); err != nil {
		return nil, err
	}

	a.lock.Lock()
	defer a.lock.Unlock()

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
		supervisor:              supervisor,
	}

	if err := child.start(); err != nil {
		return nil, err
	}

	a.putChild(child)

	return child, nil
}

func (a *Actor) validateSpawnParams(name string, messageProcessorFactory MessageProcessorFactory, channelSettings []*ChannelSettings, supervisor SupervisorStrategy, logger zerolog.Logger) error {
	if name != strings.TrimSpace(name) {
		return fmt.Errorf("name cannot have whitespace passing %q", name)
	}

	if len(name) == 0 {
		return errors.New("name cannot be blank")
	}

	// when an actor is dying, it is removed as a child in an async fashion - thus, check that the child is actually alive
	if child := a.getChild(name); child != nil && child.Alive() {
		return fmt.Errorf("Child already exists with the same name : %s", name)
	}

	if messageProcessorFactory == nil {
		return errors.New("MessageProcessorFactory is required")
	}

	return nil
}

func (a *Actor) messageProcessorEngine() *MessageProcessorEngine {
	a.lock.RLock()
	defer a.lock.RUnlock()
	return a.msgProcessorEngine
}

func (a *Actor) UnmarshalEnvelope(channel Channel, msgType MessageType, msg []byte) (*Envelope, error) {
	return a.msgProcessorEngine.UnmarshalMessage(channel, msgType, msg)
}

// Envelope returns a new Envelope with the specified settings.
// The envelope will be assigned a uid and the message creation timestamp will be set to now.
func (a *Actor) NewEnvelope(channel Channel, msgType MessageType, msg Message, replyTo *ChannelAddress, correlationId string) *Envelope {
	return NewEnvelope(a.UID, channel, msgType, msg, replyTo, correlationId)
}

func (a *Actor) messageContext(msg *Envelope) *MessageContext {
	return &MessageContext{a, msg}
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

	a.watch()
	return
}

// when this actor is killed, tear down the actor hierarchy and kill the message processor engine
func (a *Actor) watch() {
	a.Go(func() error {

		for {
			// NOTE: the child may have been restarted. Thus, we always want to get the current MessageProcessorEngine.
			// MessageProcessorEngine access is protected by a RWMutex to enable safe concurrent access.
			messageProcessorEngine := a.messageProcessorEngine()

			select {
			case <-a.Dying():
				a.stop()
				return nil
			case <-messageProcessorEngine.Dead():
				select {
				case <-a.Dying():
					a.stop()
					return nil
				default:
					// if the MessageProcessorEngine died with an error, then delegate to the supervisor
					if err := messageProcessorEngine.Err(); err != nil {
						LOG_EVENT_DEATH_ERR.Log(a.logger.Error()).Err(err).Msg("")
						a.failures.failure(err)
						a.supervisor(a, err)
					}
				}
			}
		}
	})
}

func (a *Actor) stop() {
	LOG_EVENT_DYING.Log(a.logger.Debug()).Msg("")
	a.lock.RLock()
	defer a.lock.RUnlock()

	if a.parent != nil {
		a.parent.deleteChild(a.Name())
	}

	children := a.getChildren()

	for _, child := range children {
		LOG_EVENT_KILL.Log(a.logger.Debug()).Str(LOG_FIELD_CHILD, child.Name()).Msg("")
		child.Kill(nil)
	}

	for _, child := range children {
		if err := child.Wait(); err != nil {
			LOG_EVENT_DEATH_ERR.Log(a.logger.Debug()).Str(LOG_FIELD_CHILD, child.Name()).Err(err).Msg("")
		} else {
			LOG_EVENT_DEAD.Log(a.logger.Debug()).Str(LOG_FIELD_CHILD, child.Name()).Msg("")
		}
	}
	a.msgProcessorEngine.Kill(nil)
}

func (a *Actor) init() error {
	a.uid = nuid.New()
	a.created = time.Now()
	a.id = a.uid.Next()
	a.address = &Address{a.path, a.id}

	a.logger = a.logger.With().
		Dict(LOG_FIELD_ACTOR, zerolog.Dict().
			Strs(LOG_FIELD_ACTOR_PATH, a.path).
			Str(LOG_FIELD_ACTOR_ID, a.id)).
		Logger()
	return nil
}

func (a *Actor) startMessageProcessorEngine() (err error) {
	if a.msgProcessorEngine, err = StartMessageProcessorEngine(a.messageProcessorFactory(), a.logger); err != nil {
		return
	}
	msgProcessorEngine := a.msgProcessorEngine

	if err := a.checkChannelSettingsAlignWithMessageProcessorChannels(); err != nil {
		a.msgProcessorEngine.Kill(nil)
		return err
	}

	a.channels = make(map[Channel]chan *Envelope)
	for _, channel := range msgProcessorEngine.ChannelNames() {
		channelSettings := a.channelSettings[channel]
		bufSize := 0
		if channelSettings != nil {
			bufSize = channelSettings.BufSize
		}
		c := make(chan *Envelope, bufSize)
		a.channels[channel] = c
		messageProcessorEngineChannel := msgProcessorEngine.Channel()

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

	a.Tell(a.NewEnvelope(CHANNEL_LIFECYCLE, LIFECYCLE_MSG_STARTED, STARTED, nil, ""))

	LOG_EVENT_STARTED.Log(a.logger.Debug()).Msg("")
	return
}

func (a *Actor) checkChannelSettingsAlignWithMessageProcessorChannels() error {
	if len(a.channelSettings) == 0 {
		return nil
	}

	channels := make(map[Channel]bool, len(a.msgProcessorEngine.ChannelNames()))
	for _, channel := range a.msgProcessorEngine.ChannelNames() {
		channels[channel] = true
	}

	for channel := range a.channelSettings {
		if !channels[channel] {
			return fmt.Errorf("Extra ChannelSetting was found that is not supported by the MessageProcessor : %s", channel)
		}
	}

	return nil
}

func (a *Actor) HasChannel(channel Channel) bool {
	return a.channels[channel] != nil
}

func (a *Actor) ChannelAddress(channel Channel) (*ChannelAddress, bool) {
	if a.HasChannel(channel) {
		return &ChannelAddress{Channel: channel, Address: a.address}, true
	}
	return nil, false
}

func (a *Actor) restart(mode RestartMode) (err error) {
	a.lock.Lock()
	defer a.lock.Unlock()

	if !a.Alive() {
		a.Tomb = tomb.Tomb{}
	}

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
		for _, child := range a.getChildren() {
			child.restart(mode)
		}
	case RESTART_ACTOR_KILL_CHILDREN:
		children := a.getChildren()
		for _, child := range children {
			child.Kill(nil)
		}
		for _, child := range children {
			child.Wait()
		}
	default:
		return fmt.Errorf("Unknown RestartMode : %d", mode)
	}

	if err = a.restartMessageProcessorEngine(); err != nil {
		return err
	}

	LOG_EVENT_RESTARTED.Log(a.logger.Info()).Str(LOG_FIELD_MODE, mode.String()).Msg("")

	a.restarts++
	a.lastRestartTime = time.Now()

	return nil
}

func (a *Actor) restartMessageProcessorEngine() (err error) {
	if a.msgProcessorEngine, err = StartMessageProcessorEngine(a.messageProcessorFactory(), a.logger); err != nil {
		return
	}
	msgProcessorEngine := a.msgProcessorEngine
	for _, c := range a.channels {
		messageProcessorEngineChannel := msgProcessorEngine.Channel()

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
	a.Tell(a.NewEnvelope(CHANNEL_LIFECYCLE, LIFECYCLE_MSG_STARTED, STARTED, nil, ""))
	return
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
func (a *Actor) Tell(msg *Envelope) error {
	c := a.channels[msg.Channel()]
	if c == nil {
		return &ChannelNotSupportedError{msg.Channel()}
	}

	select {
	case <-a.Dying():
		return ErrNotAlive
	case c <- msg:
	}
	return nil
}

func (a *Actor) MessageChannel(channel Channel) func(msg *Envelope) error {
	c := a.channels[channel]
	if c == nil {
		return nil
	}

	return func(msg *Envelope) error {
		select {
		case <-a.Dying():
			return ErrNotAlive
		case c <- msg:
			return nil
		}
	}
}

// Send a message to the specified address.
//
// If sending via the address path, the path is expected to be relative to this actor. The message will be sent to any
// actor instance that lives at that address.
//
// If the actor id is specified in the address, then the message will only be delivered if the same actor instance is still
// living with that id. In other words, if an actor is living at the specified address, but has a different id, then the message
// will not be delivered to that actor because it was intended for another actor instance.
func (a *Actor) Send(msg *Envelope, to *Address) error {
	actor := a.system.LookupActor(to)
	if actor == nil {
		return &ActorNotFoundError{to}
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

// ChannelMetrics returns channel metrics
func (a *Actor) ChannelMetrics() map[Channel]*ChannelMetrics {
	stats := map[Channel]*ChannelMetrics{}
	for channel, c := range a.channels {
		stats[channel] = &ChannelMetrics{
			Capacity: cap(c),
			Len:      len(c),
		}
	}
	return stats
}

func (a *Actor) Logger() zerolog.Logger {
	return a.logger
}

// path should be relative to the actor
func (a *Actor) lookupActor(path []string, id string) *Actor {
	switch len(path) {
	case 0:
		return a
	case 1:
		child := a.getChild(path[0])
		if child != nil && (child.id == id || id == "") {
			return child
		}
		return nil
		return child
	default:
		child := a.getChild(path[0])
		if child == nil {
			return nil
		}
		return child.lookupActor(path[1:], id)
	}
}
