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
	"sort"
	"sync"
	"time"
)

// State is a simple hight-level summary of where the LifeCycle is in its lifecycle
type State int

// Possible State values
// Normal service life cycle : New -> Starting -> Running -> Stopping -> Terminated
// If the service fails while starting, running, or stopping, then it goes into state LifeCycle.State.FAILED.
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

func (s State) New() bool { return s == New }

func (s State) Starting() bool { return s == Starting }

func (s State) Running() bool { return s == Running }

func (s State) Stopping() bool { return s == Stopping }

func (s State) Terminated() bool { return s == Terminated }

func (s State) Failed() bool { return s == Failed }

// Stopped returns true if the serivce is Terminated or Failed
func (s State) Stopped() bool {
	return s == Terminated || s == Failed
}

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

// States implements sort.Interface
type States []State

func (a States) Len() int           { return len(a) }
func (a States) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a States) Less(i, j int) bool { return a[i] < a[j] }

var AllStates States = []State{New, Starting, Running, Stopping, Terminated, Failed}

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

// Used to listen for service state changes
// Once a terminal state is reached, then the channel is closed
type StateChangeListener <-chan State

type StateChangeListeners []chan State

// ServiceState is used to manage the service's state in a concurrent safe manner.
type ServiceState struct {
	lock         sync.RWMutex
	state        State
	failureCause error
	timestamp    time.Time

	// registered listeners for state changes
	// once the service is stopped, the listeners are cleared
	stateChangeListeners StateChangeListeners
}

func (s *ServiceState) State() (State, time.Time) {
	s.lock.RLock()
	defer s.lock.RUnlock()
	return s.state, s.timestamp
}

// FailureCause returns the error that caused this service to fail.
// Returns nil if the service has no error.
// NOTE: If the service has a FaiureCause, then the State must be Failed
func (s *ServiceState) FailureCause() error {
	s.lock.RLock()
	defer s.lock.RUnlock()
	return s.failureCause
}

// If the current state matches the new desired state, then false is returned.
// If an illegal state transition is attempted, then the state is not changed and an error is returned.
// If a valid state transition is requested, then the timestamp is updated and true is returned with no error.
func (s *ServiceState) SetState(state State) (bool, error) {
	s.lock.Lock()
	defer s.lock.Unlock()
	if s.state == state {
		return false, nil
	}
	if s.state.ValidTransition(state) {
		s.state = state
		s.timestamp = time.Now()
		if state == Failed {
			s.failureCause = UnknownFailureCause{}
		}
		go s.notifyStateChangeListeners()
		return true, nil
	}

	return false, &InvalidStateTransition{s.state, state}
}

// Failed sets the service state to Failed with the specified error.
// If err is nil, then err is set to UnknownFailureCause
func (s *ServiceState) Failed(err error) {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.state = Failed
	s.timestamp = time.Now()
	if err == nil {
		err = UnknownFailureCause{}
	}
	s.failureCause = err
	go s.notifyStateChangeListeners()
}

// NewStateChangeListener returns a channel that clients can use to monitor the service lifecyle.
// After the service has reached a terminal state, then the channel will be closed
func (s *ServiceState) NewStateChangeListener() StateChangeListener {
	s.lock.Lock()
	defer s.lock.Unlock()
	l := make(chan State)
	if s.state.Stopped() {
		close(l)
	} else {
		s.stateChangeListeners = append(s.stateChangeListeners, l)
	}
	return l
}

func (s *ServiceState) deleteStateChangeListener(l chan State) {
	// Ignores panic if the channel is already closed
	closeQuietly := func(c chan State) {
		defer utils.IgnorePanic()
		close(c)
	}

	s.lock.Lock()
	defer s.lock.Unlock()
	for i, v := range s.stateChangeListeners {
		if l == v {
			closeQuietly(l)
			s.stateChangeListeners[i] = s.stateChangeListeners[len(s.stateChangeListeners)-1]
			s.stateChangeListeners = s.stateChangeListeners[:len(s.stateChangeListeners)-1]
			return
		}
	}
}

func (s *ServiceState) stateChangeChannel(l StateChangeListener) chan State {
	s.lock.RLock()
	defer s.lock.RUnlock()
	for _, v := range s.stateChangeListeners {
		if l == v {
			return v
		}
	}
	return nil
}

// Each StateChangeListener is notified async, i.e., concurrently.
// However, this func will block until all listeners are notified, i.e., the State is sent on the channel.
func (s *ServiceState) notifyStateChangeListeners() {
	if state := s.state; state.Stopped() {
		s.lock.Lock()
		defer s.lock.Unlock()
		waitGroup := sync.WaitGroup{}
		for _, l := range s.stateChangeListeners {
			waitGroup.Add(1)
			go func(l chan State) {
				defer func() {
					// ignore panics caused by sending on a closed channel
					recover()
					waitGroup.Done()
				}()

				l <- state
				if state.Stopped() {
					close(l)
				}
			}(l)
		}
		waitGroup.Wait()
		s.stateChangeListeners = nil
	} else {
		s.lock.RLock()
		defer s.lock.RUnlock()
		waitGroup := sync.WaitGroup{}
		for _, l := range s.stateChangeListeners {
			waitGroup.Add(1)
			go func(l chan State) {
				defer func() {
					// ignore panics caused by sending on a closed channel
					recover()
					waitGroup.Done()
					go s.deleteStateChangeListener(l)
				}()
				l <- state
			}(l)
		}
		waitGroup.Wait()
	}
}

type InvalidStateTransition struct {
	From State
	To   State
}

func (e *InvalidStateTransition) Error() string {
	return fmt.Sprintf("InvalidStateTransition: %v -> %v", e.From, e.To)
}

type UnknownFailureCause struct{}

func (_ UnknownFailureCause) Error() string {
	return "UnknownFailureCause"
}

type IllegalStateError struct {
	state State
}

func (e *IllegalStateError) Error() string {
	return e.state.String()
}
