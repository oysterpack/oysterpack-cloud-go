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

package service

import (
	"fmt"
	"sort"
)

// State is an enum representing the service lifecycle state
type State int

// State enum values
// Normal service life cycle : New -> Starting -> Running -> Stopping -> Terminated
// If the service fails while starting, running, or stopping, then it goes into state Service.State.FAILED.
// A stopped service may not be restarted.
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

// New returns true of the State is New
func (s State) New() bool { return s == New }

// Starting returns true of the State is Starting
func (s State) Starting() bool { return s == Starting }

// Running returns true of the State is Running
func (s State) Running() bool { return s == Running }

// Stopping returns true of the State is Stopping
func (s State) Stopping() bool { return s == Stopping }

// Terminated returns true of the State is Terminated
func (s State) Terminated() bool { return s == Terminated }

// Failed returns true of the State is Failed
func (s State) Failed() bool { return s == Failed }

// Stopped returns true if the serivce is Terminated or Failed
func (s State) Stopped() bool {
	return s == Terminated || s == Failed
}

// ValidTransitions returns the permitted State(s) that the current State is able to transition to
func (s State) ValidTransitions() (states States) {
	switch s {
	case New:
		states = []State{Starting, Terminated}
	case Starting:
		states = []State{Running, Stopping, Terminated, Failed}
	case Running:
		states = []State{Stopping, Terminated, Failed}
	case Stopping:
		states = []State{Terminated, Failed}
	case Terminated:
	case Failed:
	default:
		panic(fmt.Sprintf("Unknown State : %v", s))
	}
	return
}

// ValidTransition returns true is the state transition is permitted
func (s State) ValidTransition(to State) bool {
	for _, validState := range s.ValidTransitions() {
		if validState == to {
			return true
		}
	}
	return false
}

func (s State) String() string {
	switch s {
	case New:
		return "New"
	case Starting:
		return "Starting"
	case Running:
		return "Running"
	case Stopping:
		return "Stopping"
	case Terminated:
		return "Terminated"
	case Failed:
		return "Failed"
	default:
		panic(fmt.Sprintf("UNKNOWN STATE : %d", s))
	}
}

// States implements sort.Interface.
// When sorted, states will be sorted in by the state int value in increasing order, which represents the service lifecycle order.
type States []State

func (a States) Len() int           { return len(a) }
func (a States) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a States) Less(i, j int) bool { return a[i] < a[j] }

// AllStates contains all possible states in sorted order, i.e. from New to Failed
var AllStates States = []State{New, Starting, Running, Stopping, Terminated, Failed}

// Equals returns true if the to State slices contain the same set of State(s) regardless of order
func (a States) Equals(b States) bool {
	if a == nil && b == nil {
		return true
	}

	if len(a) != len(b) {
		return false
	}

	sort.Sort(a)
	sort.Sort(b)

	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}

	return true
}
