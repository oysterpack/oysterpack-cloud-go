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
// Normal component life cycle : STATE_NEW -> STATE_STARTING -> STATE_RUNNING -> STATE_STOPPING -> STATE_TERMINATED
// If the component fails while starting, running, or stopping, then it goes into state FAILED.
// The ordering of the State enum is defined such that if there is a state transition from A -> B then A < B.
const (
	// A service in this state is inactive. It does minimal work and consumes minimal resources.
	STATE_NEW State = iota
	// A service in this state is transitioning to RUNNING.
	STATE_STARTING
	// A service in this state is operational.
	STATE_RUNNING
	// A service in this state is transitioning to TERMINATED.
	STATE_STOPPING
	// A service in this state has completed execution normally. It does minimal work and consumes minimal resources.
	STATE_TERMINATED
	// A service in this state has encountered a problem and may not be operational. It cannot be started nor stopped.
	STATE_FAILED
)

// Terminal returns true if the state has reacjed a terminal state, i.e., terminated or failed state.
func (a State) Terminal() bool {
	return a == STATE_TERMINATED || a == STATE_FAILED
}

// New indicates a new state
func (a State) New() bool {
	return a == STATE_NEW
}

// Running indicates a running state
func (a State) Running() bool {
	return a == STATE_RUNNING
}

// Starting indicates a starting state
func (a State) Starting() bool {
	return a == STATE_STARTING
}

// Stoping indicates a stopping state
func (a State) Stopping() bool {
	return a == STATE_STOPPING
}

// Terminated indicates a terminated state
func (a State) Terminated() bool {
	return a == STATE_TERMINATED
}

// Failed indicates a failed state
func (a State) Failed() bool {
	return a == STATE_FAILED
}

func (a State) RunningOrLater() bool {
	return a >= STATE_RUNNING
}
