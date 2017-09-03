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

package service_test

import (
	"github.com/oysterpack/oysterpack.go/oysterpack/service"
	"sort"
	"sync"
	"testing"
)



func TestState_New(t *testing.T) {
	for i := 0; i <= int(service.Failed); i++ {
		if state := service.State(i); state == service.New {
			if !state.New() {
				t.Error("State did not recognize itself")
			}
		} else {
			if state.New() {
				t.Errorf("%v is not New", state)
			}
		}
	}
}

func TestState_Starting(t *testing.T) {
	for i := 0; i <= int(service.Failed); i++ {
		if state := service.State(i); state == service.Starting {
			if !state.Starting() {
				t.Error("State did not recognize itself")
			}
		} else {
			if state.Starting() {
				t.Errorf("%v is not Starting", state)
			}
		}
	}
}

func TestState_Running(t *testing.T) {
	for i := 0; i <= int(service.Failed); i++ {
		if state := service.State(i); state == service.Running {
			if !state.Running() {
				t.Error("State did not recognize itself")
			}
		} else {
			if state.Running() {
				t.Errorf("%v is not Running", state)
			}
		}
	}
}

func TestState_Stopping(t *testing.T) {
	for i := 0; i <= int(service.Failed); i++ {
		if state := service.State(i); state == service.Stopping {
			if !state.Stopping() {
				t.Error("State did not recognize itself")
			}
		} else {
			if state.Stopping() {
				t.Errorf("%v is not Stopping", state)
			}
		}
	}
}

func TestState_Terminated(t *testing.T) {
	for i := 0; i <= int(service.Failed); i++ {
		if state := service.State(i); state == service.Terminated {
			if !state.Terminated() {
				t.Error("State did not recognize itself")
			}

			if !state.Stopped() {
				t.Error("State should be one of the Stopped stated")
			}
		} else {
			if state.Terminated() {
				t.Errorf("%v is not New", state)
			}
		}
	}
}

func TestState_Failed(t *testing.T) {
	for i := 0; i <= int(service.Failed); i++ {
		if state := service.State(i); state == service.Failed {
			if !state.Failed() {
				t.Error("State did not recognize itself")
			}

			if !state.Stopped() {
				t.Error("State should be one of the Stopped stated")
			}
		} else {
			if state.Failed() {
				t.Errorf("%v is not New", state)
			}
		}
	}
}

func TestState_ValidTransition(t *testing.T) {
	validTransitions := map[service.State]service.States{
		service.New:        {service.Starting, service.Terminated},
		service.Starting:   {service.Running, service.Stopping, service.Terminated, service.Failed},
		service.Running:    {service.Stopping, service.Terminated, service.Failed},
		service.Stopping:   {service.Terminated, service.Failed},
		service.Terminated: {},
		service.Failed:     {},
	}

	for state, expected := range validTransitions {
		actual := state.ValidTransitions()
		if !actual.Equals(expected) {
			sort.Sort(actual)
			sort.Sort(expected)
			t.Errorf("%v : actual:%v != expected:%v", state, actual, expected)
		}
	}

	for state, validTransitions := range validTransitions {
		if state.ValidTransition(state) {
			t.Errorf("A self transition is not a valid state : %v", state)
		}
		invalidState := service.State(service.Failed + 1)
		if state.ValidTransition(invalidState) {
			t.Errorf("%v -> %v should be an invalid transition",state,invalidState)
		}
		invalidState = service.State(service.New - 1)
		if state.ValidTransition(invalidState) {
			t.Errorf("%v -> %v should be an invalid transition",state,invalidState)
		}

		for _, to := range validTransitions {
			if !state.ValidTransition(to) {
				t.Errorf("%v -> %v should be a valid transition",state,to)
			}
		}
	}

	panicked := false
	wait := sync.WaitGroup{}
	wait.Add(1)
	go func() {
		defer func() {
			if p := recover(); p != nil {
				panicked = true
				logger.Info().Msgf("%v", p)
			}
			wait.Done()
		}()
		invalidState := service.State(-1)
		invalidState.ValidTransitions()
	}()
	wait.Wait()
	if !panicked {
		t.Errorf("State.ValidTransitions should panic for an invalid state")
	}
}

func TestStates_Equals(t *testing.T) {
	states1 := service.States{}
	states2 := service.States{}
	if !states1.Equals(states2) {
		t.Errorf("Should be equal: %v %v", states1, states2)
	}

	states1 = service.States{}
	states2 = nil
	if !states1.Equals(states2) {
		t.Errorf("Should be equal: %v %v", states1, states2)
	}

	states1 = nil
	states2 = nil
	if !states1.Equals(states2) {
		t.Errorf("Should be equal: %v %v", states1, states2)
	}

	states1 = service.States{service.New, service.Starting, service.Running}
	states2 = service.States{service.New, service.Starting, service.Running}
	if !states1.Equals(states2) {
		t.Errorf("Should be equal: %v %v", states1, states2)
	}

	states1 = service.States{service.New, service.Starting, service.Running}
	states2 = service.States{service.New, service.Starting}
	if states1.Equals(states2) {
		t.Errorf("Should not be equal: %v %v", states1, states2)
	}

	states1 = service.States{service.New, service.Starting, service.Running}
	states2 = service.States{service.New, service.Starting, service.Terminated}
	if states1.Equals(states2) {
		t.Errorf("Should not be equal: %v %v", states1, states2)
	}

	states1 = service.States{service.New, service.Starting, service.Running}
	states2 = service.States{service.New, service.Running, service.Starting}
	if !states1.Equals(states2) {
		t.Errorf("Should be equal: %v %v", states1, states2)
	}

	states1 = service.States{service.New, service.Starting, service.Running, service.Stopping, service.Terminated, service.Failed}
	states2 = service.States{service.New, service.Starting, service.Running, service.Stopping, service.Terminated, service.Failed}
	sort.Reverse(states2)
	if !states1.Equals(states2) {
		t.Errorf("Should be equal: %v %v", states1, states2)
	}
}

func TestState_String(t *testing.T) {
	states := map[service.State]string{
		service.New : "New",
		service.Starting : "Starting",
		service.Running : "Running",
		service.Stopping : "Stopping",
		service.Terminated : "Terminated",
		service.Failed : "Failed",
	}

	for state, s := range states {
		if state.String() != s {
			t.Errorf("%v != %v",state,s)
		}
	}
}
