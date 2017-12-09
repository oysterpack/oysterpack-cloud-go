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
