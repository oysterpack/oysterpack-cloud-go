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

	"strings"

	"time"

	"github.com/nats-io/nuid"
	"github.com/oysterpack/oysterpack.go/pkg/actor"
)

func uid() string { return nuid.Next() }

func TestEnvelope_MarshalBinary(t *testing.T) {
	now := time.Now()
	msg := actor.PING_FACTORY.NewMessage()
	if err := actor.PING_FACTORY.Validate(msg); err != nil {
		t.Fatalf("ping message was invalid : %v", err)
	}
	channel := "ping"
	replyTo := &actor.ChannelAddress{
		Channel: "pong",
		Address: &actor.Address{
			Path: []string{"oysterpack", "config"},
			Id:   nuid.Next(),
		}}

	envelope := actor.NewEnvelope(uid, channel, actor.MESSAGE_TYPE_PING, msg, replyTo)
	t.Log(envelope)

	if envelope.Id() == "" || envelope.Message() != msg ||
		envelope.Channel() != channel ||
		!envelope.Created().After(now) ||
		envelope.ReplyTo().Channel != replyTo.Channel ||
		envelope.ReplyTo().Id != replyTo.Id ||
		strings.Join(envelope.ReplyTo().Path, "/") != strings.Join(replyTo.Path, "/") {
		t.Fatal("envelope fields are not matching")
	}

	envelopeBytes, err := envelope.MarshalBinary()
	if err != nil {
		t.Fatal(err)
	}

	envelope2 := actor.EmptyEnvelope(actor.PING_FACTORY)
	if err := envelope2.UnmarshalBinary(envelopeBytes); err != nil {
		t.Fatal(err)
	}
	t.Log(envelope2)

	if envelope.Id() != envelope2.Id() {
		t.Errorf("id did not match : %s != %s", envelope.Id(), envelope.Id())
	}

	if envelope.Id() != envelope2.Id() || envelope.Message() != envelope2.Message() || envelope.Channel() != envelope2.Channel() ||
		!envelope.Created().Equal(envelope2.Created()) ||
		envelope.ReplyTo().Channel != envelope2.ReplyTo().Channel ||
		envelope.ReplyTo().Id != envelope2.ReplyTo().Id ||
		strings.Join(envelope.ReplyTo().Path, "/") != strings.Join(envelope2.ReplyTo().Path, "/") {
		t.Fatal("envelope fields are not matching after unmarshalling")
	}
}
