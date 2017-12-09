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

import "fmt"

// CommandID is a global unique identifier for a command.
// CommandID(s) must be registered.
// The CommandID is used to track commands, e.g., metrics, errors
type CommandID uint64

func (a CommandID) Hex() string {
	return fmt.Sprintf("%x", a)
}

func (a CommandID) UInt64() uint64 {
	return uint64(a)
}

// PipelineID is a global unique identifier for a command
// PipelineID(s) must be registered
// The PipelineID is used to track pipelines, e.g., metrics, errors
type PipelineID uint64

func (a PipelineID) Hex() string {
	return fmt.Sprintf("%x", a)
}

func (a PipelineID) UInt64() uint64 {
	return uint64(a)
}
