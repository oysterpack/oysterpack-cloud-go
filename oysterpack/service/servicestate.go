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
	"sync"
	"time"
	"github.com/oysterpack/oysterpack.go/oysterpack/internal/utils"
	"fmt"
)

// StateChangeListener is a channel used to listen for service state changes.
// After a terminal state is reached, then the channel is closed.
type StateChangeListener <-chan State

// StateChangeListeners is a slice of StateChangeListener(s)
type StateChangeListeners []chan State

// ServiceState is used to manage the service's state in a concurrent safe manner.
// Use NewServiceState to construct new ServiceState instances
type ServiceState struct {
	lock         sync.RWMutex
	state        State
	failureCause error
	timestamp    time.Time

	// registered listeners for state changes
	// once the service is stopped, the listeners are cleared
	stateChangeListeners StateChangeListeners
}

// NewServiceState initializes that state timestamp to now
func NewServiceState() *ServiceState {
	return &ServiceState{
		timestamp: time.Now(),
	}
}

func (s *ServiceState) String() string {
	if (s.failureCause != nil) {
		return fmt.Sprintf("State : %v, Timestamp : %v, FailureCause : %v", s.state, s.timestamp, s.failureCause)
	}
	return fmt.Sprintf("State:%v, Timestamp:%v", s.state, s.timestamp)
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
		if state == Failed && s.failureCause == nil {
			s.failureCause = UnknownFailureCause{}
		}
		go s.notifyStateChangeListeners(state)
		return true, nil
	}

	return false, &InvalidStateTransition{s.state, state}
}

// Failed sets the service state to Failed with the specified error only if it is a valid state transition.
// If err is nil, then err is set to UnknownFailureCause
// If the current state is already Failed, then the failure cause will be updated if err is not nil, but state change
// listeners will not be notified.
func (s *ServiceState) Failed(err error) bool {
	s.lock.Lock()
	defer s.lock.Unlock()

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
		go s.notifyStateChangeListeners(state)
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

// NewStateChangeListener returns a channel that clients can use to monitor the service lifecyle.
// After the service has reached a terminal state, then the channel will be closed
func (s *ServiceState) NewStateChangeListener() StateChangeListener {
	s.lock.Lock()
	defer s.lock.Unlock()
	l := make(chan State)
	if s.state.Stopped() {
		go func() {
			l <- s.state
			closeQuietly(l)
		}()

	} else {
		s.stateChangeListeners = append(s.stateChangeListeners, l)
	}
	return l
}

func (s *ServiceState) deleteStateChangeListener(l chan State) {
	closeQuietly(l)

	s.lock.Lock()
	defer s.lock.Unlock()
	for i, v := range s.stateChangeListeners {
		if l == v {
			s.stateChangeListeners[i] = s.stateChangeListeners[len(s.stateChangeListeners)-1]
			s.stateChangeListeners = s.stateChangeListeners[:len(s.stateChangeListeners)-1]
			return
		}
	}
}

// Ignores panic if the channel is already closed
func closeQuietly(c chan State) {
	defer utils.IgnorePanic()
	close(c)
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
func (s *ServiceState) notifyStateChangeListeners(state State) {
	logger.Debug().Msgf("notifyStateChangeListeners : %v : %v",s, state)
	if state.Stopped() {
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
					go s.deleteStateChangeListener(l)
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
					if p := recover(); p != nil {
						go s.deleteStateChangeListener(l)
					}
					waitGroup.Done()
				}()
				l <- state
			}(l)
		}
		waitGroup.Wait()
	}
}
