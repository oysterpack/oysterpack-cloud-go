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

	// Stopped is invoked after the message processor is stopped. The message processor should perform any cleanup here.
	Stopped()
}

// MessageChannelHandlers implements MessageProcessor. It is simply a mapping of Channel -> Receive
// If state is required by the MessageProcessor, then simply embed this and populate the map with Receive functions that
// point to MessageProcessor methods with the Receive function signature.
type MessageChannelHandlers map[Channel]MessageTypeHandlers

// MessageTypeHandlers is maps MessageType -> Receive handler function
type MessageTypeHandlers map[MessageType]Receive

func (a MessageChannelHandlers) ChannelNames() []Channel {
	channels := make([]Channel, len(a))
	i := 0
	for channel := range a {
		channels[i] = channel
		i++
	}
	return channels
}

func (a MessageChannelHandlers) Handler(channel Channel, msgType MessageType) Receive {
	if handlers := a[channel]; handlers != nil {
		return handlers[msgType]
	}
	return nil
}

func (a MessageChannelHandlers) MessageTypes(channel Channel) []MessageType {
	handlers := a[channel]
	if handlers == nil {
		return nil
	}
	msgTypes := make([]MessageType, len(handlers))
	i := 0
	for k := range handlers {
		msgTypes[i] = k
		i++
	}
	return msgTypes
}

func (a MessageChannelHandlers) Stopped() {}

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

// MessageProcessorEngine sets up a message channel pipeline :
//
//										 |   |   |   |
//										 V   V   V   V
//   MessageProcessorEngine channels  	[ ]	[ ]	[ ]	[ ]
//										 |   |   |   |
//										 V   V   V   V
//										 |___|___|___|
//											   |
//											   V
//   MessageProcessor channel				  [ ]
//
// For each MessageProcessor Channel, the MessageProcessorEngine will create a channel.
// Messages arriving on the incoming MessageProcessorEngine channels are fanned into the core MessageProcessor channel.
// Messages received on the core MessageProcessor channel are processed by the MessageProcessor.
type MessageProcessorEngine struct {
	tomb.Tomb

	MessageProcessor

	channels map[Channel]chan *MessageContext

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
	if len(a.ChannelNames()) == 1 {
		channel := a.ChannelNames()[0]
		a.channels = map[Channel]chan *MessageContext{channel: a.c}
		a.Go(func() error {
			LOG_EVENT_STARTED.Log(a.logger.Info()).Str(LOG_FIELD_CHANNEL, channel.String()).Msg("")
			LOG_EVENT_STARTED.Log(a.logger.Info()).Msg("")
			for {
				select {
				case <-a.Dying():
					LOG_EVENT_DYING.Log(a.logger.Info()).Str(LOG_FIELD_CHANNEL, channel.String()).Msg("")
					LOG_EVENT_DYING.Log(a.logger.Info()).Msg("")
					return nil
				case msg := <-a.c:
					if msg == nil {
						LOG_EVENT_NIL_MSG.Log(a.logger.Info()).Msg("")
						return nil
					}
					receive := a.Handler(channel, msg.Envelope.MessageType())
					if receive == nil {
						LOG_EVENT_UNKNOWN_MESSAGE_TYPE.Log(a.logger.Error()).Str(LOG_FIELD_CHANNEL, channel.String()).Msg("Message dropped")
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
	} else {
		a.channels = make(map[Channel]chan *MessageContext, len(a.ChannelNames()))

		// fan in all messages received on channels into the MessageProcessorEngine channel
		a.Go(func() error {
			LOG_EVENT_STARTED.Log(a.logger.Info()).Msg("")
			for {
				select {
				case <-a.Dying():
					LOG_EVENT_DYING.Log(a.logger.Info()).Msg("")
					return nil
				case msg := <-a.c:
					if msg == nil {
						LOG_EVENT_NIL_MSG.Log(a.logger.Info()).Msg("")
						return nil
					}
					receive := a.Handler(msg.Envelope.channel, msg.Envelope.MessageType())
					if receive == nil {
						LOG_EVENT_UNKNOWN_MESSAGE_TYPE.Log(a.logger.Error()).Str(LOG_FIELD_CHANNEL, msg.Envelope.channel.String()).Msg("Message dropped")
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

		for _, channel := range a.ChannelNames() {
			c := make(chan *MessageContext)
			a.channels[channel] = c
			channelName := channel.String()
			a.Go(func() error {
				LOG_EVENT_STARTED.Log(a.logger.Info()).Str(LOG_FIELD_CHANNEL, channelName).Msg("")
				for {
					select {
					case <-a.Dying():
						LOG_EVENT_DYING.Log(a.logger.Info()).Str(LOG_FIELD_CHANNEL, channelName).Msg("")
						return nil
					case msg := <-c:
						if msg == nil {
							LOG_EVENT_NIL_MSG.Log(a.logger.Info()).Str(LOG_FIELD_CHANNEL, channelName).Msg("")
							return nil
						}
						select {
						case <-a.Dying():
							LOG_EVENT_DYING.Log(a.logger.Info()).Str(LOG_FIELD_CHANNEL, channelName).Msg("")
							return nil
						case a.c <- msg:
						}

					}
				}
			})
		}
	}
	return nil
}

// Channel returns the channel that is used for sending messages to this message processor
func (a *MessageProcessorEngine) Channel(channel Channel) chan<- *MessageContext {
	return a.channels[channel]
}
