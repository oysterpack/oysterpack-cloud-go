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

	"github.com/oysterpack/oysterpack.go/pkg/app"
)

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

// PipelineID is a type alias for ServiceID.
// A Pipeline is a Service, i.e., each Pipeline maps to a Service 1:1.
// PipelineID is simply a type alias for ServiceID - in order to provide more type safety.
type PipelineID app.ServiceID

// ServiceID converts the PipelineID back to ServiceID
func (a PipelineID) ServiceID() app.ServiceID {
	return app.ServiceID(a)
}
