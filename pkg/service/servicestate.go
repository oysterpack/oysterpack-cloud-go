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
	"sync"
	"time"

	"github.com/oysterpack/oysterpack.go/pkg/commons"
)

// ServiceState is used to manage the service's state.
// Use NewServiceState to construct new ServiceState instances
type ServiceState struct {
	// the mutex is needed to ensure that the ServiceState is consistently accessed and modified in memory from
	// multiple goroutines, which potentially may be running on different processors with its own local cache of main memeory
	mutex        sync.Mutex
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

// State returns the current State and when it transitioned to the State
func (s *ServiceState) State() (State, time.Time) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	return s.state, s.timestamp
}

// FailureCause returns the error that caused this service to fail.
// Returns nil if the service has no error.
// NOTE: If the service has a FaiureCause, then the State must be Failed
func (s *ServiceState) FailureCause() error {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	return s.failureCause
}

// SetState transitions to the specified State only if it is allowed, and records the timestamp.
// If the current state matches the new desired state, then false is returned.
// If an illegal state transition is attempted, then the state is not changed and an error is returned.
// If a valid state transition is requested, then the timestamp is updated and true is returned with no error.
func (s *ServiceState) SetState(state State) (bool, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
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
	s.mutex.Lock()
	defer s.mutex.Unlock()
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

// Starting transitions to Starting or returns an InvalidStateTransition error if it is not a legal transition
func (s *ServiceState) Starting() (bool, error) {
	return s.SetState(Starting)
}

// Running transitions to Running or returns an InvalidStateTransition error if it is not a legal transition
func (s *ServiceState) Running() (bool, error) {
	return s.SetState(Running)
}

// Stopping transitions to Stopping or returns an InvalidStateTransition error if it is not a legal transition
func (s *ServiceState) Stopping() (bool, error) {
	return s.SetState(Stopping)
}

// Terminated transitions to Terminated or returns an InvalidStateTransition error if it is not a legal transition
func (s *ServiceState) Terminated() (bool, error) {
	return s.SetState(Terminated)
}

// NewStateChangeListener returns a messages that clients can use to monitor the service lifecyle.
// After the service has reached a terminal state, then the messages will be closed
func (s *ServiceState) NewStateChangeListener() StateChangeListener {
	s.mutex.Lock()
	defer s.mutex.Unlock()
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
			func() {
				defer func() {
					if p := recover(); p != nil {
						// because ServiceState.stateChangeListeners is not thread-safe, it is possible that the slice was concurrently updated
						// this should be a rare event, and if it happens, simply retry
						s.deleteStateChangeListener(l)
					}
				}()
				s.stateChangeListeners[i] = s.stateChangeListeners[len(s.stateChangeListeners)-1]
				// in order to prevent memory leaks, i.e., to enable the channel at the last index to be garbage collected, it must be set to nil
				s.stateChangeListeners[len(s.stateChangeListeners)-1] = nil
				s.stateChangeListeners = s.stateChangeListeners[:len(s.stateChangeListeners)-1]
			}()
			return true
		}
	}
	return false
}

func (s *ServiceState) deleteAllStateChangeListeners() {
	temp := make([]chan State, len(s.stateChangeListeners))
	copy(temp, s.stateChangeListeners)
	for _, l := range temp {
		s.deleteStateChangeListener(l)
	}
}

// Ignores panic if the channel is already closed
func closeStateChanQuietly(c chan State) {
	defer commons.IgnorePanic()
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

// ContainsStateChangeListener returns true if the specified StateChangeListener is registered
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
				}()

				l <- state
			}(l)
		}
		s.deleteAllStateChangeListeners()
		return
	}
	var closedChannels []chan State
	for _, l := range s.stateChangeListeners {
		func(l chan State) {
			defer func() {
				// ignore panics caused by sending on a closed channel
				// this should normally never happen
				if p := recover(); p != nil {
					closedChannels = append(closedChannels, l)
				}
			}()
			l <- state
		}(l)
	}
	for _, l := range closedChannels {
		s.deleteStateChangeListener(l)
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

// Channel returns the channel that listener should listen on
func (a *StateChangeListener) Channel() <-chan State {
	return a.c
}

// Cancel deletes itself from the ServiceState that created it, which will also close the channel
// Any messages on the channel will be drained
func (a *StateChangeListener) Cancel() {
	a.s.mutex.Lock()
	defer a.s.mutex.Unlock()
	if !a.s.deleteStateChangeListener(a.c) {
		// the ServiceState reported that it did not own the channel
		// To be safe on the safe side, manually close the channel in case there are goroutines blocked on receiving from this channel
		closeStateChanQuietly(a.c)
	}
	// drain the channel
	for range a.c {
	}
}
