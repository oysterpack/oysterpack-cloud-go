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
	"reflect"

	"github.com/oysterpack/oysterpack.go/oysterpack/internal/utils"
)

type Service interface {
	Context() Context
}

type Context struct {
	reflect.Type

	LifeCycle
}

type LifeCycle struct {
	state ServiceState

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

// State returns the current State and when it transitioned into the State
func (c *LifeCycle) State() State {
	state, _ := c.state.State()
	return state
}

func (c *LifeCycle) FailureCause() error {
	return c.state.FailureCause()
}

func (c *LifeCycle) AwaitState(stateWaitedFor State) error {
	matches := func(currentState State) (bool, error) {
		switch {
		case currentState == stateWaitedFor:
			return true, nil
		case currentState > stateWaitedFor:
			return false, &IllegalStateError{currentState}
		default:
			return false, nil
		}
	}

	if reachedState, err := matches(c.State()); err != nil {
		return err
	} else if reachedState {
		return nil
	}
	l := c.state.NewStateChangeListener()
	// in case the service started matches in the meantime, seed the channel with the current state
	go func() {
		// ignore panics caused by sending on a closed channel
		// the channel might be closed if the service failed
		defer utils.IgnorePanic()
		if stateChangeChann := c.state.stateChangeChannel(l); stateChangeChann != nil {
			state, _ := c.state.State()
			stateChangeChann <- state
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

func (c *LifeCycle) AwaitTerminated() error {
	return c.AwaitState(Terminated)
}
