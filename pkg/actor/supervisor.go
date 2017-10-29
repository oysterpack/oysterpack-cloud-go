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

import (
	"errors"
	"fmt"
	"time"
)

type SupervisorStrategy interface {
	HandleFailure(child *Actor, err *MessageProcessingError)
}

type Decider func(err error) Directive

// Directive represents an enum. The directive specifies how the actor failure should be handled.
type Directive int

// Directive enum values
const (
	DIRECTIVE_RESTART_ACTOR = iota
	DIRECTIVE_RESTART_ACTOR_HIERARCHY
	DIRECTIVE_RESTART_ACTOR_KILL_CHILDREN
	DIRECTIVE_STOP
	DIRECTIVE_ESCALATE
)

type SupervisorStrategyFactory interface {
	DefaultSupervisorStrategy() SupervisorStrategy

	NewExponentialBackoffStrategy(backoffWindow time.Duration, initialBackoff time.Duration) SupervisorStrategy

	NewOneForOneStrategy(maxNrOfRetries int, withinDuration time.Duration, decider Decider) SupervisorStrategy

	RestartingStrategy() SupervisorStrategy
}

// NewAllForOneStrategy returns a strategy that applies the fault handling Directive (Resume, Restart, Stop) specified
// in the Decider to all children when one fails.
//
// A child actor is restarted a max number of times number of times within a time duration, negative value means no limit.
// If the limit is exceeded the child actor is stopped.
//
// An error occurs if :
//	- if maxRetries > 0, then withinDuration must also be > 0
//  - if decider is nil
func NewAllForOneStrategy(maxRetries int, withinDuration time.Duration, decider Decider) (SupervisorStrategy, error) {
	if maxRetries > 0 && withinDuration == 0 {
		return nil, fmt.Errorf("If maxRetries > 0, then withinDuration must also be > 0")
	}

	if decider == nil {
		return nil, errors.New("Descider is required")
	}
	return &allForOneStrategy{maxRetries: maxRetries, withinDuration: withinDuration, decider: decider}, nil
}

type allForOneStrategy struct {
	maxRetries     int
	withinDuration time.Duration
	decider        Decider
}

func (a *allForOneStrategy) HandleFailure(child *Actor, err *MessageProcessingError) {
	// TODO: log the error
	child.failures.failure(err.Err)
	switch a.decider(err.Err) {
	case DIRECTIVE_RESTART_ACTOR:
		if a.canRestart(child) {
			child.restart(RESTART_ACTOR)
		} else {
			child.Kill(nil)
		}
	case DIRECTIVE_RESTART_ACTOR_HIERARCHY:
		if a.canRestart(child) {
			child.restart(RESTART_ACTOR_HIERARCHY)
		} else {
			child.Kill(nil)
		}
	case DIRECTIVE_RESTART_ACTOR_KILL_CHILDREN:
		if a.canRestart(child) {
			child.restart(RESTART_ACTOR_KILL_CHILDREN)
		} else {
			child.Kill(nil)
		}
	case DIRECTIVE_STOP:
		child.Kill(nil)
	case DIRECTIVE_ESCALATE:
		child.Kill(err)
	}

}

// check if the actor can be restarted based on the strategy configuration
func (a *allForOneStrategy) canRestart(actor *Actor) bool {
	if a.maxRetries < 0 {
		return true
	}

	if a.maxRetries == 0 {
		return false
	}

	if time.Since(actor.failures.lastFailureTime) < a.withinDuration {
		return actor.failures.count <= a.maxRetries
	}
	// we are past the time limit, we can safely reset the failure count and restart
	actor.failures.reset()
	return true

}
