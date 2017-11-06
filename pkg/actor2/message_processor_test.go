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

package actor2_test

import (
	"testing"

	"errors"

	"github.com/oysterpack/oysterpack.go/pkg/actor2"
	"github.com/rs/zerolog/log"
)

func TestStartMessageProcessorEngine(t *testing.T) {
	foo := actor2.MessageHandlers{
		actor2.MessageChannelKey{actor2.CHANNEL_SYSTEM, actor2.MESSAGE_TYPE_DEFAULT}: actor2.MessageHandler{
			Receive: func(ctx *actor2.MessageContext) error {
				t.Logf("Received message: %v", ctx.Message)
				return nil
			},
			Unmarshal: func(msg []byte) (*actor2.Envelope, error) { return nil, errors.New("NOT SUPPORTED") },
		},
		actor2.MessageChannelKey{actor2.CHANNEL_LIFECYCLE, actor2.MESSAGE_TYPE_DEFAULT}: actor2.MessageHandler{
			Receive: func(ctx *actor2.MessageContext) error {
				t.Logf("Received message: %v", ctx.Message)
				return nil
			},
			Unmarshal: func(msg []byte) (*actor2.Envelope, error) { return nil, errors.New("NOT SUPPORTED") },
		},
	}

	processor, err := actor2.StartMessageProcessorEngine(foo, log.Logger)
	if err != nil {
		t.Fatal(err)
	}

	if len(processor.ChannelNames()) != 2 {
		t.Errorf("Channel count is wrong : %v", processor.ChannelNames())
	}

	if !processor.Alive() {
		t.Error("processor should be alive")
	}

	msg := actor2.PING_REQ
	processor.Channel() <- &actor2.MessageContext{
		Actor:   nil,
		Message: actor2.NewEnvelope(uid, actor2.CHANNEL_SYSTEM, actor2.SYS_MSG_PING_REQ, msg, nil, ""),
	}

	processor.Kill(nil)
	deathReason := processor.Wait()
	t.Logf("death reason : %v", deathReason)

}
