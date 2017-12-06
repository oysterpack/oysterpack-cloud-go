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

	"github.com/oysterpack/oysterpack.go/pkg/app"
	"github.com/oysterpack/oysterpack.go/pkg/app/message"
)

type ConnHandler func(ctx context.Context, conn net.Conn)

// ContextKey represents an enum for ConnHandler Context keys
type ContextKey uint

// ContextKey enum values
const (
	CTX_SERVER_SPEC = iota
	CTX_SERVICE
)

var (
	Context = ConnHandlerContext{}
)

type ConnHandlerContext struct{}

func (a ConnHandlerContext) Service(ctx context.Context) *app.Service {
	return ctx.Value(CTX_SERVICE).(*app.Service)
}

func (a ConnHandlerContext) ServerSpec(ctx context.Context) *ServerSpec {
	return ctx.Value(CTX_SERVER_SPEC).(*ServerSpec)
}

type MessageHandler func(ctx context.Context, conn net.Conn, msg message.Message)

type MessageType uint64

type MessageContext struct {
	ctx  context.Context
	conn net.Conn
	msg  message.Message
}

type MessageRoute struct {
	c chan *MessageContext
}

type MessageRouter struct {
	//messageHandlers map[MessageType]
}
