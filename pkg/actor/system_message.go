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
	"bytes"
	"reflect"

	"errors"

	"github.com/oysterpack/oysterpack.go/pkg/actor/msgs"
	"zombiezen.com/go/capnproto2"
)

type SystemMessage interface {
	Message
	SystemMessage()
}

var (
	PING         = &Ping{EMPTY}
	PING_FACTORY = &PingMessageFactory{}

	PONG_FACTORY = &PongMessageFactory{}
)

type PingMessageFactory struct{}

func (a *PingMessageFactory) NewMessage() Message {
	return PING
}

func (a *PingMessageFactory) Validate(msg Message) error {
	if _, ok := msg.(*Ping); !ok {
		return &InvalidMessageTypeError{Message: msg, ExpectedType: reflect.TypeOf(PING)}
	}
	return nil
}

type Ping struct {
	*Empty
}

func (a *Ping) SystemMessage() {}

func (a *Ping) MessageType() MessageType { return MessageType(0) }

type PongMessageFactory struct{}

func (a *PongMessageFactory) NewMessage() Message {
	return &Pong{}
}

func (a *PongMessageFactory) Validate(msg Message) error {
	if msg, ok := msg.(*Pong); !ok {
		return &InvalidMessageTypeError{Message: msg, ExpectedType: reflect.TypeOf(&Pong{})}
	} else {
		if msg.Address == nil {
			return &InvalidMessageError{Message: msg, Err: errors.New("Address is required")}
		}
	}
	return nil
}

type Pong struct {
	*Address
}

func (a *Pong) SystemMessage() {}

func (a *Pong) MessageType() MessageType { return MessageType(1) }

func (a *Pong) UnmarshalBinary(data []byte) error {
	decoder := capnp.NewPackedDecoder(bytes.NewBuffer(data))
	msg, err := decoder.Decode()
	if err != nil {
		return err
	}
	addr, err := msgs.ReadRootAddress(msg)
	if err != nil {
		return err
	}
	a.Address, err = NewAddress(addr)
	return err
}

func (a *Pong) MarshalBinary() (data []byte, err error) {
	msg, seg, err := capnp.NewMessage(capnp.SingleSegment(nil))

	addr, err := msgs.NewRootAddress(seg)
	if err != nil {
		return nil, err
	}
	a.Write(addr)

	buf := new(bytes.Buffer)
	encoder := capnp.NewPackedEncoder(buf)
	if err := encoder.Encode(msg); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}
