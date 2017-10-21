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

package comp

// State is an enum representing the service lifecycle state
type State int

// State enum values
// Normal component life cycle : New -> Starting -> Running -> Stopping -> Terminated
// If the component fails while starting, running, or stopping, then it goes into state FAILED.
// The ordering of the State enum is defined such that if there is a state transition from A -> B then A < B.
const (
	// A service in this state is inactive. It does minimal work and consumes minimal resources.
	New State = iota
	// A service in this state is transitioning to RUNNING.
	Starting
	// A service in this state is operational.
	Running
	// A service in this state is transitioning to TERMINATED.
	Stopping
	// A service in this state has completed execution normally. It does minimal work and consumes minimal resources.
	Terminated
	// A service in this state has encountered a problem and may not be operational. It cannot be started nor stopped.
	Failed
)
