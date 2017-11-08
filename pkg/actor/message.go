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

	"reflect"

	"github.com/json-iterator/go"
	"github.com/oysterpack/oysterpack.go/pkg/actor/msgs"
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

// Int64 provides a conversion back to int64
func (a MessageType) UInt64() uint64 {
	return uint64(a)
}

// ValidateMessageProcessor checks that the zero value is not set.
// The zero value is not allowed because the MessageType is always required, and we want to make sure it is set.
// By not allowing the zero value, then we are able to verify that it has been specified.
func (a MessageType) Validate() error {
	if a == 0 {
		return ErrInvalidMessageType
	}
	return nil
}

// Message that is sent between actors.
//
// All messages must know how marshal themselves to a binary format.
// All messages are delivered within an Envelope. The envelope will compress the entire envelope, thus there is no need
// to perform compression in this layer.
type Message interface {
	MessageType() MessageType

	encoding.BinaryMarshaler
	encoding.BinaryUnmarshaler
}

// NewEnvelope creates a new Envelope wrapping the specified message
// 	- uid is used to generate the envelope message id
//
// the method panics if the envelope is invalid
func NewEnvelope(uid UID, address Address, msg Message, replyTo *ReplyTo, correlationId *string) *Envelope {
	envelope := &Envelope{
		id:            uid(),
		created:       time.Now(),
		address:       address,
		message:       msg,
		msgType:       msg.MessageType(),
		replyTo:       replyTo,
		correlationId: correlationId,
	}
	return envelope
}

// EmptyEnvelope creates a new empty Envelope that can be used to unmarshal binary message envelopes
func EmptyEnvelope(emptyMessage Message) *Envelope {
	return &Envelope{message: emptyMessage, msgType: emptyMessage.MessageType()}
}

// Envelope is a message envelope. Envelope is itself a message.
type Envelope struct {
	id      string
	created time.Time

	address Address
	msgType MessageType
	message Message

	// optional
	replyTo       *ReplyTo
	correlationId *string
}

type ReplyTo struct {
	Address
	MessageType
}

func (a *ReplyTo) Validate() error {
	if a == nil {
		return nil
	}
	if err := a.Address.Validate(); err != nil {
		return err
	}
	if err := a.MessageType.Validate(); err != nil {
		return err
	}
	return nil
}

func (a *Envelope) Validate() error {
	if a.id == "" {
		return ErrEnvelopeIdBlank
	}

	if a.created.IsZero() {
		return ErrEnvelopeCreatedRquired
	}

	if err := a.address.Validate(); err != nil {
		return err
	}

	if a.message == nil {
		return ErrEnvelopeMessageRequired
	}

	if err := a.msgType.Validate(); err != nil {
		return err
	}

	if a.message.MessageType() != a.msgType {
		return envelopeMessageTypeDoesNotMatch(a)
	}

	if a.replyTo != nil {
		if err := a.replyTo.Validate(); err != nil {
			return err
		}
	}
	if a.correlationId != nil && *a.correlationId == "" {
		return ErrEnvelopeCorrelationIdBlank
	}

	return nil
}

func (a *Envelope) Id() string {
	return a.id
}

func (a *Envelope) Created() time.Time {
	return a.created
}

func (a *Envelope) Address() Address {
	return a.address
}

func (a *Envelope) MessageType() MessageType {
	return a.msgType
}

func (a *Envelope) Message() Message {
	return a.message
}

func (a *Envelope) ReplyTo() *ReplyTo {
	return a.replyTo
}

func (a *Envelope) CorrelationId() *string {
	return a.correlationId
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
	if a.created.IsZero() {
		return ErrEnvelopeCreatedRquired
	}

	if !envelope.HasAddress() {
		return ErrEnvelopeAddressRequired
	}
	capnAddr, err := envelope.Address()
	if err != nil {
		return err
	}
	addr, err := unmarshalCapnAddress(capnAddr)
	if err != nil {
		return err
	}
	a.address = *addr
	a.msgType = MessageType(envelope.MessageType())
	if err := a.msgType.Validate(); err != nil {
		return err
	}

	if envelope.HasReplyTo() {
		capnReplyTo, err := envelope.ReplyTo()
		if err != nil {
			return err
		}
		capnAddr, err := capnReplyTo.Address()
		if err != nil {
			return err
		}
		a.replyTo = &ReplyTo{}
		addr, err = unmarshalCapnAddress(capnAddr)
		if err != nil {
			return err
		}
		a.replyTo.Address = *addr
		a.replyTo.MessageType = MessageType(envelope.MessageType())
		if err := a.replyTo.Validate(); err != nil {
			return err
		}
	}

	correlationId, err := envelope.CorrelationId()
	if err != nil {
		return err
	}
	a.correlationId = &correlationId

	message, err := envelope.Message()
	if err != nil {
		return err
	}
	a.message.UnmarshalBinary(message)

	if a.msgType != a.message.MessageType() {
		return envelopeMessageTypeDoesNotMatch(a)
	}

	return nil
}

func unmarshalCapnAddress(addr msgs.Address) (*Address, error) {
	address := &Address{}
	if addr.HasPath() {
		path, err := addr.Path()
		if err != nil {
			return nil, err
		}
		address.Path = path
	}

	if addr.HasId() {
		id, err := addr.Id()
		if err != nil {
			return nil, err
		}
		address.Id = &id
	}
	if err := address.Validate(); err != nil {
		return nil, err
	}
	return address, nil
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
	addr, err := a.address.ToCapnpMessage(seg)
	if err != nil {
		return nil, err
	}
	if err := envelope.SetAddress(addr); err != nil {
		return nil, err
	}

	envelope.SetMessageType(a.Message().MessageType().UInt64())

	if err = envelope.SetMessage(msgData); err != nil {
		return nil, err
	}

	if a.replyTo != nil {
		replyTo, err := msgs.NewEnvelope_ReplyTo(seg)
		if err != nil {
			return nil, err
		}
		addr, err := a.replyTo.Address.ToCapnpMessage(seg)
		if err != nil {
			return nil, err
		}
		if err := replyTo.SetAddress(addr); err != nil {
			return nil, err
		}
		replyTo.SetMessageType(a.replyTo.MessageType.UInt64())

		if err = envelope.SetReplyTo(replyTo); err != nil {
			return nil, err
		}
	}

	if a.correlationId != nil {
		if err := envelope.SetCorrelationId(*a.correlationId); err != nil {
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
	type addr struct {
		Path string
		Id   *string
	}

	newAddr := func(a Address) addr {
		return addr{
			a.Path,
			a.Id,
		}
	}

	type replyTo struct {
		Address addr
		MessageType
	}

	newReplyTo := func(r *ReplyTo) *replyTo {
		if r == nil {
			return nil
		}
		return &replyTo{
			newAddr(r.Address),
			r.MessageType,
		}
	}

	type envelope struct {
		Id      string
		Created time.Time

		Address addr

		MessageGoType string

		ReplyTo       *replyTo
		CorrelationId *string
	}

	f := func() *envelope {
		return &envelope{
			a.id,
			a.created,
			newAddr(a.address),
			reflect.TypeOf(a.message).String(),
			newReplyTo(a.replyTo),
			a.correlationId,
		}
	}

	json, _ := jsoniter.Marshal(f())
	return string(json)
}
