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

// ServiceState is used to manage the service's state.
// Use NewServiceState to construct new ServiceState instances
type ServiceState struct {
	state        State
	failureCause error
	timestamp    time.Time

	// registered listeners for state changes
	// once the service is stopped, the listeners are cleared
	stateChangeListeners []chan State
}

// NewServiceState initializes that state timestamp to now
func NewServiceState() *ServiceState {
	return &ServiceState{
		timestamp: time.Now(),
	}
}

func (s *ServiceState) String() string {
	if s.failureCause != nil {
		return fmt.Sprintf("State : %v, Timestamp : %v, len(StateChangeListeners) : %v, FailureCause : %v", s.state, s.timestamp, len(s.stateChangeListeners), s.failureCause)
	}
	return fmt.Sprintf("State : %v, Timestamp : %v, len(StateChangeListeners) : %v", s.state, s.timestamp, len(s.stateChangeListeners))
}

func (s *ServiceState) State() (State, time.Time) {
	return s.state, s.timestamp
}

// FailureCause returns the error that caused this service to fail.
// Returns nil if the service has no error.
// NOTE: If the service has a FaiureCause, then the State must be Failed
func (s *ServiceState) FailureCause() error {
	return s.failureCause
}

// If the current state matches the new desired state, then false is returned.
// If an illegal state transition is attempted, then the state is not changed and an error is returned.
// If a valid state transition is requested, then the timestamp is updated and true is returned with no error.
func (s *ServiceState) SetState(state State) (bool, error) {
	if s.state == state {
		return false, nil
	}
	if s.state.ValidTransition(state) {
		s.state = state
		s.timestamp = time.Now()
		if state == Failed && s.failureCause == nil {
			s.failureCause = UnknownFailureCause{}
		}
		s.notifyStateChangeListeners(state)
		return true, nil
	}

	return false, &InvalidStateTransition{s.state, state}
}

// Failed sets the service state to Failed with the specified error only if it is a valid state transition.
// If err is nil, then err is set to UnknownFailureCause
// If the current state is already Failed, then the failure cause will be updated if err is not nil, but state change
// listeners will not be notified.
func (s *ServiceState) Failed(err error) bool {
	setFailureCause := func() {
		if err != nil {
			s.failureCause = err
		}
		if s.failureCause == nil {
			s.failureCause = UnknownFailureCause{}
		}
	}

	state := Failed
	if s.state == state {
		setFailureCause()
		return false
	}
	if s.state.ValidTransition(state) {
		s.state = state
		s.timestamp = time.Now()
		setFailureCause()
		s.notifyStateChangeListeners(state)
		return true
	}
	return false
}

func (s *ServiceState) Starting() (bool, error) {
	return s.SetState(Starting)
}

func (s *ServiceState) Running() (bool, error) {
	return s.SetState(Running)
}

func (s *ServiceState) Stopping() (bool, error) {
	return s.SetState(Stopping)
}

func (s *ServiceState) Terminated() (bool, error) {
	return s.SetState(Terminated)
}

// NewStateChangeListener returns a messages that clients can use to monitor the service lifecyle.
// After the service has reached a terminal state, then the messages will be closed
func (s *ServiceState) NewStateChangeListener() StateChangeListener {
	// There should be at most 4 state transitions
	// We want to ensure the state changes are sent, i.e., not blocked, even if there are no listeners actively receiving on the channel
	c := make(chan State, 4)
	if s.state.Stopped() {
		c <- s.state
		close(c)
	} else {
		s.stateChangeListeners = append(s.stateChangeListeners, c)
	}
	return StateChangeListener{c, s}
}

// deleteStateChangeListener closes the listener channel and removes it from its maintained list of StateChangeListener(s)
// All messages on the channel will be drained.
// Returns true if the listener channel existed and was deleted.
// Returns false if the listener channel is not currently owned by this ServiceState.
func (s *ServiceState) deleteStateChangeListener(l chan State) bool {
	for i, v := range s.stateChangeListeners {
		if l == v {
			closeStateChanQuietly(l)
			s.stateChangeListeners[i] = s.stateChangeListeners[len(s.stateChangeListeners)-1]
			s.stateChangeListeners = s.stateChangeListeners[:len(s.stateChangeListeners)-1]
			return true
		}
	}
	return false
}

// Ignores panic if the channel is already closed
func closeStateChanQuietly(c chan State) {
	defer utils.IgnorePanic()
	close(c)
}

// stateChangeChannel looks up the StateChangeListener channel. If it is not found, then nil is returned.
func (s *ServiceState) stateChangeChannel(l StateChangeListener) chan State {
	for _, v := range s.stateChangeListeners {
		if l.c == v {
			return v
		}
	}
	return nil
}

func (s *ServiceState) ContainsStateChangeListener(l StateChangeListener) bool {
	if l.s != s {
		return false
	}

	return s.stateChangeChannel(l) != nil
}

// Each StateChangeListener is notified async, i.e., concurrently.
func (s *ServiceState) notifyStateChangeListeners(state State) {
	if state.Stopped() {
		for _, l := range s.stateChangeListeners {
			func(l chan State) {
				defer func() {
					// ignore panics caused by sending on a closed channel
					recover()
					s.deleteStateChangeListener(l)
				}()

				l <- state
			}(l)
		}
		s.stateChangeListeners = nil
	} else {
		for _, l := range s.stateChangeListeners {
			func(l chan State) {
				defer func() {
					// ignore panics caused by sending on a closed channel
					if p := recover(); p != nil {
						s.deleteStateChangeListener(l)
					}
				}()
				l <- state
			}(l)
		}
	}
}

// StateChangeListener contains a channel used to listen for service state changes.
// After a terminal state is reached, then the channel is closed.
// If one listens on a closed (cancelled) channel, then the 'New' State will be returned because New == 0, which is the default int value.
// The 'New' state will never be delivered on an active channel because the StateChangeListener is created after the service.
// NOTE: a StateChangeListener must be created using ServiceState.NewStateChangeListener()
type StateChangeListener struct {
	c chan State
	// owns the listener channel
	s *ServiceState
}

func (a *StateChangeListener) Channel() <-chan State {
	return a.c
}

// Cancel deletes itself from the ServiceState that created it, which will also close the channel
// Any messages on the channel will be drained
func (a *StateChangeListener) Cancel() {
	if !a.s.deleteStateChangeListener(a.c) {
		// the ServiceState reported that it did not own the channel
		// To be safe on the safe side, manually close the channel in case there are goroutines blocked on receiving from this channel
		closeStateChanQuietly(a.c)
	}
	for range a.c {
	}
}
