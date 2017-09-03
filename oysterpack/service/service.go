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

type Service interface {
	Context() Context
}

type Context struct {
	reflect.Type

	LifeCycle
}

type LifeCycle struct {
	ServiceState

	// OPTIONAL : StartUp is invoked when transitioning from New -> Starting
	// After StartUp is complete, if the service has a RunAsync, then it is invoked.
	// Once the RunAsync.Run has started, then the service is transitioned to Running.
	// If the service has no RunAsync func, then the service is transitioned to Running.
	// A stopped service may not be restarted.
	StartAsync func()

	// OPTIONAL : Run the service. This func is invoked async after StartUp completes.
	// Use cases :
	// 1. services that run as servers in the background
	// 2. services that run periodic tasks in the background
	Run func()

	// OPTIONAL : StopAsync function is run async
	// If the service is starting or running, this initiates service shutdown and returns immediately.
	// If the service is new, it is terminated without having been started nor stopped.
	// If the service has already been stopped, this method returns immediately without taking action.
	StopAsync func()
}

// State returns the current State
func (c *LifeCycle) State() State {
	state, _ := c.ServiceState.State()
	return state
}

func (c *LifeCycle) FailureCause() error {
	return c.ServiceState.FailureCause()
}

// AwaitState blocks until the desired state is reached
// If the desired state has past, then a PastStateError is returned
func (c *LifeCycle) AwaitState(desiredState State) error {
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

	if reachedState, err := matches(c.State()); err != nil {
		return err
	} else if reachedState {
		return nil
	}
	l := c.ServiceState.NewStateChangeListener()
	// in case the service started matches in the meantime, seed the messages with the current state
	go func() {
		// ignore panics caused by sending on a closed messages
		// the messages might be closed if the service failed
		defer utils.IgnorePanic()
		if stateChangeChann := c.ServiceState.stateChangeChannel(l); stateChangeChann != nil {
			stateChangeChann <- c.State()
		}
	}()
	for state := range l {
		if reachedState, err := matches(state); err != nil {
			return err
		} else if reachedState {
			return nil
		}
	}

	return c.FailureCause()
}

// Waits for the Service to reach the running state
func (c *LifeCycle) AwaitRunning() error {
	return c.AwaitState(Running)
}

// Waits for the Service to terminate, i.e., reach the Terminated or Failed state
// if the service terminates in a Failed state, then the service failure cause is returned
func (c *LifeCycle) AwaitTerminated() error {
	if err := c.AwaitState(Terminated); err != nil {
		return c.failureCause
	}
	return nil
}

type StopTrigger struct {
	Timestamp time.Time
}
