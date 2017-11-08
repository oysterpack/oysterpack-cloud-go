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

	"sync"

	"strings"

	"github.com/nats-io/nuid"
	"github.com/rs/zerolog"
	"gopkg.in/tomb.v2"
)

type Actor struct {
	system *System

	name string
	path string
	id   string

	created time.Time

	parent *Actor

	childrenLock sync.RWMutex
	children     map[string]*Actor

	messageProcessorFactory MessageProcessorFactory
	messageProcessor        MessageProcessor
	errorHandler            MessageProcessorErrorHandler
	messageProcessorChan    chan *Envelope

	tomb.Tomb

	actorChan chan interface{}

	// used to :
	// 	- generate unique ids for children
	//	- generate unique message envelope ids
	uidGenerator *nuid.NUID
	logger       zerolog.Logger
}

func (a *Actor) System() *System {
	return a.system
}

func (a *Actor) Created() time.Time {
	return a.created
}

// Name returns the actor name. The name must be unique with all other siblings.
func (a *Actor) Name() string {
	return a.name
}

func (a *Actor) Path() string {
	return a.path
}

func (a *Actor) ID() string {
	return a.id
}

func (a *Actor) Address() Address {
	return Address{path: a.path, id: &a.id}
}

func (a *Actor) Parent() *Actor {
	return a.parent
}

func (a *Actor) Logger() zerolog.Logger {
	return a.logger
}

func (a *Actor) Child(name string) (child *Actor, ok bool) {
	a.childrenLock.RLock()
	child, ok = a.children[name]
	a.childrenLock.RUnlock()
	return
}

func (a *Actor) ChildrenNames() []string {
	a.childrenLock.RLock()
	children := make([]string, len(a.children))
	i := 0
	for _, child := range a.children {
		children[i] = child.name
		i++
	}
	a.childrenLock.RUnlock()
	return children
}

func (a *Actor) Children() []*Actor {
	a.childrenLock.RLock()
	children := make([]*Actor, len(a.children))
	i := 0
	for _, child := range a.children {
		children[i] = child
		i++
	}
	a.childrenLock.RUnlock()
	return children
}

func (a *Actor) putChild(child *Actor) bool {
	a.childrenLock.Lock()
	_, ok := a.children[child.name]
	if ok {
		return false
	}
	a.children[child.name] = child
	a.childrenLock.Unlock()
	return true
}

func (a *Actor) deleteChild(name string) {
	a.childrenLock.Lock()
	delete(a.children, name)
	a.childrenLock.Unlock()
}

func (a *Actor) clearChildren() {
	a.childrenLock.Lock()
	a.children = make(map[string]*Actor)
	a.childrenLock.Unlock()
}

// Err returns the death reason, or ErrStillAlive if the actor is not in a dying or dead state.
func (a *Actor) Err() error {
	err := a.Tomb.Err()
	if err == tomb.ErrStillAlive {
		return ErrStillAlive
	}
	return err
}

func (a *Actor) Spawn(name string, messageProcessorFactory MessageProcessorFactory, errorHandler MessageProcessorErrorHandler, logger zerolog.Logger) (*Actor, error) {
	if name == "" {
		return nil, ErrActorNameBlank
	}
	if strings.Contains(name, "/") {
		return nil, ErrActorNameMustNotContainPathSep
	}

	if messageProcessorFactory == nil {
		return nil, ErrMessageProcessorFactoryRequired
	}
	if errorHandler == nil {
		return nil, ErrMessageProcessorErrorHandlerRequired
	}
	processor := messageProcessorFactory()
	if err := ValidateMessageProcessor(processor); err != nil {
		return nil, err
	}

	actorId := nuid.Next()
	actor := &Actor{
		system: a.system,
		parent: a,

		name: name,
		path: strings.Join([]string{a.path, name}, "/"),
		id:   actorId,

		created: time.Now(),

		children: make(map[string]*Actor),

		messageProcessorFactory: messageProcessorFactory,
		messageProcessor:        processor,
		errorHandler:            errorHandler,
		messageProcessorChan:    make(chan *Envelope),

		actorChan: make(chan interface{}),

		logger:       logger.With().Dict("actor", zerolog.Dict().Str("path", name).Str("id", actorId)).Logger(),
		uidGenerator: nuid.New(),
	}

	if !a.putChild(actor) {
		return nil, ActorAlreadyRegisteredError{a.path}
	}

	if !a.system.registerActor(actor) {
		a.deleteChild(actor.name)
		return nil, ActorAlreadyRegisteredError{a.path}
	}

	actor.start()

	return actor, nil
}

