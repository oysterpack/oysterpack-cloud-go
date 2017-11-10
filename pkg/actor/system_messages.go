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
	"github.com/oysterpack/oysterpack.go/pkg/actor/msgs"
	"github.com/rs/zerolog"
)

var (
	PING_REQUEST         = PingRequest{}
	PING_RESPONSE        = PingResponse{}
	PING_REQUEST_HANDLER = &MessageHandler{HandlePingRequest, Unmarshaller(PING_REQUEST)}
)

type PingRequest struct {
	EmptyMessage
}

func (a PingRequest) MessageType() MessageType {
	return MessageType(msgs.PingRequest_TypeID)
}

type PingResponse struct {
	EmptyMessage
}

func (a PingResponse) MessageType() MessageType {
	return MessageType(msgs.PingResponse_TypeID)
}

func HandlePingRequest(ctx *MessageContext) error {
	if replyTo := ctx.Message.ReplyTo(); replyTo != nil {
		ctx.Logger().Info().Dict("from", zerolog.Dict().Str("path", replyTo.Address.Path).Str("id", *replyTo.Address.Id)).Msg("PingRequest")
		if actor, ok := ctx.System().Actor(replyTo.Address); ok {
			response := ctx.NewEnvelope(PING_RESPONSE, replyTo.Address, nil, &ctx.Message.id)
			if err := actor.Tell(response); err != nil {
				ctx.Logger().Error().Err(err).Msg("Unable to send PingResponse")
			}
		} else {
			LOG_EVENT_ACTOR_NOT_FOUND.Log(ctx.Logger().Error()).Err(ActorNotFoundError{replyTo.Address}).Msg("Unable to send PingResponse")
		}
	} else {
		ctx.Logger().Info().Msg("PingRequest")
	}
	return nil
}
