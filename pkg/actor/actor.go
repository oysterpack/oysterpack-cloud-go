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

	"github.com/nats-io/nuid"
	"github.com/oysterpack/oysterpack.go/pkg/commons"
	"github.com/rs/zerolog"
)

type Actor struct {
	system *System

	name    string
	path    string
	id      string
	address *Address

	created time.Time

	parent *Actor

	childrenLock sync.RWMutex
	children     map[string]*Actor

	messageProcessorFactory MessageProcessorFactory
	errorHandler            MessageProcessorErrorHandler

	messageProcessor     MessageProcessor
	messageProcessorChan chan *Envelope
	err                  error

	actorChan chan interface{}

	dying chan struct{}
	dead  chan struct{}

	// used to :
	// 	- generate unique ids for children
	//	- generate unique message envelope ids
	uidGenerator *nuid.NUID
	uidLock      sync.Mutex

	logger zerolog.Logger
}

func (a *Actor) Alive() bool {
	select {
	case <-a.dead:
		return false
	case <-a.dying:
		return false
	default:
		return true
	}
}

func (a *Actor) Dying() <-chan struct{} {
	return a.dying
}

func (a *Actor) Dead() <-chan struct{} {
	return a.dead
}

func (a *Actor) Kill() {
	if a.Alive() {
		commons.CloseQuietly(a.dying)
	}
}

// Err returns the death reason, or ErrStillAlive if the actor is not in a dying or dead state.
func (a *Actor) Err() error {
	return a.err
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

func (a *Actor) Address() *Address {
	return a.address
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

func (a *Actor) ChildrenCount() int {
	a.childrenLock.RLock()
	count := len(a.children)
	a.childrenLock.RUnlock()
	return count
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
	if a.children == nil {
		a.children = map[string]*Actor{child.name: child}
		a.childrenLock.Unlock()
		return true
	}
	_, ok := a.children[child.name]
	if ok {
		a.childrenLock.Unlock()
		return false
	}
	a.children[child.name] = child
	a.childrenLock.Unlock()
	return true
}

func (a *Actor) deleteChild(name string) {
	a.childrenLock.Lock()
	if a.children != nil {
		delete(a.children, name)
	}
	a.childrenLock.Unlock()
}

func (a *Actor) clearChildren() {
	a.childrenLock.Lock()
	a.children = make(map[string]*Actor)
	a.childrenLock.Unlock()
}

func (a *Actor) start() {
	gofuncs <- func() {
		defer close(a.dead)

		if err := a.startedMessageProcessor(); err != nil {
			a.err = err
			return
		}

		LOG_EVENT_STARTED.Log(a.logger.Debug()).Msg("")

		ctx := &MessageContext{Actor: a}

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
				ctx.Message = msg
				if err := receive(ctx); err != nil {
					LOG_EVENT_MESSAGE_PROCESSING_ERR.Log(a.logger.Error()).Err(err).
						Str(LOG_FIELD_MESSAGE_ID, msg.id).
						Uint64(LOG_FIELD_MESSAGE_TYPE, msg.MessageType().UInt64()).
						Msg("")
					a.errorHandler(ctx, err)
				}
			case <-a.dying:
				LOG_EVENT_DYING.Log(a.logger.Debug()).Msg("")
				a.stoppingMessageProcessor()
				a.stop()
				a.stoppedMessageProcessor()
				LOG_EVENT_DEAD.Log(a.logger.Debug()).Msg("")
				return
			case msg := <-a.actorChan:
				a.handleActorMessage(msg)
			}
		}
	}
}

func (a *Actor) startedMessageProcessor() error {
	if receive, ok := a.messageProcessor.Handler(STARTED.MessageType()); ok {
		if err := receive(&MessageContext{a, a.NewEnvelope(STARTED, a.Address(), nil, nil)}); err != nil {
			LOG_EVENT_START_FAILED.Log(a.logger.Error()).Err(err).Msg("")
			return err
		}
	}
	return nil
}

func (a *Actor) stoppingMessageProcessor() {
	if receive, ok := a.messageProcessor.Handler(STOPPING.MessageType()); ok {
		if err := receive(&MessageContext{a, a.NewEnvelope(STOPPING, a.Address(), nil, nil)}); err != nil {
			LOG_EVENT_MESSAGE_PROCESSING_ERR.Log(a.logger.Error()).Err(err).Uint64(LOG_FIELD_MESSAGE_TYPE, STOPPING.MessageType().UInt64()).Msg("")
		}
	}
}

func (a *Actor) stoppedMessageProcessor() {
	if receive, ok := a.messageProcessor.Handler(STOPPED.MessageType()); ok {
		if err := receive(&MessageContext{a, a.NewEnvelope(STOPPED, a.Address(), nil, nil)}); err != nil {
			LOG_EVENT_MESSAGE_PROCESSING_ERR.Log(a.logger.Error()).Err(err).Uint64(LOG_FIELD_MESSAGE_TYPE, STOPPED.MessageType().UInt64()).Msg("")
		}
	}
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
		child.Kill()
	}
	for _, child := range children {
		<-child.Dead()
	}
	a.clearChildren()
}

// Envelope returns a new Envelope for the specified message that is addressed to this actor
func (a *Actor) Envelope(msg Message, replyTo *ReplyTo, correlationId *string) *Envelope {
	return NewEnvelope(a.NextUID, a.Address(), msg, replyTo, correlationId)
}

func (a *Actor) NewEnvelope(msg Message, to *Address, replyTo *ReplyTo, correlationId *string) *Envelope {
	return NewEnvelope(a.NextUID, to, msg, replyTo, correlationId)
}

func (a *Actor) NextUID() string {
	a.uidLock.Lock()
	uid := a.uidGenerator.Next()
	a.uidLock.Unlock()
	return uid
}

// Tell delivers the message to the Actor's MessageProcessor. The operation is non-blocking and back-pressured.
// The MessageProcessor channel naturally applies back pressure. If the channel buffer is full, then ErrChannelBlocked
// is returned.
//
// errors:
// 	ErrNotAlive - messages cannot be sent to an actor that is not alive
//	ErrChannelBlocked - means the MessageProcessor channel is full, which is how back-pressure is applied
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
	case <-a.dying:
		return ErrNotAlive
	case <-a.dead:
		return ErrNotAlive
	default:
		return ErrChannelBlocked
	}
}

// TellBlocking will block until either the message is delivered or the actor is no longer alive.
// When the actor is not alive, ErrNotAlive is returned.
func (a *Actor) TellBlocking(msg *Envelope) error {
	if msg == nil {
		return ErrEnvelopeNil
	}
	if _, ok := a.messageProcessor.Handler(msg.MessageType()); !ok {
		return MessageTypeNotSupportedError{msg.MessageType()}
	}
	select {
	case a.messageProcessorChan <- msg:
		return nil
	case <-a.dying:
		return ErrNotAlive
	case <-a.dead:
		return ErrNotAlive
	}
}

func (a *Actor) handleActorMessage(msg interface{}) {
	// TODO - e.g., get metrics
}
