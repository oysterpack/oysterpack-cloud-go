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
	"github.com/oysterpack/oysterpack.go/pkg/app/uid"
)

type ContextKey struct{}

// *app.Error
type ctx_cmd_err ContextKey

// time.Time
type ctx_pipeline_workflow_start_time ContextKey

// time.Time
type ctx_created_on ContextKey

// chan<- context.Context
type ctx_output_channel ContextKey

// uid.UIDHash - used for tracking purposes.
type ctx_workflow_id ContextKey

// ping-pong is used to test how long does it take to traverse the pipeline - command functions are not run

// struct{} - used to mark the context as a ping-pong context
type ctx_ping ContextKey

// time.Time - when the pong occurred
type ctx_pong ContextKey

var zero_time = time.Unix(0, 0)

// NewContext returns a new Context with the following values :
// 	- created on (time.Time)
//	- workflow id (uid.UIDHash)
func NewContext() context.Context {
	ctx := context.WithValue(context.Background(), ctx_created_on{}, time.Now())
	return WithWorkflowID(ctx)
}

// NewPingContext
func NewPingContext() context.Context {
	ctx := context.WithValue(context.Background(), ctx_created_on{}, time.Now())
	ctx = WithWorkflowID(ctx)
	return context.WithValue(ctx, ctx_ping{}, struct{}{})
}

// IsPong returns true if this is a Ping Context, which tells the pipeline to send it through the workflow bypassing commands.
func IsPing(ctx context.Context) bool {
	return ctx.Value(ctx_ping{}) != nil
}

// WithPong puts the pong time in the context
func WithPong(ctx context.Context) context.Context {
	return context.WithValue(ctx, ctx_pong{}, time.Now())
}

func PongTime(ctx context.Context) (t time.Time, ok bool) {
	t, ok = ctx.Value(ctx_pong{}).(time.Time)
	return
}

// Error is used to communicate back any application errors.
func Error(ctx context.Context) *app.Error {
	err := ctx.Value(ctx_cmd_err{})
	if err == nil {
		return nil
	}
	return ctx.Value(ctx_cmd_err{}).(*app.Error)
}

// WithError returns a new Context with the Error added. The command id (in Hex format) is added as a tag to the Error.
func WithError(ctx context.Context, commandID CommandID, err *app.Error) context.Context {
	return context.WithValue(ctx, ctx_cmd_err{}, err.WithTag(commandID.Hex()))
}

// WorkflowStartTime returns the time when the context workflow is started
func WorkflowStartTime(ctx context.Context) time.Time {
	start, ok := ctx.Value(ctx_pipeline_workflow_start_time{}).(time.Time)
	if ok {
		return start
	}
	return zero_time
}

// StartWorkflowTimer returns a new Context with the pipeline start time set to now
func StartWorkflowTimer(ctx context.Context) context.Context {
	return context.WithValue(ctx, ctx_pipeline_workflow_start_time{}, time.Now())
}

// ContextCreatedOn returns the time when the Context was created - (key=ctx_created_on)
func ContextCreatedOn(ctx context.Context) time.Time {
	createdOn, ok := ctx.Value(ctx_created_on{}).(time.Time)
	if ok {
		return createdOn
	}
	return zero_time
}

// OutputChannel is used to send the pipeline workflow's final result . If set, then the pipeline output channel will not be used
// for this workflow.
func OutputChannel(ctx context.Context) (c chan<- context.Context, ok bool) {
	c, ok = ctx.Value(ctx_output_channel{}).(chan<- context.Context)
	return
}

func WithOutputChannel(ctx context.Context, c chan<- context.Context) context.Context {
	return context.WithValue(ctx, ctx_output_channel{}, c)
}

func WorkflowID(ctx context.Context) (id uid.UIDHash, ok bool) {
	id, ok = ctx.Value(ctx_workflow_id{}).(uid.UIDHash)
	return
}

// WithWorkflowID will assign an id to the workflow, if it does not already have a workflow id.
// The workflow id can be assigned externally, e.g. if the Context was flowing through multiple pipelines
func WithWorkflowID(ctx context.Context) context.Context {
	return context.WithValue(ctx, ctx_workflow_id{}, uid.NextUIDHash())
}
