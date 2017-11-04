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

import "github.com/oysterpack/oysterpack.go/pkg/logging"

type MessageHandlers map[MessageChannelKey]MessageHandler

type MessageHandler struct {
	Receive
	Unmarshal func(msg []byte) (*Envelope, error)
}

type MessageChannelKey struct {
	Channel
	MessageType
}

func (a MessageHandlers) ChannelNames() []Channel {
	channels := make([]Channel, len(a))
	i := 0
	for channel := range a {
		channels[i] = channel.Channel
		i++
	}
	return channels
}

func (a MessageHandlers) Handler(channel Channel, msgType MessageType) Receive {
	return a[MessageChannelKey{channel, msgType}].Receive
}

func (a MessageHandlers) UnmarshalMessage(channel Channel, msgType MessageType, msg []byte) (*Envelope, error) {
	unmarshal := a[MessageChannelKey{channel, msgType}].Unmarshal
	if unmarshal == nil {
		return nil, &ChannelMessageTypeNotSupportedError{channel, msgType}
	}
	return unmarshal(msg)
}

func (a MessageHandlers) MessageTypes(channel Channel) []MessageType {
	msgTypes := []MessageType{}
	for key := range a {
		if key.Channel == channel {
			msgTypes = append(msgTypes, key.MessageType)
		}
	}
	return msgTypes
}

func (a MessageHandlers) Stopped() {
	LOG_EVENT_STOPPED.Log(logger.Debug()).Str(logging.TYPE, "MessageHandlers").Msg("")
}
