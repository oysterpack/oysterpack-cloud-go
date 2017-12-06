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

package message

import (
	"bytes"
	"compress/zlib"
	"io/ioutil"

	"zombiezen.com/go/capnproto2"
)

// Data returns the message data as a capnp.Message.
//
// errors:
//	- ERR_NODATA
//	- capnp errors
func Data(msg *Message) (*capnp.Message, error) {
	if msg == nil {
		return nil, ERR_MESSAGE_NIL
	}
	if !msg.HasData() {
		return nil, ERR_NODATA
	}

	data, err := msg.Data()
	if err != nil {
		return nil, err
	}
	switch msg.Compression() {
	case Message_Compression_none:
		return capnp.UnmarshalPacked(data)
	case Message_Compression_zlib:
		buf := bytes.NewBuffer(data)
		reader, err := zlib.NewReader(buf)
		if err != nil {
			return nil, err
		}
		data, err = ioutil.ReadAll(reader)
		if err != nil {
			return nil, err
		}
		return capnp.UnmarshalPacked(data)
	default:
		return nil, NewUnsupportedCompressionError(msg.Compression())
	}
}

// SetData will compress the data based on the message's compression setting
func SetData(msg *Message, data *capnp.Message) error {
	if msg == nil {
		return ERR_MESSAGE_NIL
	}
	dataBytes, err := data.MarshalPacked()
	if err != nil {
		return err
	}
	switch msg.Compression() {
	case Message_Compression_none:
		return msg.SetData(dataBytes)
	case Message_Compression_zlib:
		buf := new(bytes.Buffer)
		writer := zlib.NewWriter(buf)
		if _, err = writer.Write(dataBytes); err != nil {
			return err
		}
		if err := writer.Close(); err != nil {
			return err
		}
		return msg.SetData(buf.Bytes())
	default:
		return NewUnsupportedCompressionError(msg.Compression())
	}
}
