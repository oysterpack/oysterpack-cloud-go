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

	"github.com/oysterpack/oysterpack.go/pkg/metrics"
)

// HealthChecks service health checks
type HealthChecks interface {

	// HealthChecks returns all registered service health checks
	HealthChecks() []metrics.HealthCheck

	// FailedHealthChecks returns all failed health checks based on the last result.
	FailedHealthChecks() []metrics.HealthCheck

	// SucceededHealthChecks returns all health checks that succeeded based on the last result.
	SucceededHealthChecks() []metrics.HealthCheck

	// RunAllHealthChecks runs all registered health checks.
	// After a health check is run it is delivered on the channel.
	RunAllHealthChecks() <-chan metrics.HealthCheck

	// RunAllFailedHealthChecks all failed health checks based on the last result
	RunAllFailedHealthChecks() <-chan metrics.HealthCheck
}

// NewHealthChecks is a factory method for HealthChecks
func NewHealthChecks(checks ...metrics.HealthCheck) HealthChecks {
	return &healthchecks{checks}
}

type healthchecks struct {
	healthchecks []metrics.HealthCheck
}

func (a *healthchecks) HealthChecks() []metrics.HealthCheck {
	return a.healthchecks
}

// FailedHealthChecks returns all failed health checks based on the last result.
// If a HealthCheck has not yet been run, then it will be run.
func (a *healthchecks) FailedHealthChecks() []metrics.HealthCheck {
	failed := []metrics.HealthCheck{}
	for _, healthcheck := range a.healthchecks {
		if healthcheck.LastResult() != nil && !healthcheck.LastResult().Success() {
			failed = append(failed, healthcheck)
		}
	}
	return failed
}

// SucceededHealthChecks returns all health checks that succeeded based on the last result.
// If a HealthCheck has not yet been run, then it will be run.
func (a *healthchecks) SucceededHealthChecks() []metrics.HealthCheck {
	succeeded := []metrics.HealthCheck{}
	for _, healthcheck := range a.healthchecks {
		if healthcheck.LastResult() != nil && healthcheck.LastResult().Success() {
			succeeded = append(succeeded, healthcheck)
		}
	}
	return succeeded
}

// RunAllHealthChecks runs all registered health checks.
// After a health check is run it is delivered on the channel.
func (a *healthchecks) RunAllHealthChecks() <-chan metrics.HealthCheck {
	count := len(a.healthchecks)
	c := make(chan metrics.HealthCheck, count)
	wait := sync.WaitGroup{}
	wait.Add(count)

	for _, healthcheck := range a.healthchecks {
		go func(healthcheck metrics.HealthCheck) {
			defer func() {
				c <- healthcheck
				wait.Done()
			}()
			healthcheck.Run()
		}(healthcheck)
	}

	go func() {
		wait.Wait()
		close(c)
	}()

	return c
}

// RunAllFailedHealthChecks all failed health checks based on the last result
func (a *healthchecks) RunAllFailedHealthChecks() <-chan metrics.HealthCheck {
	failedHealthChecks := a.FailedHealthChecks()
	c := make(chan metrics.HealthCheck, len(failedHealthChecks))
	wait := sync.WaitGroup{}
	wait.Add(len(failedHealthChecks))

	for _, healthcheck := range failedHealthChecks {
		go func(healthcheck metrics.HealthCheck) {
			defer func() {
				c <- healthcheck
				wait.Done()
			}()
			healthcheck.Run()
		}(healthcheck)
	}

	go func() {
		wait.Wait()
		close(c)
	}()

	return c
}
