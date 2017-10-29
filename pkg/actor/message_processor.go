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

	"github.com/rs/zerolog"
	"gopkg.in/tomb.v2"
)

type MessageProcessorFactory func() MessageProcessor

type MessageProcessor interface {
	// ChannelNames returns the channels that are supported by this message processor
	ChannelNames() []Channel

	Handler(channel Channel) Receive
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

type MessageProcessorEngine struct {
	tomb.Tomb

	MessageProcessor

	channels map[Channel]chan *MessageContext

	c chan *MessageContext

	logger zerolog.Logger
}

func (a *MessageProcessorEngine) start() {
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
					if err := receive(msg); err != nil {
						return err
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
					if err := a.Handler(msg.Envelope.channel)(msg); err != nil {
						return err
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
						a.c <- msg
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
