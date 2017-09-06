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
	"github.com/oysterpack/oysterpack.go/oysterpack/internal/utils"
	"time"
)

// Server represents the service backend. The server runs in its own goroutine.
// The server has a life cycle linked to its ServiceState.
// It must be created using NewServer
//
// Servers must be designed to be concurrent, but also concurrent safe.
// The solution is to design the server as a message processing server leveraging channels and goroutines.
// There would be a channel for each type of message a server can handle. Goroutines are leveraged for concurrent message processing.
// The design pattern is for the backend server Run function to process messages sent via channels.
// The Run function should be designed to multiplex on channels - the StopTrigger channel and a channel for each type of
// message that the server can handle.
//
// Each service will live in its own package. The service will be defined by the functions that are exposed by the package.
// Functions relay messages to the server backend via messages. If a reply is required, then the message will have a reply channel.
//
type Server struct {
	serviceState *ServiceState

	stopTriggered bool
	// closing the channel signals to the Run function that stop has been triggered
	stopTrigger chan struct{}

	init    Init
	run     Run
	destroy Destroy
}

type Init func() error

// Run is responsible to responding to a message on the StopTrigger channel.
// When a message is received from the StopTrigger, then the Run function should stop running ASAP.
type Run func(StopTrigger) error

type Destroy func() error

// NewServer creates and returns a new Server instance in the 'New' state
// All server life cycle functions are optional.
// Any panics that occur in the supplied functions are converted to errors.
func NewServer(init Init, run Run, destroy Destroy) *Server {
	if init == nil {
		init = func() error { return nil }
	} else {
		_init := init
		init = func() (err error) {
			defer func() {
				if p := recover(); p != nil {
					err = fmt.Errorf("init panic : %v", p)
				}
			}()
			return _init()
		}
	}

	if run == nil {
		run = func(stopTrigger StopTrigger) error {
			<-stopTrigger
			return nil
		}
	} else {
		_run := run
		run = func(stopTrigger StopTrigger) (err error) {
			defer func() {
				if p := recover(); p != nil {
					err = fmt.Errorf("run panic :%v", p)
				}
			}()
			return _run(stopTrigger)
		}
	}

	if destroy == nil {
		destroy = func() error { return nil }
	} else {
		_destroy := destroy
		destroy = func() (err error) {
			defer func() {
				if p := recover(); p != nil {
					err = fmt.Errorf("destroy panic :%v", p)
				}
			}()
			return _destroy()
		}
	}

	return &Server{
		serviceState: NewServiceState(),
		init:         init,
		run:          run,
		destroy:      destroy,
	}
}

// State returns the current State
func (svc *Server) State() State {
	state, _ := svc.serviceState.State()
	return state
}

func (svc *Server) FailureCause() error {
	return svc.serviceState.FailureCause()
}

// awaitState blocks until the desired state is reached
// If the desired state has past, then a PastStateError is returned
func (svc *Server) awaitState(desiredState State, wait time.Duration) error {
	matches := func(currentState State) (bool, error) {
		switch {
		case currentState == desiredState:
			return true, nil
		case currentState > desiredState:
			if svc.State().Failed() {
				return false, svc.FailureCause()
			}
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
		timer := time.AfterFunc(wait, l.Cancel)
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
func (svc *Server) AwaitRunning(wait time.Duration) error {
	return svc.awaitState(Running, wait)
}

// Waits for the Service to terminate, i.e., reach the Terminated or Failed state
// if the service terminates in a Failed state, then the service failure cause is returned
func (svc *Server) AwaitTerminated(wait time.Duration) error {
	if err := svc.awaitState(Terminated, wait); err != nil {
		return svc.serviceState.failureCause
	}
	return nil
}

// If the service state is 'New', this initiates service startup and returns immediately.
// Returns an IllegalStateError if the service state is not 'New'.
// A stopped service may not be restarted.
func (svc *Server) StartAsync() error {
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
func (svc *Server) StopAsyc() {
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

func (svc *Server) StopTriggered() bool {
	return svc.stopTriggered
}

type StopTrigger <-chan struct{}
