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
)

// CommandFunc will process the input context and return an outout Context.
// Commands that fail should return a Context with an *app.Error - see WithError()
type CommandFunc func(ctx context.Context) context.Context

func NewCommand(id CommandID, f CommandFunc) Command {
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
	run CommandFunc
}

func (a Command) CommandID() CommandID {
	return a.id
}

func (a Command) Run(ctx context.Context) context.Context {
	return a.run(ctx)
}
