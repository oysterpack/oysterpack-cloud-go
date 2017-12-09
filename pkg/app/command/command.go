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

package command

import (
	"context"

	"github.com/oysterpack/oysterpack.go/pkg/app"
)

type Command struct {
	id  CommandID
	run func(ctx context.Context) (context.Context, error)
}

func (a Command) CommandID() CommandID {
	return a.id
}

func (a Command) Run(ctx context.Context) (context.Context, error) {
	return a.run(ctx)
}

// ContextKey represents a global namespace for command keys.
type ContextKey uint64

const (
	CTX_KEY_CMD_ERR = ContextKey(0)
)

func Error(ctx context.Context) (err *app.Error, ok bool) {
	err, ok = ctx.Value(CTX_KEY_CMD_ERR).(*app.Error)
	return
}

func SetError(ctx context.Context, commandID CommandID, err *app.Error) context.Context {
	return context.WithValue(ctx, CTX_KEY_CMD_ERR, err.WithTag(commandID.Hex()))
}
