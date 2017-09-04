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
	"github.com/oysterpack/oysterpack.go/oysterpack/internal/utils"
	"reflect"
	"time"
)

// TODO: metrics
// TODO: healthchecks
// TODO: events
// TODO: alarms
// TODO: config (JSON)
// TODO: readiness probe
// TODO: liveliness probe
// TODO: devops
type Service interface {
	Context() Context
}

type Context struct {
	reflect.Type

	LifeCycle
}

type LifeCycle struct {
	serviceState ServiceState

	stopTriggered bool
	stopTrigger   chan struct{}

	init    func() error
	run     func(StopTrigger) error
	destroy func() error
}

// State returns the current State
func (svc *LifeCycle) State() State {
	state, _ := svc.serviceState.State()
	return state
}

func (svc *LifeCycle) FailureCause() error {
	return svc.serviceState.FailureCause()
}

// awaitState blocks until the desired state is reached
// If the desired state has past, then a PastStateError is returned
func (svc *LifeCycle) awaitState(desiredState State, wait time.Duration) error {
	matches := func(currentState State) (bool, error) {
		switch {
		case currentState == desiredState:
			return true, nil
		case currentState > desiredState:
			return false, &PastStateError{Past: desiredState, Current: currentState}
		default:
			return false, nil
		}
	}

	if reachedState, err := matches(svc.State()); err != nil {
		return err
	} else if reachedState {
		return nil
	}
	l := svc.serviceState.NewStateChangeListener()
	if wait > 0 {
		timer := time.NewTimer(wait)
		go func() {
			l.Cancel()
		}()
		defer func() {
			timer.Stop()
			l.Cancel()
		}()
	} else {
		defer l.Cancel()
	}
	// in case the service started matches in the meantime, seed the messages with the current state
	go func() {
		// ignore panics caused by sending on a closed messages
		// the messages might be closed if the service failed
		defer utils.IgnorePanic()
		if stateChangeChann := svc.serviceState.stateChangeChannel(l); stateChangeChann != nil {
			stateChangeChann <- svc.State()
		}
	}()
	for state := range l.Channel() {
		if reachedState, err := matches(state); err != nil {
			return err
		} else if reachedState {
			return nil
		}
	}

	return svc.FailureCause()
}

// Waits for the Service to reach the running state
func (svc *LifeCycle) AwaitRunning(wait time.Duration) error {
	return svc.awaitState(Running, wait)
}

// Waits for the Service to terminate, i.e., reach the Terminated or Failed state
// if the service terminates in a Failed state, then the service failure cause is returned
func (svc *LifeCycle) AwaitTerminated(wait time.Duration) error {
	if err := svc.awaitState(Terminated, wait); err != nil {
		return svc.serviceState.failureCause
	}
	return nil
}

// If the service state is 'New', this initiates service startup and returns immediately.
// Returns an IllegalStateError if the service state is not 'New'.
// A stopped service may not be restarted.
func (svc *LifeCycle) StartAsync() error {
	if svc.serviceState.state.New() {
		go func() {
			svc.stopTrigger = make(chan struct{})
			svc.serviceState.Starting()
			if err := svc.init(); err != nil {
				svc.destroy()
				svc.serviceState.Failed(&ServiceError{State: Starting, Err: err})
				return
			}
			svc.serviceState.Running()
			if err := svc.run(svc.stopTrigger); err != nil {
				svc.destroy()
				svc.serviceState.Failed(&ServiceError{State: Running, Err: err})
				return
			}
			svc.serviceState.Stopping()
			if err := svc.destroy(); err != nil {
				svc.serviceState.Failed(&ServiceError{State: Stopping, Err: err})
				return
			}
			svc.serviceState.Terminated()
		}()
		return nil
	}
	return &IllegalStateError{
		State:   svc.serviceState.state,
		Message: "A service can only be started in the 'New' state",
	}
}

// If the service is starting or running, this initiates service shutdown and returns immediately.
// If the service is new, it is terminated without having been started nor stopped.
// If the service has already been stopped, this method returns immediately without taking action.
func (svc *LifeCycle) StopAsyc() {
	if svc.serviceState.state.Stopped() {
		return
	}
	if svc.serviceState.state.New() {
		svc.serviceState.Terminated()
		return
	}
	svc.stopTriggered = true
	close(svc.stopTrigger)
}

type StopTrigger <-chan struct{}
