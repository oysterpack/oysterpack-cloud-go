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

const (
	CONTEXT_EXPIRED = app.LogEventID(0xb2c9b8df32d61bd3)
	CONTEXT_FAILED  = app.LogEventID(0xc82b54ad45672f0a)
)

func contextFailed(pipeline *Pipeline, ctx context.Context) {
	workflowID, _ := WorkflowID(ctx)
	CONTEXT_FAILED.Log(pipeline.Service.Logger().Warn()).Uint64("workflow", workflowID.UInt64()).Msg("context failed")
}

func contextExpired(pipeline *Pipeline, ctx context.Context) {
	workflowID, _ := WorkflowID(ctx)
	CONTEXT_EXPIRED.Log(pipeline.Service.Logger().Warn()).Uint64("workflow", workflowID.UInt64()).Msg("context expired")
}
