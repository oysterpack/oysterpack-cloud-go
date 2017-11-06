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

package actor2

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
	replyTo := ctx.Message.replyTo
	if replyTo == nil {
		return nil
	}
	response := ctx.NewEnvelope(replyTo.Channel, SYS_MSG_PING_RESP, &PingResponse{ctx.Address()}, nil, ctx.Message.Id())

	if err := ctx.Send(response, replyTo.Address); err != nil {
		IgnoreActorNotFoundError(err)
	}
	return nil
}

// HandleHearbeatRequest will send back a HeartbeatResponse if a replyTo address was specified
func HandleHearbeatRequest(ctx *MessageContext) error {
	ctx.logger.Debug().Msg("HEARTBEAT")
	replyTo := ctx.Message.replyTo
	if replyTo == nil {
		return nil
	}
	response := ctx.NewEnvelope(replyTo.Channel, SYS_MSG_HEARTBEAT_RESP, HEARTBEAT_RESP, nil, ctx.Message.Id())
	if err := ctx.Send(response, replyTo.Address); err != nil {
		IgnoreActorNotFoundError(err)
	}

	return nil
}

func IgnoreActorNotFoundError(err error) error {
	switch err.(type) {
	case *ActorNotFoundError:
		return nil
	default:
		return err
	}
}
