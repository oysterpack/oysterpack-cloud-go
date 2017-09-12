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
	"sync"
	"testing"
	"time"
)

func TestServiceState_NewServiceState(t *testing.T) {
	now := time.Now()
	serviceState := service.NewServiceState()
	t.Log(serviceState.String())
	if state, ts := serviceState.State(); state != service.New {
		t.Errorf("A new ServiceState should initially be in a New State, but it was : %v", state)
		if ts.Before(now) {
			t.Errorf("The state timestamp is too old. It should have been around now")
		}
	}
}

func TestServiceState_SetState(t *testing.T) {
	serviceState := service.NewServiceState()
	state1, ts1 := serviceState.State()
	t.Logf("state1: %v, ts: %v", state1, ts1)
	serviceState.SetState(service.Starting)
	state2, ts2 := serviceState.State()
	t.Logf("state2: %v, ts: %v", state2, ts2)
	if state2 != service.Starting {
		t.Errorf("State should have changed to Starting but is : %v", state2)
	}
	if !ts2.After(ts1) {
		t.Errorf("The timestamp after changing the state should be after the previous state timestamp : previous (%v) current (%v)", ts1, ts2)
	}

	if set, err := serviceState.SetState(service.Starting); set || err != nil {
		t.Errorf("Setting to the same state, should not update the state")
	}
	if _, ts3 := serviceState.State(); ts3 != ts2 {
		t.Errorf("Setting to the same state, should not update the state, which means the timestamp should not have changed : %v -> %v", ts2, ts3)
	}

	if set, err := serviceState.SetState(service.New); set || err == nil {
		t.Errorf("An invalid state transition should have failed : set = %v, err = %v", set, err)
	}

}

func TestNewServiceState_SetState_Failed(t *testing.T) {
	serviceState := service.NewServiceState()
	serviceState.SetState(service.Starting)
	serviceState.SetState(service.Failed)
	if err := serviceState.FailureCause(); err == nil {
		t.Error("FailureCause should be UnknownFailureCause but was nil")
	} else {
		switch err.(type) {
		case service.UnknownFailureCause:
		default:
			t.Errorf("FailureCause should be UnknownFailureCause but was : %T", err)
		}
	}
}


func TestServiceState_New2Fail(t *testing.T) {
	serviceState := service.NewServiceState()
	_, stateTS := serviceState.State()
	serviceState.Failed(nil)
	if state, ts := serviceState.State(); state != service.New || !ts.Equal(stateTS) {
		if state != service.Failed {
			t.Errorf("The state should not have changed because it is invalid to transition from New -> Fail", state)
		}
		if !ts.Equal(stateTS) {
			t.Errorf("The state timestamp should not have changed : %v -> %v", stateTS, ts)
		}
	}
	if err := serviceState.FailureCause(); err != nil {
		t.Errorf("FailureCayse should be nil, but was : %v", err)
	}
}

func TestServiceState_Starting2Failed(t *testing.T) {
	serviceState := service.NewServiceState()
	serviceState.SetState(service.Starting)
	_, stateTS := serviceState.State()
	serviceState.Failed(nil)
	if state, ts := serviceState.State(); state != service.Failed || !ts.After(stateTS) {
		if state != service.Failed {
			t.Errorf("The state should have been Failed, but was %v", state)
		}
		if !ts.After(stateTS) {
			t.Errorf("Setting the state to Failed should have also updated the timestamp : %v -> %v", stateTS, ts)
		}
	}
	switch serviceState.FailureCause().(type) {
	case service.UnknownFailureCause:
	default:
		t.Errorf("The FailureCause type should have been UnknownFailureCause but was %T", serviceState.FailureCause())
	}
}

func TestServiceState_Failed2Failed(t *testing.T) {
	serviceState := service.NewServiceState()
	serviceState.SetState(service.Starting)
	_, stateTS := serviceState.State()
	serviceState.Failed(nil)
	_, stateTS = serviceState.State()
	// fail the service again
	if serviceState.Failed(&service.IllegalStateError{State: service.Failed}) {
		t.Error("The state should not have been updated because it should already be Failed")
	}
	if state, ts := serviceState.State(); state != service.Failed || !ts.Equal(stateTS) {
		if state != service.Failed {
			t.Errorf("The state should have been Failed, but was %v", state)
		}
		if !ts.Equal(stateTS) {
			t.Errorf("The state timestamp should not have changed : %v -> %v", stateTS, ts)
		}
	}
	switch serviceState.FailureCause().(type) {
	case *service.IllegalStateError:
		t.Logf("%v", serviceState)
	default:
		t.Errorf("The FailureCause type should have been IllegalStateError but was %T", serviceState.FailureCause())
	}
}

