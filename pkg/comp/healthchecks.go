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

package comp

import (
	"sync"

	"github.com/oysterpack/oysterpack.go/pkg/metrics"
)

type HealthChecks []metrics.HealthCheck

// FailedHealthChecks returns all failed health checks based on the last result.
// If a HealthCheck has not yet been run, then it will be run.
func (a HealthChecks) FailedHealthChecks() []metrics.HealthCheck {
	failed := []metrics.HealthCheck{}
	for _, healthcheck := range a {
		if healthcheck.LastResult() != nil && !healthcheck.LastResult().Success() {
			failed = append(failed, healthcheck)
		}
	}
	return failed
}

// SucceededHealthChecks returns all health checks that succeeded based on the last result.
// If a HealthCheck has not yet been run, then it will be run.
func (a HealthChecks) SucceededHealthChecks() []metrics.HealthCheck {
	succeeded := []metrics.HealthCheck{}
	for _, healthcheck := range a {
		if healthcheck.LastResult() != nil && healthcheck.LastResult().Success() {
			succeeded = append(succeeded, healthcheck)
		}
	}
	return succeeded
}

// RunAllHealthChecks runs all registered health checks.
// After a health check is run it is delivered on the channel.
func (a HealthChecks) RunAllHealthChecks() <-chan metrics.HealthCheck {
	count := len(a)
	c := make(chan metrics.HealthCheck, count)
	wait := sync.WaitGroup{}
	wait.Add(count)

	for _, healthcheck := range a {
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
func (a HealthChecks) RunAllFailedHealthChecks() <-chan metrics.HealthCheck {
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
