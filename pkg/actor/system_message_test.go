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

package actor_test

import (
	"testing"
	"time"

	"strings"

	"github.com/oysterpack/oysterpack.go/pkg/actor"
)

func TestPong_MarshalBinary(t *testing.T) {
	now := time.Now()
	msg := actor.PONG_FACTORY.NewMessage().(*actor.Pong)
	msg.Address = &actor.Address{
		Path: []string{"oysterpack", "capnp"},
		Id:   uid(),
	}

	if err := actor.PONG_FACTORY.Validate(msg); err != nil {
		t.Fatalf("ping message was invalid : %v", err)
	}
	if err := actor.PING_FACTORY.Validate(msg); err == nil {
		t.Error("a pong is not a ping")
	}

	channel := actor.Channel("pong")

	envelope := actor.NewEnvelope(uid, channel, msg, nil)
	t.Log(envelope)

	if envelope.Id() == "" || envelope.Message() != msg ||
		envelope.Channel() != channel ||
		!envelope.Created().After(now) ||
		envelope.ReplyTo() != nil {
		t.Fatal("envelope fields are not matching")
	}

	envelopeBytes, err := envelope.MarshalBinary()
	if err != nil {
		t.Fatal(err)
	}

	envelope2 := actor.EmptyEnvelope(actor.PONG_FACTORY)
	if err := envelope2.UnmarshalBinary(envelopeBytes); err != nil {
		t.Fatal(err)
	}
	t.Log(envelope2)

	if envelope.Id() != envelope2.Id() {
		t.Errorf("id did not match : %s != %s", envelope.Id(), envelope.Id())
	}

	if envelope.Id() != envelope2.Id() || envelope.Channel() != envelope2.Channel() ||
		!envelope.Created().Equal(envelope2.Created()) ||
		envelope.Message().(*actor.Pong).Id != envelope2.Message().(*actor.Pong).Id ||
		strings.Join(envelope.Message().(*actor.Pong).Path, "/") != strings.Join(envelope2.Message().(*actor.Pong).Path, "/") {
		t.Fatal("envelope fields are not matching after unmarshalling")
	}
}