func (a *Actor) start() {
	a.Go(func() error {
		LOG_EVENT_STARTED.Log(a.logger.Debug()).Msg("")

		if receive, ok := a.messageProcessor.Handler(STARTED.MessageType()); ok {
			if err := receive(MessageContext{a, a.NewEnvelope(STARTED, a.Address(), nil, nil)}); err != nil {
				return err
			}
		}

	LOOP:
		for {
			select {
			case msg := <-a.messageProcessorChan:
				receive, ok := a.messageProcessor.Handler(msg.MessageType())
				if !ok {
					LOG_EVENT_MESSAGE_DROPPED.Log(a.logger.Error()).Err(MessageTypeNotSupportedError{msg.MessageType()})
					// TODO: if the envelope has a replyTo address, then reply with a MessageProcessingError
					continue LOOP
				}
				ctx := MessageContext{a, msg}
				if err := receive(ctx); err != nil {
					LOG_EVENT_MESSAGE_PROCESSING_ERR.Log(a.logger.Error()).Err(err).
						Str(LOG_FIELD_MESSAGE_ID, msg.id).
						Uint64(LOG_FIELD_MESSAGE_TYPE, msg.MessageType().UInt64()).
						Msg("")
					a.errorHandler(ctx, err)
				}
			case <-a.Dying():
				LOG_EVENT_DYING.Log(a.logger.Debug()).Msg("")
				if receive, ok := a.messageProcessor.Handler(STOPPING.MessageType()); ok {
					if err := receive(MessageContext{a, a.NewEnvelope(STOPPING, a.Address(), nil, nil)}); err != nil {
						LOG_EVENT_MESSAGE_PROCESSING_ERR.Log(a.logger.Error()).Err(err).Uint64(LOG_FIELD_MESSAGE_TYPE, STOPPING.MessageType().UInt64()).Msg("")
					}
				}
				a.stop()
				if receive, ok := a.messageProcessor.Handler(STOPPED.MessageType()); ok {
					if err := receive(MessageContext{a, a.NewEnvelope(STOPPED, a.Address(), nil, nil)}); err != nil {
						LOG_EVENT_MESSAGE_PROCESSING_ERR.Log(a.logger.Error()).Err(err).Uint64(LOG_FIELD_MESSAGE_TYPE, STOPPED.MessageType().UInt64()).Msg("")
					}
				}
				LOG_EVENT_DEAD.Log(a.logger.Debug()).Msg("")
				return nil
			case msg := <-a.actorChan:
				a.handleActorMessage(msg)
			}
		}
		return nil
	})

}

func (a *Actor) stop() {
	a.system.unregisterActor(a)
	LOG_EVENT_UNREGISTERED.Log(a.logger.Debug()).Msg("")
	if a.parent != nil {
		a.parent.deleteChild(a.name)
	}

	// kill all children
	children := a.Children()
	for _, child := range children {
		child.Kill(nil)
	}
	for _, child := range children {
		<-child.Dead()
	}
	a.clearChildren()
}

func (a *Actor) NewEnvelope(msg Message, to Address, replyTo *ReplyTo, correlationId *string) *Envelope {
	return NewEnvelope(a.uidGenerator.Next, to, msg, replyTo, correlationId)
}

// Tell blocks until the Actor MessageProcessor receives the message.
func (a *Actor) Tell(msg *Envelope) error {
	if msg == nil {
		return ErrEnvelopeNil
	}
	if _, ok := a.messageProcessor.Handler(msg.MessageType()); !ok {
		return MessageTypeNotSupportedError{msg.MessageType()}
	}
	select {
	case a.messageProcessorChan <- msg:
		return nil
	case <-a.Dying():
		return ErrNotAlive
	}
}

func (a *Actor) handleActorMessage(msg interface{}) {
	// TODO - e.g., get metrics
}
