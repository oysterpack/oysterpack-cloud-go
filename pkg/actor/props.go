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
	"time"

	"github.com/nats-io/nuid"
	"github.com/rs/zerolog"
)

type Props struct {
	MessageProcessor *MessageProcessorProps

	zerolog.Logger
}

func (a *Props) Validate() error {
	if a.MessageProcessor.Factory == nil {
		return ErrMessageProcessorFactoryRequired
	}
	if a.MessageProcessor.ErrorHandler == nil {
		return ErrMessageProcessorErrorHandlerRequired
	}
	return nil
}

func (a *Props) checkActorName(name string) error {
	if name == "" {
		return ErrActorNameBlank
	}
	if strings.Contains(name, "/") {
		return ErrActorNameMustNotContainPathSep
	}
	return nil
}

func (a *Props) NewChild(parent *Actor, name string) (*Actor, error) {
	if parent == nil {
		return nil, ErrMessageProcessorFactoryRequired
	}
	if err := a.checkActorName(name); err != nil {
		return nil, err
	}

	if _, ok := parent.Child(name); ok {
		return nil, ActorAlreadyRegisteredError{strings.Join([]string{parent.path, name}, "/")}
	}

	processor, err := a.MessageProcessor.New()
	if err != nil {
		return nil, err
	}

	actorId := nuid.Next()
	path := strings.Join([]string{parent.path, name}, "/")
	actor := &Actor{
		system: parent.system,
		parent: parent,

		name:    name,
		path:    path,
		id:      actorId,
		address: &Address{path, &actorId},

		created: time.Now(),

		messageProcessorFactory: a.MessageProcessor.Factory,
		errorHandler:            a.MessageProcessor.ErrorHandler,
		messageProcessor:        processor,
		messageProcessorChan:    a.MessageProcessor.Channel(),

		actorChan: make(chan interface{}, 1),

		dying: make(chan struct{}),
		dead:  make(chan struct{}),

		logger:       logger.With().Dict("actor", zerolog.Dict().Str("path", path).Str("id", actorId)).Logger(),
		uidGenerator: nuid.New(),
	}

	if !parent.putChild(actor) {
		return nil, ActorAlreadyRegisteredError{actor.path}
	}

	if !parent.system.registerActor(actor) {
		parent.deleteChild(actor.name)
		return nil, ActorAlreadyRegisteredError{actor.path}
	}

	actor.start()

	return actor, nil
}

// NewRootActor creates a new Actor, registers it as top level root actor, and starts the Actor.
//
// - name cannot contain a '/'
func (a *Props) NewRootActor(system *System, name string) (*Actor, error) {
	if system == nil {
		return nil, ErrSystemNil
	}
	if err := a.checkActorName(name); err != nil {
		return nil, err
	}

	if _, ok := system.Actor(&Address{Path: name}); ok {
		return nil, ActorAlreadyRegisteredError{name}
	}

	processor, err := a.MessageProcessor.New()
	if err != nil {
		return nil, err
	}

	actorId := nuid.Next()
	actor := &Actor{
		system: system,

		name:    name,
		path:    name,
		id:      actorId,
		address: &Address{name, &actorId},

		created: time.Now(),

		children: make(map[string]*Actor),

		messageProcessorFactory: a.MessageProcessor.Factory,
		errorHandler:            a.MessageProcessor.ErrorHandler,
		messageProcessor:        processor,
		messageProcessorChan:    a.MessageProcessor.Channel(),

		actorChan: make(chan interface{}, 1),
		dying:     make(chan struct{}),
		dead:      make(chan struct{}),

		logger:       logger.With().Dict("actor", zerolog.Dict().Str("path", name).Str("id", actorId)).Logger(),
		uidGenerator: nuid.New(),
	}

	if !system.registerActor(actor) {
		return nil, ActorAlreadyRegisteredError{actor.path}
	}

	actor.start()

	return actor, nil
}

func (a MessageProcessorProps) New() (MessageProcessor, error) {
	processor := a.Factory()
	if err := ValidateMessageProcessor(processor); err != nil {
		return nil, err
	}
	return processor, nil
}

func (a MessageProcessorProps) Channel() chan *Envelope {
	if a.ChannelSize > 1 {
		return make(chan *Envelope, a.ChannelSize)
	}
	return make(chan *Envelope, 1)
}

type MessageProcessorProps struct {
	Factory      MessageProcessorFactory
	ErrorHandler MessageProcessorErrorHandler
	ChannelSize  uint
}
