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

package actor2

import (
	"errors"
	"fmt"
	"reflect"
)

var (
	ErrStillAlive = errors.New("Still alive")

	ErrAlreadyStarted = errors.New("Already started")

	ErrNotAlive = errors.New("Not alive")
)

// ChannelNotSupportedError indicates the channel is unknown
type ChannelNotSupportedError struct {
	Channel
}

func (a *ChannelNotSupportedError) Error() string {
	return a.String()
}

func (a *ChannelNotSupportedError) String() string {
	return fmt.Sprintf("Invalid channel : %s", a.Channel)
}

// ChannelNotSupportedError indicates the channel is unknown
type ChannelMessageTypeNotSupportedError struct {
	Channel
	MessageType
}

func (a *ChannelMessageTypeNotSupportedError) Error() string {
	return a.String()
}

func (a *ChannelMessageTypeNotSupportedError) String() string {
	return fmt.Sprintf("Channel (%s) does not support message type : %d", a.Channel, a.MessageType)
}

type InvalidMessageTypeError struct {
	Message      interface{}
	ExpectedType reflect.Type
}

func (a *InvalidMessageTypeError) Error() string {
	return a.String()
}

func (a *InvalidMessageTypeError) String() string {
	return fmt.Sprintf("Invalid message type : %T. Expected message type : %v", a.Message, a.ExpectedType)
}

// InvalidChannelMessageTypeError indicates the channel does not support the specified message type
type InvalidChannelMessageTypeError struct {
	Channel
	Message interface{}
}

func (a *InvalidChannelMessageTypeError) Error() string {
	return a.String()
}

func (a *InvalidChannelMessageTypeError) String() string {
	return fmt.Sprintf("Channel (%s) does not support message type : %T", a.Channel, a.Message)
}

// InvalidMessageError indicates the message invalid.
type InvalidMessageError struct {
	Message interface{}
	Err     error
}

func (a *InvalidMessageError) Error() string {
	return a.String()
}

func (a *InvalidMessageError) String() string {
	return fmt.Sprintf("Message was invalid : %T : %v", a.Message, a.Err.Error())
}

// ActorNotAliveError indicates an invalid operation was performed on an actor that is not alive.
type ActorNotAliveError struct {
	*Address
}

func (a *ActorNotAliveError) Error() string {
	return a.String()
}

func (a *ActorNotAliveError) String() string {
	return fmt.Sprintf("Actor is not alive : %s", a.Address)
}

// ActorNotFoundError indicates that no actor lives at the specified address
type ActorNotFoundError struct {
	*Address
}

func (a *ActorNotFoundError) Error() string {
	return a.String()
}

func (a *ActorNotFoundError) String() string {
	return fmt.Sprintf("No actor lives at address : %s", a.Address)
}

type ChildAlreadyExists struct {
	Path []string
}

func (a *ChildAlreadyExists) Error() string {
	return a.String()
}

func (a *ChildAlreadyExists) String() string {
	return fmt.Sprintf("Child already exists at path : %v", a.Path)
}

type ProducerError struct {
	Err error
}

func (a *ProducerError) Error() string {
	return a.String()
}

func (a *ProducerError) String() string {
	return fmt.Sprintf("Child already exists at path : %v", a.Err)
}

// MessageProcessingError
type MessageProcessingError struct {
	Path    []string
	Message *Envelope
	Err     error
}

func (a *MessageProcessingError) Error() string {
	return a.String()
}

func (a *MessageProcessingError) String() string {
	return fmt.Sprintf("MessageProcessingError : %v : %v : %v", a.Path, a.Err, a.Message)
}
