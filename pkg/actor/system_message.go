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

func PingMessageFactory() MessageFactory { return ping_msg_factory }

var (
	ping_msg         = &ping{EMPTY}
	ping_msg_factory = &pingMessageFactory{}

	pong_msg_factory = &pongMessageFactory{}
)

type pingMessageFactory struct{}

func (a *pingMessageFactory) NewMessage() Message {
	return ping_msg
}

func (a *pingMessageFactory) Validate(msg Message) error {
	if _, ok := msg.(*ping); !ok {
		return &InvalidMessageTypeError{Message: msg, ExpectedType: reflect.TypeOf(ping_msg)}
	}
	return nil
}

type ping struct {
	*Empty
}

func (a *ping) SystemMessage() {}

func PongMessageFactory() MessageFactory { return pong_msg_factory }

type pongMessageFactory struct{}

func (a *pongMessageFactory) NewMessage() Message {
	return &Pong{}
}

func (a *pongMessageFactory) Validate(msg Message) error {
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
