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
	"time"

	"encoding"

	"bytes"
	"compress/zlib"

	"strings"

	"errors"

	"reflect"

	"github.com/json-iterator/go"
	"github.com/oysterpack/oysterpack.go/pkg/actor2/msgs"
	"zombiezen.com/go/capnproto2"
)

type MessageType uint8

func (a MessageType) UInt8() uint8 {
	return uint8(a)
}

const MESSAGE_TYPE_DEFAULT MessageType = 0

// Message that is sent between actors.
//
// All messages must know how marshal themselves to a binary format.
// All messages are delivered within an Envelope. The envelope will compress the entire envelope, thus there is no need
// to perform compression in this layer.
type Message interface {
	encoding.BinaryMarshaler
	encoding.BinaryUnmarshaler
}

// NewEnvelope creates a new Envelope wrapping the specified message
// 	- uid is used to generate the envelope message id
func NewEnvelope(uid UID, channel Channel, msgType MessageType, msg Message, replyTo *ChannelAddress, correlationId string) *Envelope {
	return &Envelope{
		id:            uid(),
		created:       time.Now(),
		channel:       channel,
		msgType:       msgType,
		message:       msg,
		replyTo:       replyTo,
		correlationId: correlationId,
	}
}

// UID is a function that returns a unique id.
type UID func() string

func EmptyEnvelope(emptyMessage Message) *Envelope {
	return &Envelope{message: emptyMessage}
}

// Envelope is a message envelope. Envelope is itself a message.
type Envelope struct {
	id      string
	created time.Time

	channel Channel
	msgType MessageType
	message Message

	replyTo *ChannelAddress

	correlationId string
}

func (a *Envelope) Validate() error {
	if strings.TrimSpace(a.id) == "" {
		return errors.New("Envelope.Id cannot be blank")
	}

	if a.created.IsZero() {
		return errors.New("Envelope.Created is not set")
	}

	if err := a.channel.Validate(); err != nil {
		return err
	}

	if a.message == nil {
		return errors.New("Envelope.Message is required")
	}

	if err := a.replyTo.Validate(); err != nil {
		return err
	}

	return nil
}

func (a *Envelope) CorrelationId() string {
	return a.id
}

func (a *Envelope) Id() string {
	return a.id
}

func (a *Envelope) Created() time.Time {
	return a.created
}

func (a *Envelope) Channel() Channel {
	return a.channel
}

func (a *Envelope) MessageType() MessageType {
	return a.msgType
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

	a.id, err = envelope.Id()
	if err != nil {
		return err
	}
	a.created = time.Unix(0, envelope.Created())

	channel, err := envelope.Channel()
	if err != nil {
		return err
	}
	a.channel = Channel(channel)

	a.msgType = MessageType(envelope.MessageType())

	a.correlationId, err = envelope.CorrelationId()
	if err != nil {
		return err
	}

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
	if err = envelope.SetChannel(a.channel.String()); err != nil {
		return nil, err
	}
	envelope.SetMessageType(a.MessageType().UInt8())
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

	if a.correlationId != "" {
		envelope.SetCorrelationId(a.correlationId)
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

		Channel     Channel
		MessageType MessageType
		GoType      string

		ReplyTo       *ChannelAddress
		CorrelationId string
	}

	f := func() *envelope {
		return &envelope{
			a.id,
			a.created,
			a.channel,
			a.MessageType(),
			reflect.TypeOf(a.message).String(),
			a.replyTo,
			a.correlationId,
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