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

	"time"

	"github.com/oysterpack/oysterpack.go/pkg/app"
)

func NewCommand(id CommandID, f func(ctx context.Context) (context.Context, *app.Error)) Command {
	if id == CommandID(0) {
		panic("CommandID must not be 0")
	}
	if f == nil {
		panic("Command function must not be nil")
	}
	return Command{id, f}
}

type Command struct {
	id  CommandID
	run func(ctx context.Context) (context.Context, *app.Error)
}

func (a Command) CommandID() CommandID {
	return a.id
}

func (a Command) Run(ctx context.Context) (context.Context, *app.Error) {
	return a.run(ctx)
}

// ContextKey represents a global namespace for command keys.
type ContextKey uint64

// Standard ContextKey(s)
const (
	// *app.Error
	CTX_KEY_CMD_ERR = ContextKey(0)

	// time.Time
	CTX_KEY_PIPELINE_WORKFLOW_START_TIME = ContextKey(1)

	// time.Time
	CTX_KEY_CREATED_ON = ContextKey(2)
)

var ZeroTime = time.Unix(0, 0)

func Error(ctx context.Context) *app.Error {
	err := ctx.Value(CTX_KEY_CMD_ERR)
	if err == nil {
		return nil
	}
	return ctx.Value(CTX_KEY_CMD_ERR).(*app.Error)
}

func SetError(ctx context.Context, commandID CommandID, err *app.Error) context.Context {
	return context.WithValue(ctx, CTX_KEY_CMD_ERR, err.WithTag(commandID.Hex()))
}

// PipelineWorkflowStartTime returns the time when the context workflow is started
func PipelineWorkflowStartTime(ctx context.Context) time.Time {
	start, ok := ctx.Value(CTX_KEY_PIPELINE_WORKFLOW_START_TIME).(time.Time)
	if ok {
		return start
	}
	return ZeroTime
}

func PipelineWorkflowStarted(ctx context.Context) context.Context {
	return context.WithValue(ctx, CTX_KEY_PIPELINE_WORKFLOW_START_TIME, time.Now())
}

// NewContext returns a new Context with the created on time set (key=CTX_KEY_CREATED_ON)
func NewContext() context.Context {
	return context.WithValue(context.Background(), CTX_KEY_CREATED_ON, time.Now())
}

// ContextCreatedOn returns the time when the Context was created - (key=CTX_KEY_CREATED_ON)
func ContextCreatedOn(ctx context.Context) time.Time {
	createdOn, ok := ctx.Value(CTX_KEY_CREATED_ON).(time.Time)
	if ok {
		return createdOn
	}
	return ZeroTime
}
