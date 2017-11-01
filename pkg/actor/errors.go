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
	"bytes"
	"errors"
	"fmt"
	"reflect"

	"github.com/oysterpack/oysterpack.go/pkg/actor/msgs"
	"zombiezen.com/go/capnproto2"
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

func (a *MessageProcessingError) UnmarshalBinary(data []byte) error {
	decoder := capnp.NewPackedDecoder(bytes.NewBuffer(data))
	msg, err := decoder.Decode()
	if err != nil {
		return err
	}
	failure, err := msgs.ReadRootMessageProcessingError(msg)
	if err != nil {
		return err
	}

	if failure.HasPath() {
		pathList, err := failure.Path()
		if err != nil {
			return err
		}
		a.Path = make([]string, pathList.Len())
		for i := 0; i < pathList.Len(); i++ {
			a.Path[i], err = pathList.At(i)
			if err != nil {
				return err
			}
		}
	}

	msgBytes, err := failure.Message()
	if err != nil {
		return err
	}
	if err := a.Message.UnmarshalBinary(msgBytes); err != nil {
		return err
	}

	errMsg, err := failure.Err()
	if err != nil {
		return err
	}
	a.Err = errors.New(errMsg)
	return err
}

func (a *MessageProcessingError) MarshalBinary() ([]byte, error) {
	msg, seg, err := capnp.NewMessage(capnp.MultiSegment(nil))

	failure, err := msgs.NewRootMessageProcessingError(seg)
	if err != nil {
		return nil, err
	}

	if len(a.Path) > 0 {
		path, err := capnp.NewTextList(seg, int32(len(a.Path)))
		if err != nil {
			return nil, err
		}
		for i, p := range a.Path {
			if err := path.Set(i, p); err != nil {
				return nil, err
			}
		}
		if err = failure.SetPath(path); err != nil {
			return nil, err
		}
	}

	messageBytes, err := a.Message.MarshalBinary()
	if err != nil {
		return nil, err
	}
	if err = failure.SetMessage(messageBytes); err != nil {
		return nil, err
	}

	if err := failure.SetErr(a.Err.Error()); err != nil {
		return nil, err
	}

	buf := new(bytes.Buffer)
	encoder := capnp.NewPackedEncoder(buf)
	if err = encoder.Encode(msg); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}