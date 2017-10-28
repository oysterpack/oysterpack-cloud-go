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
	"time"

	"encoding"

	"bytes"
	"compress/zlib"

	"github.com/json-iterator/go"
	"github.com/oysterpack/oysterpack.go/pkg/actor/msgs"
	"zombiezen.com/go/capnproto2"
)

// Message that is sent between actors.
// All messages must know how marshal themselves to a binary format.
// All messages are delivered within an Envelope. The envelope will compress the entire envelope, thus there is no need
// to perform compression in this layer.
type Message interface {
	encoding.BinaryMarshaler
	encoding.BinaryUnmarshaler
}

type MessageFactory interface {
	// NewMessage creates a new Message instance
	NewMessage() Message

	// Validate checks if the msg is compatible and valid for message types produced by this MessageFactory
	Validate(msg Message) error
}

// UID is a function that returns a unique id.
type UID func() string

// NewEnvelope creates a new Envelope wrapping the specified message
// 	- uid is used to generate the envelope message id
func NewEnvelope(uid UID, channel string, msg Message, replyTo *ChannelAddress) *Envelope {
	return &Envelope{
		id:      uid(),
		created: time.Now(),
		channel: channel,
		message: msg,
		replyTo: replyTo,
	}
}

func EmptyEnvelope(messageFactory MessageFactory) *Envelope {
	return &Envelope{
		message: messageFactory.NewMessage(),
	}
}

// Envelope is a message envelope. Envelope is itself a message.
type Envelope struct {
	id      string
	created time.Time

	channel string
	message Message

	replyTo *ChannelAddress
}

func (a *Envelope) Id() string {
	return a.id
}

func (a *Envelope) Created() time.Time {
	return a.created
}

func (a *Envelope) Channel() string {
	return a.channel
}

func (a *Envelope) Message() Message {
	return a.message
}

func (a *Envelope) ReplyTo() *ChannelAddress {
	return a.replyTo
}

// UnmarshalBinary implements encoding.BinaryUnmarshaler
func (a *Envelope) UnmarshalBinary(data []byte) error {
	decompressor, _ := zlib.NewReader(bytes.NewBuffer(data))
	decoder := capnp.NewPackedDecoder(decompressor)
	msg, err := decoder.Decode()
	if err != nil {
		return err
	}
	envelope, err := msgs.ReadRootEnvelope(msg)
	if err != nil {
		return err
	}

	id, err := envelope.Id()
	if err != nil {
		return err
	}
	a.id = id
	a.created = time.Unix(0, envelope.Created())

	channel, err := envelope.Channel()
	if err != nil {
		return err
	}
	a.channel = channel

	message, err := envelope.Message()
	if err != nil {
		return err
	}
	a.message.UnmarshalBinary(message)

	if envelope.HasReplyTo() {
		replyTo, err := envelope.ReplyTo()
		if err != nil {
			return err
		}
		channelAddr, err := NewChannelAddress(replyTo)
		if err != nil {
			return err
		}
		a.replyTo = channelAddr
	}

	return nil
}

// MarshalBinary implements encoding.BinaryMarshaler
func (a *Envelope) MarshalBinary() ([]byte, error) {
	msgData, err := a.message.MarshalBinary()
	if err != nil {
		return nil, err
	}

	msg, seg, err := capnp.NewMessage(capnp.MultiSegment(nil))
	if err != nil {
		return nil, err
	}
	envelope, err := msgs.NewRootEnvelope(seg)
	if err != nil {
		return nil, err
	}
	if err = envelope.SetId(a.id); err != nil {
		return nil, err
	}
	envelope.SetCreated(a.created.UnixNano())
	if err = envelope.SetChannel(a.channel); err != nil {
		return nil, err
	}

	if err = envelope.SetChannel(a.channel); err != nil {
		return nil, err
	}
	if err = envelope.SetMessage(msgData); err != nil {
		return nil, err
	}

	if a.replyTo != nil {
		channelAddress, err := msgs.NewChannelAddress(seg)
		if err != nil {
			return nil, err
		}
		if err := a.replyTo.Write(channelAddress); err != nil {
			return nil, err
		}
		if err = envelope.SetReplyTo(channelAddress); err != nil {
			return nil, err
		}
	}

	buf := new(bytes.Buffer)
	compressor := zlib.NewWriter(buf)
	encoder := capnp.NewPackedEncoder(compressor)
	if err := encoder.Encode(msg); err != nil {
		return nil, err
	}
	compressor.Close()

	return buf.Bytes(), nil
}

func (a *Envelope) String() string {
	type envelope struct {
		Id      string
		Created time.Time

		Channel string
		Message Message

		ReplyTo *ChannelAddress
	}

	f := func() *envelope {
		return &envelope{
			a.id,
			a.created,
			a.channel,
			a.message,
			a.replyTo,
		}
	}

	json, _ := jsoniter.Marshal(f())
	return string(json)
}

var (
	EMPTY = &Empty{}
)

type Empty struct{}

func (a *Empty) UnmarshalBinary(data []byte) error {
	return nil
}

func (a *Empty) MarshalBinary() (data []byte, err error) {
	return []byte{}, nil
}
