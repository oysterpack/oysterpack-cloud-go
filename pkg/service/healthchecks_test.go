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

package service_test

import (
	"testing"

	"time"

	"errors"

	"github.com/oysterpack/oysterpack.go/pkg/metrics"
	"github.com/oysterpack/oysterpack.go/pkg/service"
	"github.com/prometheus/client_golang/prometheus"
)

func TestService_HealthChecks(t *testing.T) {
	defer metrics.ResetRegistry()
	metrics.ResetRegistry()
	metrics.Registry = prometheus.NewPedanticRegistry()

	var error1, error2, error3 *error

	healthchecks := []metrics.HealthCheck{
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
	}

	server := service.NewService(service.Settings{
		ServiceInterface: EchoServiceInterface,
		Version:          service.NewVersion("1.0.0"),
		HealthChecks:     healthchecks,
	})

	serviceHealthChecks := server.(service.HealthChecks)
	if len(serviceHealthChecks.HealthChecks()) != 3 {
		t.Errorf("There should be 3 healthchecks registered : %v", serviceHealthChecks.HealthChecks())
	}

	if len(serviceHealthChecks.FailedHealthChecks()) != 0 || len(serviceHealthChecks.SucceededHealthChecks()) != 0 {
		t.Errorf("No healthchecks have been run yet")
	}

}

func TestService_RunAllHealthChecks(t *testing.T) {
	defer metrics.ResetRegistry()
	metrics.ResetRegistry()
	metrics.Registry = prometheus.NewPedanticRegistry()

	var error1, error2, error3 *error

	healthchecks := []metrics.HealthCheck{
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
	}

	server := service.NewService(service.Settings{
		ServiceInterface: EchoServiceInterface,
		Version:          service.NewVersion("1.0.0"),
		HealthChecks:     healthchecks,
	})

	serviceHealthChecks := server.(service.HealthChecks)
	if len(serviceHealthChecks.HealthChecks()) != 3 {
		t.Errorf("There should be 3 healthchecks registered : %v", serviceHealthChecks.HealthChecks())
	}

	for healthcheck := range serviceHealthChecks.RunAllHealthChecks() {
		t.Log(healthcheck.LastResult().String())
		if !healthcheck.LastResult().Success() {
			t.Errorf("healthcheck should have passed : %v", healthcheck.Key())
		}
	}

	if len(serviceHealthChecks.FailedHealthChecks()) != 0 || len(serviceHealthChecks.SucceededHealthChecks()) != 3 {
		t.Errorf("There should be no failed health checks")
	}
}

func TestService_RunAllFailedHealthChecks(t *testing.T) {
	defer metrics.ResetRegistry()
	metrics.ResetRegistry()
	metrics.Registry = prometheus.NewPedanticRegistry()

	var error1, error2, error3 *error

	healthchecks := []metrics.HealthCheck{
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
	}

	server := service.NewService(service.Settings{
		ServiceInterface: EchoServiceInterface,
		Version:          service.NewVersion("1.0.0"),
		HealthChecks:     healthchecks,
	})

	serviceHealthChecks := server.(service.HealthChecks)

	err := errors.New("ERROR 1")
	error1 = &err
	t.Logf("healthcheck_1 should fail")
	for healthcheck := range serviceHealthChecks.RunAllHealthChecks() {
		t.Log(healthcheck.LastResult().String())
	}
	if len(serviceHealthChecks.FailedHealthChecks()) != 1 || len(serviceHealthChecks.SucceededHealthChecks()) != 2 {
		t.Errorf("There should be 1 failed health checks")
	}

	error1 = nil
	t.Logf("All healthchecks should pass after RunAllFailedHealthChecks")
	failedHealthCheckCount := 0
	for healthcheck := range serviceHealthChecks.RunAllFailedHealthChecks() {
		failedHealthCheckCount++
		t.Log(healthcheck.LastResult().String())
	}
	if failedHealthCheckCount != 1 {
		t.Errorf("There should be 1 failed health checks")
	}
	if len(serviceHealthChecks.FailedHealthChecks()) != 0 || len(serviceHealthChecks.SucceededHealthChecks()) != 3 {
		t.Errorf("There should be no failed health checks")
	}
}
