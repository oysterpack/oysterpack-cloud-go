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

	handler.Receive = func(ctx actor.MessageContext) error {
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
	// Given an empty MessageHandlers
	handlers := actor.MessageHandlers{}
	// Then validation should fail because at least 1 MessageType handler needs to be defined
	if err := actor.ValidateMessageProcessor(handlers); err == nil {
		t.Error("at least 1 handler must be defined")
	}

	// When a Handler is defined
	const MSG_TYPE_1 = actor.MessageType(1)
	handlers[MSG_TYPE_1] = actor.MessageHandler{
		Receive: func(ctx actor.MessageContext) error {
			return nil
		},
		Unmarshal: func(msg []byte) (*actor.Envelope, error) {
			return nil, errors.New("ERROR")
		},
	}
	// Then validation passes
	if err := actor.ValidateMessageProcessor(handlers); err != nil {
		t.Error(err)
	}

	// And MessageHandlers implements the MessageProcessor interface
	var messageProcessor actor.MessageProcessor = handlers
	if receive, ok := messageProcessor.Handler(MSG_TYPE_1); !ok {
		t.Error("handler should not be nil")
	} else {
		receive(actor.MessageContext{})
	}

	// When the handler is looked up for a supported MessageType
	// Then is it successfully returned
	if receive, ok := messageProcessor.Handler(MSG_TYPE_1); !ok || receive == nil {
		t.Error("%v : %v", receive, ok)
	}

	// Given a MessageHandler mapped to a MessageType(0)
	handlers[actor.MessageType(0)] = actor.MessageHandler{
		Receive: func(ctx actor.MessageContext) error {
			return nil
		},
		Unmarshal: func(msg []byte) (*actor.Envelope, error) {
			return nil, errors.New("ERROR")
		},
	}
	// Then validation should fail, because MessageType(0) is not permitted
	if err := actor.ValidateMessageProcessor(handlers); err == nil {
		t.Error("MessageType(0) is not valid")
	}

	// When the invalid mapping is removed
	delete(handlers, actor.MessageType(0))
	if err := actor.ValidateMessageProcessor(handlers); err != nil {
		t.Error(err)
	}
	// Then validation passes
	if err := actor.ValidateMessageProcessor(handlers); err != nil {
		t.Error(err)
	}

	// When the returned handler Receive is set to nil
	handler := handlers[MSG_TYPE_1]
	handler.Receive = nil
	// Then it does not affect the stored handler because the map returns a copy
	if _, ok := messageProcessor.Handler(actor.MessageType(1)); !ok {
		t.Error("handler should not be nil")
	}
}
