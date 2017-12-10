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

var (
	zero_time = time.Unix(0, 0)
)

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

type ContextKey struct{}

// *app.Error
type ctx_cmd_err ContextKey

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

// time.Time
type ctx_pipeline_workflow_start_time ContextKey

// WorkflowStartTime returns the time when the context workflow is started
func WorkflowStartTime(ctx context.Context) time.Time {
	start, ok := ctx.Value(ctx_pipeline_workflow_start_time{}).(time.Time)
	if ok {
		return start
	}
	return zero_time
}

// startWorkflowTimer returns a new Context with the pipeline start time set to now
func startWorkflowTimer(ctx context.Context) context.Context {
	return context.WithValue(ctx, ctx_pipeline_workflow_start_time{}, time.Now())
}

// time.Time
type ctx_created_on ContextKey

// ContextCreatedOn returns the time when the Context was created - (key=ctx_created_on)
func ContextCreatedOn(ctx context.Context) time.Time {
	createdOn, ok := ctx.Value(ctx_created_on{}).(time.Time)
	if ok {
		return createdOn
	}
	return zero_time
}

// chan<- context.Context
type ctx_output_channel ContextKey

// OutputChannel is used to send the pipeline workflow's final result . If set, then the pipeline output channel will not be used
// for this workflow.
func OutputChannel(ctx context.Context) (c chan<- context.Context, ok bool) {
	c, ok = ctx.Value(ctx_output_channel{}).(chan<- context.Context)
	return
}

func WithOutputChannel(ctx context.Context, c chan<- context.Context) context.Context {
	return context.WithValue(ctx, ctx_output_channel{}, c)
}

// chan<- context.Context
type ctx_stage_output_channel ContextKey

// StageOutputChannel - if set, then Context will be delivered to this channel after each processing stage.
// If the intermediate context is not able to be put on the channel immediately, then the intermediate context will be dropped,
// i.e., not delivered to the stage output channel.
//
// NOTE: This will be ignored for ping-pong Context(s), i.e., this does not apply to ping-pong Context(s).
//
// uses cases:
//	- for debugging purposes
//	- to monitor progress and enable the workflow to be cancelled based on intermediate results
func StageOutputChannel(ctx context.Context) (c chan<- context.Context, ok bool) {
	c, ok = ctx.Value(ctx_stage_output_channel{}).(chan<- context.Context)
	return
}

// WithStageOutputChannel returns a new context with the specified channel added as a stage output channel
func WithStageOutputChannel(ctx context.Context, c chan<- context.Context) context.Context {
	return context.WithValue(ctx, ctx_stage_output_channel{}, c)
}

// chan<- context.Context
type ctx_expired_channel ContextKey

// ExpiredOutputChannel is used to deliver Context(s) that have expired on the pipeline.
// If the expired context is not able to be put on the channel immediately, then the expired context will be dropped,
// i.e., not delivered to the expired output channel.
func ExpiredOutputChannel(ctx context.Context) (c chan<- context.Context, ok bool) {
	c, ok = ctx.Value(ctx_expired_channel{}).(chan<- context.Context)
	return
}

// WithExpiredOutputChannel returns a new context with the specified channel added as an expired output channel
func WithExpiredOutputChannel(ctx context.Context, c chan<- context.Context) context.Context {
	return context.WithValue(ctx, ctx_expired_channel{}, c)
}

// CommandID
type ctx_stage_command_id ContextKey

// StageCommandID returns the last stage that processed the Context
// NOTE: This will not be set for ping-pong contexts.
func StageCommandID(ctx context.Context) (id CommandID, ok bool) {
	id, ok = ctx.Value(ctx_stage_command_id{}).(CommandID)
	return
}

func withStageCommandID(ctx context.Context, id CommandID) context.Context {
	return context.WithValue(ctx, ctx_stage_command_id{}, id)
}

// uid.UIDHash - used for tracking purposes.
type ctx_workflow_id ContextKey

func WorkflowID(ctx context.Context) (id uid.UIDHash, ok bool) {
	id, ok = ctx.Value(ctx_workflow_id{}).(uid.UIDHash)
	return
}

// WithWorkflowID will assign an id to the workflow, if it does not already have a workflow id.
// The workflow id can be assigned externally, e.g. if the Context was flowing through multiple pipelines
func WithWorkflowID(ctx context.Context) context.Context {
	return context.WithValue(ctx, ctx_workflow_id{}, uid.NextUIDHash())
}

// struct{} - used to mark the context as a ping-pong context
// ping-pong is used to test how long does it take to traverse the pipeline - command functions are not run
type ctx_ping ContextKey

// IsPong returns true if this is a Ping Context, which tells the pipeline to send it through the workflow bypassing commands.
func IsPing(ctx context.Context) bool {
	return ctx.Value(ctx_ping{}) != nil
}

// time.Time - when the pong occurred
type ctx_pong ContextKey

// withPong puts the pong time in the context
func withPong(ctx context.Context) context.Context {
	return context.WithValue(ctx, ctx_pong{}, time.Now())
}

// PongTime returns the time the pong occurred, i.e., it indicates that the Context made it all the way through the pipeline.
func PongTime(ctx context.Context) (t time.Time, ok bool) {
	t, ok = ctx.Value(ctx_pong{}).(time.Time)
	return
}
