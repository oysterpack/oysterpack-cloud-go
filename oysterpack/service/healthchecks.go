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

import "github.com/oysterpack/oysterpack.go/oysterpack/metrics"

// HealthChecks service health checks
type HealthChecks interface {

	// HealthChecks returns all registered service health checks
	HealthChecks() []metrics.HealthCheck

	// FailedHealthChecks returns all failed health checks based on the last result.
	// If a HealthCheck has not yet been run, then it will be run.
	FailedHealthChecks() []metrics.HealthCheck

	// SucceededHealthChecks returns all health checks that succeeded based on the last result.
	// If a HealthCheck has not yet been run, then it will be run.
	SucceededHealthChecks() []metrics.HealthCheck

	// RunAllHealthChecks runs all registered health checks.
	// After a health check is run it is delivered on the channel.
	RunAllHealthChecks() <-chan metrics.HealthCheck

	// RunAllFailedHealthChecks all failed health checks based on the last result
	// If a HealthCheck has not yet been run, then it will be run.
	RunAllFailedHealthChecks() <-chan metrics.HealthCheck
}
