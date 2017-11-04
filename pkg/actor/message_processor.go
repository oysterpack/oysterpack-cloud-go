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
	"fmt"

	"errors"

	"sync"

	"github.com/oysterpack/oysterpack.go/pkg/logging"
	"github.com/rs/zerolog"
	"gopkg.in/tomb.v2"
)

// MessageProcessorFactory is used to provide MessageProcessor instances
type MessageProcessorFactory func() MessageProcessor

// MessageProcessor is the backend actor message processor
type MessageProcessor interface {
	// ChannelNames returns the channels that are supported by this message processor
	ChannelNames() []Channel

	// MessageTypes returns the message types that are supported for the specified channel
	MessageTypes(channel Channel) []MessageType

	// Handler returns a Receive function that will be used to handle messages on the specified channel for the
	// specified message type
	Handler(channel Channel, msgType MessageType) Receive

	// UnmarshalMessage is used to unmarshal incoming messages - in support of distributed messaging.
	//
	// errors :
	// 	- ChannelMessageTypeNotSupportedError
	//  - unmarshalling errors
	UnmarshalMessage(channel Channel, msgType MessageType, msg []byte) (*Envelope, error)

	// Stopped is invoked after the message processor is stopped. The message processor should perform any cleanup here.
	Stopped()
}

// ValidateMessageProcessor performs the following checks :
//
// 1. At least 1 channel must be defined
// 2. For each channel, at least 1 MessageType must be defined
// 3. For each MessageType, the handler function must not defined, i.e., not nil
func ValidateMessageProcessor(p MessageProcessor) error {
	if len(p.ChannelNames()) == 0 {
		return errors.New("MessageProcessor must have at least 1 channel defined")
	}

	for _, channel := range p.ChannelNames() {
		if err := channel.Validate(); err != nil {
			return err
		}

		msgTypes := p.MessageTypes(channel)
		if len(msgTypes) == 0 {
			fmt.Errorf("Channel has no message types : %s", channel)
		}

		for _, msgType := range msgTypes {
			if p.Handler(channel, msgType) == nil {
				fmt.Errorf("Handler is missing for channel : %s", channel)
			}
		}
	}

	return nil
}

// StartMessageProcessorEngine creates a new MessageProcessorEngine for the MessageProcessor and starts it
func StartMessageProcessorEngine(messageProcessor MessageProcessor, logger zerolog.Logger) (*MessageProcessorEngine, error) {
	engine := &MessageProcessorEngine{
		MessageProcessor: messageProcessor,
		logger:           logger.With().Str(logging.TYPE, "MessageProcessorEngine").Logger(),
	}
	if err := engine.start(); err != nil {
		return nil, err
	}
	return engine, nil
}

// MessageProcessorEngine sets up a channel to receive incoming messages and routes them to message handlers based on
// channel and message type.
type MessageProcessorEngine struct {
	tomb.Tomb

	MessageProcessor

	c chan *MessageContext

	logger zerolog.Logger

	lock sync.Mutex
}

func (a *MessageProcessorEngine) start() error {
	if err := ValidateMessageProcessor(a); err != nil {
		return err
	}

	a.lock.Lock()
	defer a.lock.Unlock()

	go func() {
		a.Wait()
		a.Stopped()
	}()

	a.c = make(chan *MessageContext)
	a.Go(func() error {
		LOG_EVENT_STARTED.Log(a.logger.Debug()).Msg("")
		for {
			select {
			case <-a.Dying():
				LOG_EVENT_DYING.Log(a.logger.Debug()).Msg("")
				return nil
			case msg := <-a.c:
				if msg == nil {
					LOG_EVENT_NIL_MSG.Log(a.logger.Debug()).Msg("")
					return nil
				}
				receive := a.Handler(msg.Envelope.channel, msg.Envelope.MessageType())
				if receive == nil {
					LOG_EVENT_UNSUPPORTED_MESSAGE.Log(a.logger.Error()).
						Str(LOG_FIELD_CHANNEL, msg.Envelope.channel.String()).
						Uint8(LOG_FIELD_MSG_TYPE, msg.Envelope.msgType.UInt8()).
						Msg("Message dropped")
					continue
				}
				if err := receive(msg); err != nil {
					return &MessageProcessingError{
						Path:    msg.Path(),
						Message: msg.Envelope,
						Err:     err,
					}
				}
			}
		}
	})
	return nil
}

// Channel returns the channel that is used for sending messages to this message processor
func (a *MessageProcessorEngine) Channel() chan<- *MessageContext {
	return a.c
}
