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

	"github.com/oysterpack/oysterpack.go/pkg/actor/msgs"
	"zombiezen.com/go/capnproto2"
)

type SystemMessage interface {
	Message
	SystemMessage()
}

var (
	PING_REQ  = &PingRequest{EMPTY}
	HEARTBEAT = &Heartbeat{}
)

const (
	SYSTEM_MESSAGE_PING_REQ  MessageType = 0
	SYSTEM_MESSAGE_PING_RESP MessageType = 1

	SYSTEM_MESSAGE_FAILURE MessageType = 2

	SYSTEM_MESSAGE_HEARTBEAT MessageType = 3
)

type Heartbeat struct {
	*Empty
}

func (a *Heartbeat) SystemMessage() {}

func (a *Heartbeat) MessageType() MessageType { return SYSTEM_MESSAGE_HEARTBEAT }

type PingRequest struct {
	*Empty
}

func (a *PingRequest) SystemMessage() {}

func (a *PingRequest) MessageType() MessageType { return SYSTEM_MESSAGE_PING_REQ }

type PingResponse struct {
	*Address
}

func (a *PingResponse) SystemMessage() {}

func (a *PingResponse) MessageType() MessageType { return SYSTEM_MESSAGE_PING_RESP }

func (a *PingResponse) UnmarshalBinary(data []byte) error {
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

func (a *PingResponse) MarshalBinary() (data []byte, err error) {
	msg, seg, err := capnp.NewMessage(capnp.SingleSegment(nil))

	addr, err := msgs.NewRootAddress(seg)
	if err != nil {
		return nil, err
	}
	a.Write(addr)

	buf := new(bytes.Buffer)
	encoder := capnp.NewPackedEncoder(buf)
	if err = encoder.Encode(msg); err != nil {
		return
	}

	return buf.Bytes(), nil
}
