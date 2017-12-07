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

package message_test

import (
	"testing"
	"time"

	"github.com/oysterpack/oysterpack.go/pkg/app"
	"github.com/oysterpack/oysterpack.go/pkg/app/message"
	"github.com/oysterpack/oysterpack.go/pkg/app/uid"
	"zombiezen.com/go/capnproto2"
)

func TestData(t *testing.T) {
	app.Reset()

	t.Run("compression = zlib", func(t *testing.T) {
		_, msg, err := newMessage()
		if err != nil {
			t.Fatal(err)
		}
		msg.SetCompression(message.Message_Compression_zlib)

		capnpMsgData, msgData, err := newMessage()
		if err != nil {
			t.Fatal(err)
		}

		message.SetData(msg, capnpMsgData)

		unmarshalledMsg, err := message.Data(msg)
		if err != nil {
			t.Fatal(err)
		}
		msgData2, err := message.ReadRootMessage(unmarshalledMsg)
		if err != nil {
			t.Fatal(err)
		}
		if msgData.Id() != msgData2.Id() {
			t.Error("message ids do not match")
		}
	})

	t.Run("compression = none", func(t *testing.T) {
		_, msg, err := newMessage()
		if err != nil {
			t.Fatal(err)
		}

		capnpMsgData, msgData, err := newMessage()
		if err != nil {
			t.Fatal(err)
		}

		msg.SetCompression(message.Message_Compression_none)
		message.SetData(msg, capnpMsgData)

		unmarshalledMsg, err := message.Data(msg)
		if err != nil {
			t.Fatal(err)
		}
		msgData2, err := message.ReadRootMessage(unmarshalledMsg)
		if err != nil {
			t.Fatal(err)
		}
		if msgData.Id() != msgData2.Id() {
			t.Error("message ids do not match")
		}
	})

}

func newMessage() (*capnp.Message, *message.Message, error) {
	capnpMsg, seg, err := capnp.NewMessage(capnp.SingleSegment(nil))
	if err != nil {
		return nil, nil, err
	}
	msg, err := message.NewRootMessage(seg)
	if err != nil {
		return nil, nil, err
	}
	msg.SetId(uint64(uid.NextUIDHash()))
	msg.SetType(uint64(uid.NextUIDHash()))
	msg.SetCorrelationID(uint64(uid.NextUIDHash()))
	msg.SetTimestamp(time.Now().UnixNano())
	return capnpMsg, &msg, nil
}

func BenchmarkData(b *testing.B) {
	app.Reset()

	b.Run("compression = zlib, packed = false", func(b *testing.B) {
		capnpMsg, msg, err := newMessage()
		if err != nil {
			b.Fatal(err)
		}
		capnpMsg.TraverseLimit = 64 << 40
		msg.SetCompression(message.Message_Compression_zlib)
		msg.SetPacked(false)
		capnpMsgData, _, err := newMessage()
		if err != nil {
			b.Fatal(err)
		}
		message.SetData(msg, capnpMsgData)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			if _, err := message.Data(msg); err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("compression = zlib, packed = true", func(b *testing.B) {
		capnpMsg, msg, err := newMessage()
		if err != nil {
			b.Fatal(err)
		}
		capnpMsg.TraverseLimit = 64 << 40
		msg.SetCompression(message.Message_Compression_zlib)
		msg.SetPacked(true)
		capnpMsgData, _, err := newMessage()
		if err != nil {
			b.Fatal(err)
		}
		message.SetData(msg, capnpMsgData)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			if _, err := message.Data(msg); err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("compression = none, packed = false", func(b *testing.B) {
		capnpMsg, msg, err := newMessage()
		if err != nil {
			b.Fatal(err)
		}
		capnpMsg.TraverseLimit = 64 << 40
		b.Logf("capnpMsg.TraverseLimit = %v, default = %v", capnpMsg.TraverseLimit/1024/1024, (64<<20)/1024/1024)
		capnpMsgData, _, err := newMessage()
		if err != nil {
			b.Fatal(err)
		}
		msg.SetCompression(message.Message_Compression_none)
		msg.SetPacked(false)
		message.SetData(msg, capnpMsgData)

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			if _, err := message.Data(msg); err != nil {
				b.Fatal(err)
			}
		}

	})
	b.Run("compression = none, packed = true", func(b *testing.B) {
		capnpMsg, msg, err := newMessage()
		if err != nil {
			b.Fatal(err)
		}
		capnpMsg.TraverseLimit = 64 << 40
		b.Logf("capnpMsg.TraverseLimit = %v, default = %v", capnpMsg.TraverseLimit/1024/1024, (64<<20)/1024/1024)
		capnpMsgData, _, err := newMessage()
		if err != nil {
			b.Fatal(err)
		}
		msg.SetCompression(message.Message_Compression_none)
		msg.SetPacked(true)
		message.SetData(msg, capnpMsgData)

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			if _, err := message.Data(msg); err != nil {
				b.Fatal(err)
			}
		}

	})

}
