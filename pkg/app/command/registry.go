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
	"fmt"
	"sync"
)

var (
	commandsMutex sync.RWMutex
	commands      = make(map[CommandID]CommandFunc)
)

// RegisterCommand will register a command.
// A panic is triggered if a command is already registered using the same CommandID
func MustRegisterCommand(id CommandID, f CommandFunc) {
	if id == CommandID(0) {
		panic("CommandID(0) is illegal.")
	}
	if f == nil {
		panic("The CommandFunc must not be nil")
	}
	commandsMutex.Lock()
	defer commandsMutex.Unlock()
	if _, ok := commands[id]; ok {
		panic(fmt.Sprintf("A command is already registered for CommandID(0x%x", id))
	}
	commands[id] = f
}

func GetCommand(id CommandID) (f CommandFunc, ok bool) {
	commandsMutex.RLock()
	defer commandsMutex.RUnlock()
	f, ok = commands[id]
	return
}

func RegisteredCommandIDs() []CommandID {
	commandsMutex.RLock()
	defer commandsMutex.RUnlock()
	ids := make([]CommandID, len(commands))
	i := 0
	for id := range commands {
		ids[i] = id
		i++
	}
	return ids
}

// ClearCommandRegistry is only exposed for testing purposes
func ClearCommandRegistry() {
	commandsMutex.Lock()
	defer commandsMutex.Unlock()
	commands = make(map[CommandID]CommandFunc)
}