func TestServiceState_NewStateChangeListener(t *testing.T) {
	serviceState := service.NewServiceState()
	l := serviceState.NewStateChangeListener()

	stateChanges := []service.State{}
	lClosed := sync.WaitGroup{}
	lClosed.Add(1)

	// wait for each state change notification to be received before transitioning to the next state
	starting, running, stopping, terminated := sync.WaitGroup{}, sync.WaitGroup{}, sync.WaitGroup{}, sync.WaitGroup{}
	starting.Add(1)
	running.Add(1)
	stopping.Add(1)
	terminated.Add(1)
	go func() {
		defer func() {
			lClosed.Done()
			t.Logf("StateChangeListener channel is closed : %v", serviceState)
		}()
		for state := range l.Channel() {
			stateChanges = append(stateChanges, state)
			t.Logf("State changed to : %v", state)
			switch state {
			case service.Starting:
				starting.Done()
			case service.Running:
				running.Done()
			case service.Stopping:
				stopping.Done()
			case service.Terminated:
				terminated.Done()
			default:
				t.Errorf("Unexpected state change : %v", state)
			}
		}
	}()

	serviceState.Starting()
	starting.Wait()

	serviceState.Running()
	running.Wait()

	serviceState.Stopping()
	stopping.Wait()

	serviceState.Terminated()
	terminated.Wait()

	lClosed.Wait()

	t.Logf("stateChanges : %v", stateChanges)

	if len(stateChanges) != 4 {
		t.Errorf("Expected 4 State transitions but got %d : %v", len(stateChanges), stateChanges)
	}
	if stateChanges[0] != service.Starting {
		t.Errorf("Expected state[0] to be Starting but was : %v", stateChanges[0])
	}
}

func TestServiceState_NewStateChangeListener_NonBlocking(t *testing.T) {
	serviceState := service.NewServiceState()

	t.Logf("before creating new state change listener : %v", serviceState)
	l := serviceState.NewStateChangeListener()
	t.Logf("after creating new state change listener : %v", serviceState)

	stateChanges := []service.State{}
	lClosed := sync.WaitGroup{}
	lClosed.Add(1)
	go func() {
		defer func() {
			lClosed.Done()
			t.Logf("StateChangeListener channel is closed : %v", serviceState)
		}()
		for state := range l.Channel() {
			stateChanges = append(stateChanges, state)
			t.Logf("State changed to : %v :: %v", state, serviceState)
		}
	}()

	serviceState.Starting()
	serviceState.Running()
	serviceState.Stopping()
	serviceState.Terminated()

	lClosed.Wait()

	t.Logf("stateChanges : %v", stateChanges)
	t.Logf("after terminated : %v", serviceState)

	if len(stateChanges) != 4 {
		t.Errorf("Expected 4 State transitions but got %d : %v", len(stateChanges), stateChanges)
	}
	if stateChanges[0] != service.Starting {
		t.Errorf("Expected state[0] to be Starting but was : %v", stateChanges[0])
	}

	l = serviceState.NewStateChangeListener()
	for state := range l.Channel() {
		if state != service.Terminated {
			t.Errorf("Expected state to be Terminated, but was %v", state)
		}
	}
}

func TestStateChangeListener_Cancel(t *testing.T) {
	serviceState := service.NewServiceState()
	l := serviceState.NewStateChangeListener()
	if !serviceState.ContainsStateChangeListener(l) {
		t.Error("ServiceState should have contained the StateChangeListener")
	}
	l.Cancel()
	if serviceState.ContainsStateChangeListener(l) {
		t.Error("ServiceState should not contain the StateChangeListener because it was cancelled")
	}
	// Cancelling again should be ok
	l.Cancel()

	l = serviceState.NewStateChangeListener()
	if !serviceState.ContainsStateChangeListener(l) {
		t.Error("ServiceState should have contained the StateChangeListener")
	}
	serviceState2 := service.NewServiceState()
	if serviceState2.ContainsStateChangeListener(l) {
		t.Error("ServiceState should not contain the StateChangeListener because it was created by a different ServiceState")
	}

	serviceState.Terminated()
	if serviceState.ContainsStateChangeListener(l) {
		t.Error("ServiceState should not contain the StateChangeListener because the ServiceState reached a terminal state")
	}
}
