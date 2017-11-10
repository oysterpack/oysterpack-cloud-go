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

type MessageProcessorFactory func() MessageProcessor

// MessageProcessor maps message receive handlers to MessageChannel/MessageType
// It also knows how to unmarshal supported message types.
type MessageProcessor interface {

	// MessageTypes returns the message types that are supported
	MessageTypes() []MessageType

	// Handler returns a Handler function that will be used to handle messages for the specified message type.
	// If the MessageType is not supported, then false is returned.
	Handler(msgType MessageType) (Receive, bool)

	// UnmarshalMessage is used to unmarshal incoming messages - in support of distributed messaging.
	// If the MessageType is not supported, then false is returned.
	Unmarshaller(msgType MessageType) (Unmarshal, bool)
}

// MessageProcessorErrorHandler handles message processing errors
type MessageProcessorErrorHandler func(ctx *MessageContext, err error)

type Receive func(ctx *MessageContext) error

type Unmarshal func(msg []byte) (*Envelope, error)

func Unmarshaller(prototype Message) Unmarshal {
	return func(msg []byte) (*Envelope, error) {
		envelope := EmptyEnvelope(prototype)
		if err := envelope.UnmarshalBinary(msg); err != nil {
			return nil, err
		}
		return envelope, nil
	}
}

// MessageHandlers implements MessageProcessor.
// For statefule message handlers, embed this within a struct.
type MessageHandlers map[MessageType]*MessageHandler

type MessageHandler struct {
	Receive
	Unmarshal
}

func (a MessageHandler) Validate() error {
	if a.Receive == nil {
		return ErrMessageHandlerReceiveRequired
	}
	if a.Unmarshal == nil {
		return ErrMessageHandlerUnmarshalRequired
	}
	return nil
}

func (a MessageHandlers) Handler(msgType MessageType) (Receive, bool) {
	if handler := a[msgType]; handler != nil {
		return handler.Receive, true
	}

	return nil, false
}

func (a MessageHandlers) Unmarshaller(msgType MessageType) (Unmarshal, bool) {
	if handler := a[msgType]; handler != nil {
		return handler.Unmarshal, true
	}
	return nil, false
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

func ValidateMessageProcessor(a MessageProcessor) error {
	if len(a.MessageTypes()) == 0 {
		return ErrMessageProcessorNoMessageTypes
	}
	if _, ok := a.Handler(MessageType(0)); ok {
		return ErrInvalidMessageType
	}
	for _, messageType := range a.MessageTypes() {
		if receive, ok := a.Handler(messageType); !ok || receive == nil {
			return messageProcessorMissingHandler(messageType)
		}

		if unmarshaller, ok := a.Unmarshaller(messageType); !ok || unmarshaller == nil {
			return messageProcessorMissingUnmarshaller(messageType)
		}
	}
	return nil
}

type MessageProcessingError struct {
	actorAddress Address

	messageId   string
	messageType MessageType

	errCode    int64
	errMessage string
}
