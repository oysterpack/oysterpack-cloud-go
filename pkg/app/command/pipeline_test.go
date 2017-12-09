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

package command_test

import (
	"context"
	"testing"

	"github.com/oysterpack/oysterpack.go/pkg/app"
	"github.com/oysterpack/oysterpack.go/pkg/app/command"
	"github.com/oysterpack/oysterpack.go/pkg/app/uid"
)

func TestStartPipeline(t *testing.T) {
	SERVICE_ID := app.ServiceID(uid.NextUIDHash())

	t.Run("single stage - no error", func(t *testing.T) {
		app.Reset()
		defer app.Reset()

		service := app.NewService(SERVICE_ID)

		type Key int

		const (
			A = Key(iota)
			B
			SUM
		)

		p, err := command.StartPipeline(service,
			command.NewStage(
				command.NewCommand(command.CommandID(1), func(ctx context.Context) (context.Context, *app.Error) {
					a := ctx.Value(A).(int)
					b := ctx.Value(B).(int)
					return context.WithValue(ctx, SUM, a+b), nil
				}),
				1,
			),
		)
		if err != nil {
			t.Fatal(err)
		}

		ctx := context.WithValue(context.Background(), A, 1)
		ctx = context.WithValue(ctx, B, 2)
		p.InputChan() <- ctx
		ctx = <-p.OutputChan()
		sum := ctx.Value(SUM).(int)
		t.Logf("sum = %d", sum)
		if sum != 3 {
			t.Errorf("The pipeline did not process the workflow correctly : sum = %d", sum)
		}
	})

	t.Run("10 stage pipeline", func(t *testing.T) {
		app.Reset()
		defer app.Reset()

		service := app.NewService(SERVICE_ID)
		type Key int

		const (
			N = Key(iota)
		)
		stage := command.NewStage(
			command.NewCommand(command.CommandID(1), func(ctx context.Context) (context.Context, *app.Error) {
				n := ctx.Value(N).(int)
				return context.WithValue(ctx, N, n+1), nil
			}),
			1,
		)

		stages := []command.Stage{}
		for i := 0; i < 10; i++ {
			stages = append(stages, stage)
		}
		p, err := command.StartPipeline(service, stages...)
		if err != nil {
			t.Fatal(err)
		}
		ctx := context.WithValue(context.Background(), N, 0)
		p.InputChan() <- ctx
		ctx = <-p.OutputChan()
		n := ctx.Value(N).(int)
		t.Logf("n = %d", n)
		if n != 10 {
			t.Errorf("The pipeline did not process the workflow correctly : n = %d", n)
		}
	})
}
