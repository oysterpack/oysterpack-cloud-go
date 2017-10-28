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

package actor

import "time"

type Failures struct {
	totalCount int

	lastFailureTime time.Time
	lastFailure     error

	count                 int
	resetTime             time.Time
	lastFailureSinceReset error
}

// Reset count back to 0. The count is reset when a new instance is created.
// It may also be be reset by a supervisor strategy, e.g., exponential back off strategy
func (a *Failures) reset() {
	a.count = 0
	a.resetTime = time.Now()
	a.lastFailureSinceReset = nil
}

func (a *Failures) failure(cause error) {
	a.totalCount++
	a.lastFailureTime = time.Now()
	a.lastFailure = cause

	a.count++
	a.lastFailureSinceReset = cause
}

// TotalCount returns the total number of errors that have occurred across all instances.
// For example, when an actor is restarted, a new instance is created. The count is reset, but the total count is not reset.
func (a *Failures) TotalCount() int {
	return a.totalCount
}

// LastFailureTime returns the last time a failure occurred for the actor
func (a *Failures) LastFailureTime() time.Time {
	return a.lastFailureTime
}

// LastFailure is the last actor failure that occurred
func (a *Failures) LastFailure() error {
	return a.lastFailure
}

// Count is the current failure count that is being tracked.
func (a *Failures) Count() int {
	return a.count
}

// ResetTime is the last time failures were reset
func (a *Failures) ResetTime() time.Time {
	return a.resetTime
}

// LastFailureSinceReset returns the last failure since being reset
func (a *Failures) LastFailureSinceReset() error {
	return a.lastFailureSinceReset
}
