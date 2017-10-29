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

	"github.com/oysterpack/oysterpack.go/pkg/commons"
	"github.com/rs/zerolog"
	"gopkg.in/tomb.v2"
)

// MessageProcessorFactory is used to provide MessageProcessor instances
type MessageProcessorFactory func() MessageProcessor

// MessageProcessor is the backend actor message processor
type MessageProcessor interface {
	// ChannelNames returns the channels that are supported by this message processor
	ChannelNames() []Channel

	// Handler returns a Receive function that will be used to handle messages on the specified channel
	Handler(channel Channel) Receive

	// Stopped is invoked after the message processer is stopped. The message processor should perform any cleanup here.
	Stopped()
}

func ValidateMessageProcessor(p MessageProcessor) error {
	if len(p.ChannelNames()) == 0 {
		return errors.New("MessageProcessor must have at least 1 channel defined")
	}

	for _, channel := range p.ChannelNames() {
		if err := channel.Validate(); err != nil {
			return err
		}
		//if p.Channel(channel) == nil {
		//	fmt.Errorf("Channel is missing : %s", channel)
		//}
		if p.Handler(channel) == nil {
			fmt.Errorf("Handler is missing for channel : %s", channel)
		}
	}

	return nil
}

func StartMessageProcessorEngine(messageProcessor MessageProcessor, logger zerolog.Logger) (*MessageProcessorEngine, error) {
	if err := ValidateMessageProcessor(messageProcessor); err != nil {
		return nil, err
	}
	engine := &MessageProcessorEngine{
		MessageProcessor: messageProcessor,
		logger:           logger,
	}
	engine.start()
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

func (a *MessageProcessorEngine) closeChannels() {
	a.lock.Lock()
	defer a.lock.Unlock()

	closeQuietly := func(c chan *MessageContext) {
		defer commons.IgnorePanic()
		close(c)
	}

	for _, c := range a.channels {
		closeQuietly(c)
	}

	closeQuietly(a.c)
}

func (a *MessageProcessorEngine) start() {
	a.lock.Lock()
	defer a.lock.Unlock()

	a.c = make(chan *MessageContext)
	if len(a.ChannelNames()) == 1 {
		channel := a.ChannelNames()[0]
		receive := a.Handler(channel)
		a.channels = map[Channel]chan *MessageContext{channel: a.c}
		a.Go(func() error {
			MESSAGE_PROCESSOR_CHANNEL_STARTED.Log(a.logger.Info()).Str(CHANNEL, channel.String()).Msg("")
			MESSAGE_PROCESSOR_STARTED.Log(a.logger.Info()).Msg("")
			for {
				select {
				case <-a.Dying():
					MESSAGE_PROCESSOR_CHANNEL_DYING.Log(a.logger.Info()).Str(CHANNEL, channel.String()).Msg("")
					MESSAGE_PROCESSOR_DYING.Log(a.logger.Info()).Msg("")
					return nil
				case msg := <-a.c:
					if msg == nil {
						MESSAGE_PROCESSOR_CHANNEL_CLOSED.Log(a.logger.Info()).Msg("")
						return nil
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
			MESSAGE_PROCESSOR_STARTED.Log(a.logger.Info()).Msg("")
			for {
				select {
				case <-a.Dying():
					MESSAGE_PROCESSOR_DYING.Log(a.logger.Info()).Msg("")
					return nil
				case msg := <-a.c:
					if msg == nil {
						MESSAGE_PROCESSOR_CHANNEL_CLOSED.Log(a.logger.Info()).Msg("")
						return nil
					}
					if err := a.Handler(msg.Envelope.channel)(msg); err != nil {
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
				MESSAGE_PROCESSOR_CHANNEL_STARTED.Log(a.logger.Info()).Str(CHANNEL, channelName).Msg("")
				for {
					select {
					case <-a.Dying():
						MESSAGE_PROCESSOR_CHANNEL_DYING.Log(a.logger.Info()).Str(CHANNEL, channelName).Msg("")
						return nil
					case msg := <-c:
						if msg == nil {
							MESSAGE_PROCESSOR_CHANNEL_CLOSED.Log(a.logger.Info()).Str(CHANNEL, channelName).Msg("")
							return nil
						}
						select {
						case <-a.Dying():
							MESSAGE_PROCESSOR_CHANNEL_DYING.Log(a.logger.Info()).Str(CHANNEL, channelName).Msg("")
							return nil
						case a.c <- msg:
						}

					}
				}
			})
		}

	}
}

// Channel returns the channel that is used for sending messages to this message processor
func (a *MessageProcessorEngine) Channel(channel Channel) chan<- *MessageContext {
	return a.channels[channel]
}
