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

package comp_test

import (
	"testing"

	"time"

	"errors"

	"github.com/oysterpack/oysterpack.go/pkg/comp"
	"github.com/oysterpack/oysterpack.go/pkg/metrics"
	"github.com/prometheus/client_golang/prometheus"
)

func TestHealthChecks(t *testing.T) {
	defer metrics.ResetRegistry()
	metrics.ResetRegistry()
	metrics.Registry = prometheus.NewPedanticRegistry()

	var error1, error2, error3 *error

	healthchecks := comp.NewHealthChecks(
		metrics.NewHealthCheck(
			prometheus.GaugeOpts{Name: "healthcheck_1", Help: "HealthCheck 1"},
			0,
			func() error {
				if error1 != nil {
					return *error1
				}
				return nil
			}),
		metrics.NewHealthCheck(
			prometheus.GaugeOpts{Name: "healthcheck_2", Help: "HealthCheck 2"},
			0,
			func() error {
				if error2 != nil {
					return *error2
				}
				return nil
			}),
		metrics.NewHealthCheck(
			prometheus.GaugeOpts{Name: "healthcheck_3", Help: "HealthCheck 3"},
			0,
			func() error {
				if error3 != nil {
					return *error3
				}
				return nil
			}),
	)

	if len(healthchecks.HealthChecks()) != 3 {
		t.Errorf("There should be 3 healthchecks registered : %v", healthchecks.HealthChecks())
	}

	if len(healthchecks.FailedHealthChecks()) != 0 || len(healthchecks.SucceededHealthChecks()) != 0 {
		t.Errorf("No healthchecks have been run yet")
	}

}

func TestHealthchecks_RunAllHealthChecks(t *testing.T) {
	defer metrics.ResetRegistry()
	metrics.ResetRegistry()
	metrics.Registry = prometheus.NewPedanticRegistry()

	var error1, error2, error3 *error

	healthchecks := comp.NewHealthChecks(
		metrics.NewHealthCheck(
			prometheus.GaugeOpts{Name: "healthcheck_1", Help: "HealthCheck 1"},
			15*time.Second,
			func() error {
				if error1 != nil {
					return *error1
				}
				return nil
			}),
		metrics.NewHealthCheck(
			prometheus.GaugeOpts{Name: "healthcheck_2", Help: "HealthCheck 2"},
			15*time.Second,
			func() error {
				if error2 != nil {
					return *error2
				}
				return nil
			}),
		metrics.NewHealthCheck(
			prometheus.GaugeOpts{Name: "healthcheck_3", Help: "HealthCheck 3"},
			15*time.Second,
			func() error {
				if error3 != nil {
					return *error3
				}
				return nil
			}),
	)

	if len(healthchecks.HealthChecks()) != 3 {
		t.Errorf("There should be 3 healthchecks registered : %v", healthchecks.HealthChecks())
	}

	for healthcheck := range healthchecks.RunAllHealthChecks() {
		t.Log(healthcheck.LastResult().String())
		if !healthcheck.LastResult().Success() {
			t.Errorf("healthcheck should have passed : %v", healthcheck.Key())
		}
	}

	if len(healthchecks.FailedHealthChecks()) != 0 || len(healthchecks.SucceededHealthChecks()) != 3 {
		t.Errorf("There should be no failed health checks")
	}
}

func TestHealthChecks_RunAllFailedHealthChecks(t *testing.T) {
	defer metrics.ResetRegistry()
	metrics.ResetRegistry()
	metrics.Registry = prometheus.NewPedanticRegistry()

	var error1, error2, error3 *error

	healthchecks := comp.NewHealthChecks(
		metrics.NewHealthCheck(
			prometheus.GaugeOpts{Name: "healthcheck_1", Help: "HealthCheck 1"},
			15*time.Second,
			func() error {
				if error1 != nil {
					return *error1
				}
				return nil
			}),
		metrics.NewHealthCheck(
			prometheus.GaugeOpts{Name: "healthcheck_2", Help: "HealthCheck 2"},
			15*time.Second,
			func() error {
				if error2 != nil {
					return *error2
				}
				return nil
			}),
		metrics.NewHealthCheck(
			prometheus.GaugeOpts{Name: "healthcheck_3", Help: "HealthCheck 3"},
			15*time.Second,
			func() error {
				if error3 != nil {
					return *error3
				}
				return nil
			}),
	)

	err := errors.New("ERROR 1")
	error1 = &err
	t.Logf("healthcheck_1 should fail")
	for healthcheck := range healthchecks.RunAllHealthChecks() {
		t.Log(healthcheck.LastResult().String())
	}
	if len(healthchecks.FailedHealthChecks()) != 1 || len(healthchecks.SucceededHealthChecks()) != 2 {
		t.Errorf("There should be 1 failed health checks")
	}

	error1 = nil
	t.Logf("All healthchecks should pass after RunAllFailedHealthChecks")
	failedHealthCheckCount := 0
	for healthcheck := range healthchecks.RunAllFailedHealthChecks() {
		failedHealthCheckCount++
		t.Log(healthcheck.LastResult().String())
	}
	if failedHealthCheckCount != 1 {
		t.Errorf("There should be 1 failed health checks")
	}
	if len(healthchecks.FailedHealthChecks()) != 0 || len(healthchecks.SucceededHealthChecks()) != 3 {
		t.Errorf("There should be no failed health checks")
	}
}
