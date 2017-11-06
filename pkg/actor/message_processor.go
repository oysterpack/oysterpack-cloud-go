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

import "errors"

type MessageProcessorFactory func() MessageProcessor

// MessageProcessor maps message receive handlers to MessageChannel/MessageType
// It also knows how to unmarshal supported message types.
type MessageProcessor interface {

	// MessageTypes returns the message types that are supported for the specified channel
	MessageTypes() []MessageType

	// Handler returns a Handler function that will be used to handle messages on the specified channel for the
	// specified message type
	Handler(msgType MessageType) Receive

	// UnmarshalMessage is used to unmarshal incoming messages - in support of distributed messaging.
	//
	// errors :
	// 	- ChannelMessageTypeNotSupportedError
	//  - unmarshalling errors
	Unmarshaler(msgType MessageType) Unmarshal
}

type Receive func(ctx *MessageContext) error

type Unmarshal func(msg []byte) (*Envelope, error)

type MessageContext struct {
	*Actor
	Message *Envelope
}

// MessageHandlers implements MessageProcessor.
// For statefule message handlers, embed this within a struct.
type MessageHandlers map[MessageType]MessageHandler

type MessageHandler struct {
	Receive
	Unmarshal
}

func (a MessageHandler) Validate() error {
	if a.Receive == nil {
		return errors.New("MessageHandler.Handler is required")
	}
	if a.Unmarshal == nil {
		return errors.New("MessageHandler.Unmarshal is required")
	}
	return nil
}

func (a MessageHandlers) Handler(msgType MessageType) Receive {
	return a[msgType].Receive
}

func (a MessageHandlers) Unmarshaler(msgType MessageType) Unmarshal {
	return a[msgType].Unmarshal
}

func (a MessageHandlers) MessageTypes() []MessageType {
	msgTypes := make([]MessageType, len(a))
	i := 0
	for k := range a {
		msgTypes[i] = k
		i++
	}
	return msgTypes
}

func (a MessageHandlers) Validate() error {
	if len(a) == 0 {
		return errors.New("No handlers are defined")
	}
	for k, v := range a {
		if err := k.Validate(); err != nil {
			return err
		}
		if err := v.Validate(); err != nil {
			return err
		}
	}
	return nil
}
