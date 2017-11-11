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

package messages

import (
	"bytes"
	"compress/zlib"
	"encoding"
	"time"

	"github.com/oysterpack/oysterpack.go/pkg/app/messages/capnpmsgs"
	"zombiezen.com/go/capnproto2"
)

// MessageType is the unique identifier for a message type.
// MessageType provides a relatively short yet unambiguous way to refer to a type from another context.
//
// Most systems prefer instead to define a symbolic global namespace , but this would have some important disadvantages:
//	1. Programmers often feel the need to change symbolic names and organization in order to make their code cleaner,
//     but the renamed code should still work with existing encoded data.
//	2. It’s easy for symbolic names to collide, and these collisions could be hard to detect in a large distributed
//     system with many different binaries.
//	3. Fully-qualified type names may be large and waste space when transmitted on the wire.
type MessageType uint64

func (a MessageType) Validate() error {
	if a == 0 {
		return ErrMessageTypeZero
	}
	return nil
}

// Message should be implemented by messages that are meant to be distributed, i.e., sent over the network written or persisted
//
// All messages must know how marshal themselves to a binary format.
// All messages are delivered within an Envelope. The envelope will compress the entire envelope, thus there is no need
// to perform compression in this layer.
type Message interface {
	// MessageType returns the message's type - it is normally used to decide what go type to cast it to
	MessageType() MessageType

	encoding.BinaryMarshaler
	encoding.BinaryUnmarshaler
}

type MessageID string

func (a MessageID) Validate() error {
	if a == "" {
		return ErrMessageIDBlank
	}
	return nil
}

type Address string

func (a Address) Validate() error {
	if a == "" {
		return ErrAddressBlank
	}
	return nil
}

type ReplyTo struct {
	Address
	MessageType
}

func (a ReplyTo) Validate() error {
	if err := a.Address.Validate(); err != nil {
		return err
	}
	if err := a.MessageType.Validate(); err != nil {
		return err
	}
	return nil
}

type Envelope struct {
	MessageID MessageID
	CreatedOn time.Time
	To        Address

	ReplyTo       *ReplyTo
	CorrelationId *MessageID

	MessageType MessageType
	Message     Message
}

func (a *Envelope) Validate() error {
	if err := a.MessageID.Validate(); err != nil {
		return err
	}

	if a.CreatedOn.IsZero() {
		return ErrEnvelopeCreatedOnZero
	}

	if err := a.To.Validate(); err != nil {
		return err
	}

	if a.Message == nil {
		return ErrEnvelopeMessageRequired
	}

	if err := a.MessageType.Validate(); err != nil {
		return err
	}

	if a.Message.MessageType() != a.MessageType {
		return ErrMessageTypeDoesNotMatch
	}

	if a.ReplyTo != nil {
		if err := a.ReplyTo.Validate(); err != nil {
			return err
		}
	}
	if a.CorrelationId != nil && *a.CorrelationId == "" {
		return ErrCorrelationIdBlank
	}

	return nil
}

// MarshalBinary implements encoding.BinaryMarshaler
func (a *Envelope) MarshalBinary() ([]byte, error) {
	msgData, err := a.Message.MarshalBinary()
	if err != nil {
		return nil, err
	}

	msg, seg, err := capnp.NewMessage(capnp.MultiSegment(nil))
	if err != nil {
		return nil, err
	}
	envelope, err := capnpmsgs.NewRootEnvelope(seg)
	if err != nil {
		return nil, err
	}
	if err = envelope.SetId(string(a.MessageID)); err != nil {
		return nil, err
	}
	envelope.SetCreated(a.CreatedOn.UnixNano())
	if err := envelope.SetTo(string(a.To)); err != nil {
		return nil, err
	}

	envelope.SetMessageType(uint64(a.Message.MessageType()))

	if err = envelope.SetMessage(msgData); err != nil {
		return nil, err
	}

	if a.ReplyTo != nil {
		replyTo, err := capnpmsgs.NewEnvelope_ReplyTo(seg)
		if err != nil {
			return nil, err
		}
		if err := replyTo.SetTo(string(a.ReplyTo.Address)); err != nil {
			return nil, err
		}
		replyTo.SetMessageType(uint64(a.ReplyTo.MessageType))

		if err = envelope.SetReplyTo(replyTo); err != nil {
			return nil, err
		}
	}

	if a.CorrelationId != nil {
		if err := envelope.SetCorrelationId(string(*a.CorrelationId)); err != nil {
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

// UnmarshalBinary implements encoding.BinaryUnmarshaler
func (a *Envelope) UnmarshalBinary(data []byte) error {
	if a.Message == nil {
		return ErrEnvelopeMessagePrototypeNil
	}

	decompressor, _ := zlib.NewReader(bytes.NewBuffer(data))
	decoder := capnp.NewPackedDecoder(decompressor)
	msg, err := decoder.Decode()
	if err != nil {
		return err
	}
	envelope, err := capnpmsgs.ReadRootEnvelope(msg)
	if err != nil {
		return err
	}

	messageId, err := envelope.Id()
	if err != nil {
		return err
	}
	a.MessageID = MessageID(messageId)

	a.CreatedOn = time.Unix(0, envelope.Created())
	if a.CreatedOn.IsZero() {
		return ErrEnvelopeCreatedOnZero
	}

	if !envelope.HasTo() {
		return ErrAddressBlank
	}
	to, err := envelope.To()
	if err != nil {
		return err
	}
	a.To = Address(to)
	a.MessageType = MessageType(envelope.MessageType())
	if err := a.MessageType.Validate(); err != nil {
		return err
	}

	if envelope.HasReplyTo() {
		capnReplyTo, err := envelope.ReplyTo()
		if err != nil {
			return err
		}
		to, err := capnReplyTo.To()
		if err != nil {
			return err
		}
		a.ReplyTo = &ReplyTo{Address(to), MessageType(envelope.MessageType())}
		if err := a.ReplyTo.Validate(); err != nil {
			return err
		}
	}

	correlationId, err := envelope.CorrelationId()
	if err != nil {
		return err
	}
	corrMsgId := MessageID(correlationId)
	a.CorrelationId = &corrMsgId

	message, err := envelope.Message()
	if err != nil {
		return err
	}
	a.Message.UnmarshalBinary(message)

	if a.MessageType != a.Message.MessageType() {
		return ErrMessageTypeDoesNotMatch
	}

	return nil
}
