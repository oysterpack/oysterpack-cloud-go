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

var sysMsgProcessor = func() MessageProcessor {
	msgProcessor := &systemMessageProcessor{}
	msgProcessor.MessageHandlers = MessageHandlers{
		MessageChannelKey{CHANNEL_SYSTEM, SYS_MSG_HEARTBEAT_REQ}: MessageHandler{
			Receive:   HandleHearbeatRequest,
			Unmarshal: UnmarshalHeartbeatRequest,
		},
		MessageChannelKey{CHANNEL_SYSTEM, SYS_MSG_PING_REQ}: MessageHandler{
			Receive:   HandlePingRequest,
			Unmarshal: UnmarshalPingRequest,
		},
	}
	return msgProcessor
}()

type systemMessageProcessor struct {
	MessageHandlers
}

// HandlePingRequest returns a PingResponse if a replyTo address was specified on the request
func HandlePingRequest(ctx *MessageContext) error {
	replyTo := ctx.Envelope.replyTo
	if replyTo == nil {
		return nil
	}
	return ctx.Send(ctx.MessageEnvelope(replyTo.Channel, SYS_MSG_PING_RESP, &PingResponse{ctx.address}), replyTo.Address)
}

// HandleHearbeatRequest will send back a HeartbeatResponse if a replyTo address was specified
func HandleHearbeatRequest(ctx *MessageContext) error {
	replyTo := ctx.Envelope.replyTo
	if replyTo == nil {
		return nil
	}
	return ctx.Send(ctx.RequestEnvelope(replyTo.Channel, SYS_MSG_HEARTBEAT_RESP, HEARTBEAT_RESP, CHANNEL_SYSTEM), replyTo.Address)
}
