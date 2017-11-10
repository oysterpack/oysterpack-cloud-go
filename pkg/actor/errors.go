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
	"errors"
	"fmt"
)

var (
	ErrInvalidMessageType = errors.New("MessageType must not be 0")

	ErrEnvelopeNil                = errors.New("Envelope is nil")
	ErrEnvelopeIdBlank            = errors.New("Envelope.Id cannot be blank")
	ErrEnvelopeCreatedRquired     = errors.New("Envelope.Created is not set")
	ErrEnvelopeMessageRequired    = errors.New("Envelope.Message is required")
	ErrEnvelopeCorrelationIdBlank = errors.New("Envelope.CorrelationId cannot be blank")
	ErrEnvelopeAddressRequired    = errors.New("Envelope.Address is required")

	ErrMessageHandlerReceiveRequired   = errors.New("MessageHandler.Receive is required")
	ErrMessageHandlerUnmarshalRequired = errors.New("MessageHandler.Unmarshal is required")

	ErrActorNameBlank                       = errors.New("Actor.Name cannot be blank")
	ErrActorNameMustNotContainPathSep       = errors.New("Actor.Name cannot contain '/'")
	ErrMessageProcessorFactoryRequired      = errors.New("MessageProcessorFactory is required")
	ErrMessageProcessorErrorHandlerRequired = errors.New("MessageProcessorErrorHandler is required")

	ErrStillAlive = errors.New("Still alive")
	ErrNotAlive   = errors.New("Not alive")

	ErrMessageProcessorNoMessageTypes = errors.New("At least 1 MessageType must be defined")

	ErrChannelBlocked = errors.New("Channel blocked")

	ErrMessageNil = errors.New("Message is nil")
	ErrSystemNil  = errors.New("System is nil")
)

func envelopeMessageTypeDoesNotMatch(envelope *Envelope) error {
	return fmt.Errorf("Envelope.MessageType(%x) does not match Message.MessageType(%x)", envelope.msgType, envelope.message.MessageType())
}

func messageProcessorMissingHandler(messageType MessageType) error {
	return fmt.Errorf("MessageProcessor is missing Receive handler for MessageType(%x)", messageType)
}

func messageProcessorMissingUnmarshaller(messageType MessageType) error {
	return fmt.Errorf("MessageProcessor is missing Unmarshaller for MessageType(%x)", messageType)
}

type ActorAlreadyRegisteredError struct {
	Path string
}

func (a ActorAlreadyRegisteredError) Error() string {
	return fmt.Sprint("Actor already registered at path : ", a.Path)
}

// MessageTypeNotSupportedError indicates the actor received a message it does not support
type MessageTypeNotSupportedError struct {
	MessageType
}

func (a MessageTypeNotSupportedError) Error() string {
	return a.String()
}

func (a MessageTypeNotSupportedError) String() string {
	return fmt.Sprintf("Actor does not support message type : %x", a.MessageType)
}

type ActorNotFoundError struct {
	*Address
}

func (a ActorNotFoundError) Error() string {
	return fmt.Sprint("Actor not found at : %+v", a.Address)
}
