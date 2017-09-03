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

import "fmt"

// InvalidStateTransition indicates an invalid transition was attempted
type InvalidStateTransition struct {
	From State
	To   State
}

func (e *InvalidStateTransition) Error() string {
	return fmt.Sprintf("InvalidStateTransition: %v -> %v", e.From, e.To)
}

// IllegalStateError indicates we are in an illegal state
type IllegalStateError struct {
	State
}

func (e *IllegalStateError) Error() string {
	return e.State.String()
}

// UnknownFailureCause indicates that the service is in a Failed state, but the failure cause is unknown.
type UnknownFailureCause struct{}

func (_ UnknownFailureCause) Error() string {
	return "UnknownFailureCause"
}

// PastStateError indicates that we currently in a state that is past the desired state
type PastStateError struct {
	Past    State
	Current State
}

func (e *PastStateError) Error() string {
	return fmt.Sprintf("Current state (%v) is past state (%v)", e.Current, e.Past)
}
