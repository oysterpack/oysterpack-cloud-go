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
)

type CommandAKey uint64

type CommandBKey uint64

// this test proves that type is used as part of the key
func TestContextKeyType(t *testing.T) {
	// Given that 2 keys have the same underlying value but different Go types
	const (
		A1 = CommandAKey(1)
		B1 = CommandBKey(1)
	)
	// When a value is stored with key A!
	ctx := context.WithValue(context.Background(), A1, "a")
	t.Logf("A1 -> %v, B1 -> %v", ctx.Value(A1), ctx.Value(B1))
	// Then the value cannot be retrieved with key B1
	if ctx.Value(B1) != nil {
		t.Error("We were able to retrieve the A1 value with the B1 key")
	}
	// And it can be retrieved with key A1
	if ctx.Value(A1) == nil {
		t.Error("We were not able to retrieve the A1 value with the A1 key")
	}

}

func TestCommandErr(t *testing.T) {
	ctx := context.Background()
	if _, ok := command.Error(ctx); ok {
		t.Error("There should be no error returned")
	}
	ctx = command.SetError(ctx, command.CommandID(1), app.AppNotAliveError(app.ServiceID(1)))
	if err, ok := command.Error(ctx); !ok {
		t.Error("There should be an error returned")
	} else {
		var err2 = err
		t.Log(err2)
		if !err.HasTag(command.CommandID(1).Hex()) {
			t.Errorf("CommandError CommandID does not match : %v", err.Tags)
		}
		if !app.IsError(err, app.ErrSpec_AppNotAlive.ErrorID) {
			t.Errorf("app.Err did not match : %v", err)
		}
	}
}
