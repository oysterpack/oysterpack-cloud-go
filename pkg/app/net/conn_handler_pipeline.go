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

package net

import (
	"context"
	"net"

	"io"

	"time"

	"github.com/oysterpack/oysterpack.go/pkg/app"
	"github.com/oysterpack/oysterpack.go/pkg/app/command"
	"github.com/oysterpack/oysterpack.go/pkg/app/message"
	"zombiezen.com/go/capnproto2"
)

type ctx_request_message command.ContextKey
type ctx_response_message command.ContextKey

func RequestMessage(ctx context.Context) *message.Message {
	msg, ok := ctx.Value(ctx_request_message{}).(*message.Message)
	if ok {
		return msg
	}
	return nil
}

func WithRequestMessage(ctx context.Context, msg *message.Message) context.Context {
	return context.WithValue(ctx, ctx_request_message{}, msg)
}

func ResponseMessage(ctx context.Context) *capnp.Message {
	msg, ok := ctx.Value(ctx_response_message{}).(*capnp.Message)
	if ok {
		return msg
	}
	return nil
}

func WithResponseMessage(ctx context.Context, msg *capnp.Message) context.Context {
	return context.WithValue(ctx, ctx_response_message{}, msg)
}

func NewMessagePipelineConnHandler(pipelineID command.PipelineID) ConnHandler {
	pipeline := command.GetPipeline(pipelineID)
	if pipeline == nil {
		service := app.Services.Service(pipelineID.ServiceID())
		if service == nil {
			service = app.NewService(pipelineID.ServiceID())
			app.Services.Register(service)
		}
		pipeline = command.StartPipelineFromConfig(service)
	}

	service := pipeline.Service
	in := pipeline.InputChan()
	out := pipeline.OutputChan()

	return func(ctx context.Context, conn net.Conn) {
		defer conn.Close()

		decoder := capnp.NewPackedDecoder(conn)
		decoder.ReuseBuffer()

		service.Go(func() error {
			encoder := capnp.NewPackedEncoder(conn)
			for {
				select {
				case <-service.Dying():
					return nil
				case responseCtx := <-out:
					select {
					case <-responseCtx.Done():
						// context is expired - response will not be sent
						// NOTE: metrics and log events are recorded by the pipeline
					default:
						responseMsg := ResponseMessage(responseCtx)
						if responseMsg != nil {
							if err := encoder.Encode(responseMsg); err != nil {
								if service.Alive() {
									service.Logger().Error().Err(err).Msgf("encoder.Encode(responseMsg) failed : %T", err)
								}
								return err
							}
						}
					}

				}
			}
		})

		for {
			msg, err := decoder.Decode()
			if err != nil {
				if err == io.EOF {
					return
				}
				if service.Alive() || app.Alive() {
					service.Logger().Error().Err(err).Msgf("decoder.Decode() failed : %T", err)
				}
				return
			}
			request, err := message.ReadRootMessage(msg)
			if err != nil {
				service.Logger().Error().Err(err).Msg("message.ReadRootMessage(msg) failed")
				return
			}

			requestCtx := service.Context(WithRequestMessage(ctx, &request))

			switch request.Deadline().Which() {
			case message.Message_deadline_Which_timeoutMSec:
				requestCtx, _ = context.WithDeadline(requestCtx, time.Unix(0, request.Deadline().ExpiresOn()))
			case message.Message_deadline_Which_expiresOn:
				requestCtx, _ = context.WithTimeout(requestCtx, time.Millisecond*time.Duration(request.Deadline().TimeoutMSec()))
			default:
				service.Logger().Error().Msgf("Unexpected deadline type : %v", request.Deadline().Which())
			}

			select {
			case <-requestCtx.Done():
			case in <- requestCtx:
			}
		}
	}
}
