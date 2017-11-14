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

	"errors"

	"github.com/oysterpack/oysterpack.go/pkg/app"
	"github.com/oysterpack/oysterpack.go/pkg/app/messages/capnpmsgs"
	"zombiezen.com/go/capnproto2"
)

type MessageProcessingError struct {
	AppID            app.AppID
	AppInstanceID    app.InstanceID
	RemoteFunctionID app.RemoteFunctionID

	MessageID   MessageID
	MessageType MessageType

	Err *app.Err
}

// MarshalBinary implements encoding.BinaryMarshaler
func (a *MessageProcessingError) MarshalBinary() ([]byte, error) {
	msg, seg, err := capnp.NewMessage(capnp.SingleSegment(nil))
	if err != nil {
		return nil, err
	}

	capnpMsg, err := capnpmsgs.NewRootMessageProcessingError(seg)
	if err != nil {
		return nil, err
	}
	capnpMsg.SetAppId(uint64(a.AppID))
	if err := capnpMsg.SetAppInstanceId(string(a.AppInstanceID)); err != nil {
		return nil, err
	}
	capnpMsg.SetFunctionId(uint64(a.RemoteFunctionID))
	capnpMsg.SetMessageId(string(a.MessageID))
	capnpMsg.SetMessageType(uint64(a.MessageType))

	capnpMsg.SetErrCode(uint64(a.Err.ErrorID))
	if err := capnpMsg.SetErrMessage(a.Err.Error()); err != nil {
		return nil, err
	}

	buf := new(bytes.Buffer)
	encoder := capnp.NewPackedEncoder(buf)
	if err := encoder.Encode(msg); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// UnmarshalBinary implements encoding.BinaryUnmarshaler
func (a *MessageProcessingError) UnmarshalBinary(data []byte) error {
	decoder := capnp.NewPackedDecoder(bytes.NewBuffer(data))
	msg, err := decoder.Decode()
	if err != nil {
		return err
	}
	capnpMsg, err := capnpmsgs.ReadRootMessageProcessingError(msg)
	if err != nil {
		return err
	}
	a.AppID = app.AppID(capnpMsg.AppId())
	instanceId, err := capnpMsg.AppInstanceId()
	if err != nil {
		return err
	}
	a.AppInstanceID = app.InstanceID(instanceId)
	a.RemoteFunctionID = app.RemoteFunctionID(capnpMsg.FunctionId())

	msgId, err := capnpMsg.AppInstanceId()
	if err != nil {
		return err
	}
	a.MessageID = MessageID(msgId)
	a.MessageType = MessageType(capnpMsg.MessageType())

	errMsg, err := capnpMsg.ErrMessage()
	if err != nil {
		return err
	}
	a.Err = &app.Err{app.ErrorID(capnpMsg.ErrCode()), errors.New(errMsg)}
	return nil
}
