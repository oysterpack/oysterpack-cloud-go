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
	"github.com/oysterpack/oysterpack.go/pkg/app"

	"context"
	"time"
)

// ErrSpec(s)
var (
	ErrSpec_ContextExpired = app.ErrSpec{app.ErrorID(0xd56f1203ea740414), app.ErrorType_KNOWN_EDGE_CASE, app.ErrorSeverity_MEDIUM}
)

// as a side effect, update pipeline metrics will be updated
func pipelineContextExpired(ctx context.Context, pipeline *Pipeline, commandID CommandID) *app.Error {
	pipeline.contextExpiredCounter.Inc()
	if IsPing(ctx) {
		pipeline.lastPingExpiredTime.Set(float64(time.Now().Unix()))
	} else {
		pipeline.lastExpiredTime.Set(float64(time.Now().Unix()))
		if c, ok := ExpiredOutputChannel(ctx); ok {
			select {
			case c <- ctx:
			default:
			}
		}
	}

	return app.NewError(ctx.Err(), "Context expired on Pipeline", ErrSpec_ContextExpired, pipeline.Service.ID(), commandID)
}
