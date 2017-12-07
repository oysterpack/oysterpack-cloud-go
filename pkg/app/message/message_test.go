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

	"encoding/json"

	"encoding/gob"

	"bytes"

	"github.com/json-iterator/go"
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

// BenchmarkMarshalling/json.Marshal-8              1000000              2422 ns/op             688 B/op          4 allocs/op
// BenchmarkMarshalling/json.Unmarshal-8            1000000              2244 ns/op             392 B/op          4 allocs/op
//
// BenchmarkMarshalling/jsoniter.Marshal-8          1000000              1490 ns/op             304 B/op          3 allocs/op
// BenchmarkMarshalling/jsoniter.Unmarshal-8        3000000               378 ns/op              96 B/op          2 allocs/op
//
// BenchmarkMarshalling/gob_Encoder-8               1000000              1002 ns/op             179 B/op          1 allocs/op
// BenchmarkMarshalling/gob_Decoder-8               2000000               643 ns/op               0 B/op          0 allocs/op
//
// BenchmarkMarshalling/capnp_Marshal-8            10000000               176 ns/op              84 B/op          2 allocs/op
// BenchmarkMarshalling/capnp_Unmarshal-8          10000000               209 ns/op             208 B/op          3 allocs/op
//
// capnp is the winner
func BenchmarkMarshalling(b *testing.B) {
	type MessageJSON struct {
		ID            uint64
		Type          uint64
		CorrelationID uint64
		Timestamp     int64

		Compression int16
		Packed      bool

		Data []byte

		TimeoutMSec int16
		ExpiresOn   int64
	}

	msg := MessageJSON{
		ID:            uid.NextUIDHash().Uint64(),
		Type:          uid.NextUIDHash().Uint64(),
		CorrelationID: uid.NextUIDHash().Uint64(),
		Timestamp:     time.Now().UnixNano(),
		Compression:   int16(message.Message_Compression_none),
		Packed:        false,
		TimeoutMSec:   500,
	}

	jsonBytes, err := jsoniter.Marshal(msg)
	if err != nil {
		b.Fatal(err)
	}
	b.Log(string(jsonBytes))

	b.Run("json.Marshal", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			json.Marshal(msg)
		}
	})

	var msg2 MessageJSON
	b.Run("json.Unmarshal", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			json.Unmarshal(jsonBytes, msg2)
		}
	})

	b.Run("jsoniter.Marshal", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			jsoniter.Marshal(msg)
		}
	})

	b.Run("jsoniter.Unmarshal", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			jsoniter.Unmarshal(jsonBytes, msg2)
		}
	})

	b.Run("gob Encoder", func(b *testing.B) {
		var network bytes.Buffer
		enc := gob.NewEncoder(&network)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			err := enc.Encode(msg)
			if err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("gob Decoder", func(b *testing.B) {
		var network bytes.Buffer
		enc := gob.NewEncoder(&network)
		for i := 0; i < b.N; i++ {
			enc.Encode(msg)
		}
		dec := gob.NewDecoder(&network)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			err := dec.Decode(&msg2)
			if err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("capnp Marshal", func(b *testing.B) {
		capnpMsg, msg, err := newMessage()
		if err != nil {
			b.Fatal(err)
		}
		capnpMsg.TraverseLimit = 64 << 40
		msg.SetCompression(message.Message_Compression_none)
		msg.SetPacked(false)

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, err := capnpMsg.Marshal()
			if err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("capnp Unmarshal", func(b *testing.B) {
		capnpMsg, msg, err := newMessage()
		if err != nil {
			b.Fatal(err)
		}
		capnpMsg.TraverseLimit = 64 << 40
		msg.SetCompression(message.Message_Compression_none)
		msg.SetPacked(false)

		bytes, err := capnpMsg.Marshal()
		if err != nil {
			b.Fatal(err)
		}
		if _, err := capnp.Unmarshal(bytes); err != nil {
			b.Fatal(err)
		}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, err := capnp.Unmarshal(bytes)
			if err != nil {
				b.Fatal(err)
			}
		}
	})

}
