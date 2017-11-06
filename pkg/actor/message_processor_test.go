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

	"errors"

	"github.com/oysterpack/oysterpack.go/pkg/actor"
)

func TestMessageHandler_Validate(t *testing.T) {

	handler := actor.MessageHandler{}
	if err := handler.Validate(); err == nil {
		t.Error("should not be valid")
	}

	handler.Receive = func(ctx *actor.MessageContext) error {
		return nil
	}

	if err := handler.Validate(); err == nil {
		t.Error("should not be valid")
	}

	handler.Unmarshal = func(msg []byte) (*actor.Envelope, error) {
		return nil, errors.New("ERROR")
	}

	if err := handler.Validate(); err != nil {
		t.Error("should be valid")
	}

	handler.Receive = nil
	if err := handler.Validate(); err == nil {
		t.Error("should not be valid")
	}
}

func TestMessageHandlers(t *testing.T) {

	handlers := actor.MessageHandlers{}
	if err := handlers.Validate(); err == nil {
		t.Error("at least 1 handler must be defined")
	}

	handlers[actor.MessageType(1)] = actor.MessageHandler{
		Receive: func(ctx *actor.MessageContext) error {
			return nil
		},
		Unmarshal: func(msg []byte) (*actor.Envelope, error) {
			return nil, errors.New("ERROR")
		},
	}
	if err := handlers.Validate(); err != nil {
		t.Error(err)
	}

	var messageProcessor actor.MessageProcessor = handlers
	if messageProcessor.Handler(actor.MessageType(1)) == nil {
		t.Error("handler should not be nil")
	}

	handlers[actor.MessageType(0)] = actor.MessageHandler{
		Receive: func(ctx *actor.MessageContext) error {
			return nil
		},
		Unmarshal: func(msg []byte) (*actor.Envelope, error) {
			return nil, errors.New("ERROR")
		},
	}
	if err := handlers.Validate(); err == nil {
		t.Error("MessageType(0) is not valid")
	}
	delete(handlers, actor.MessageType(0))
	if err := handlers.Validate(); err != nil {
		t.Error(err)
	}

	handler := handlers[actor.MessageType(1)]
	handler.Receive = nil
	if messageProcessor.Handler(actor.MessageType(1)) == nil {
		t.Error("handler should not be nil")
	}
}
