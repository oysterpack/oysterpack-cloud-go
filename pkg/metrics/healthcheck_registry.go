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

package metrics

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

// HealthCheckRegistry is a HealthCheck registry.
// HealthChecks are registered via one of the NewHealthCheck methods.
type HealthCheckRegistry struct {
	sync.RWMutex
	healthchecks map[string]HealthCheck
}

// NewHealthCheck creates a new HealthCheck.
// check is required - panics if nil.
// The healthcheck metrics are registered. Failing to registering the metrics will trigger a panic.
func (a *HealthCheckRegistry) NewHealthCheck(opts prometheus.GaugeOpts, runInterval time.Duration, check RunHealthCheck) HealthCheck {
	a.Lock()
	defer a.Unlock()

	key := GaugeFQName(&opts)
	if _, exists := a.healthchecks[key]; exists {
		logger.Panic().Msgf("HealthCheck is already registered for : %q", key)
	}

	if check == nil {
		panic("check is required")
	}

	// check for metric collision
	if Registered(key) {
		logger.Panic().Err(ErrMetricAlreadyRegistered).Msg("")
	}

	opts.ConstLabels = addLabels(healthcheckLabelsStatus, opts.ConstLabels)

	h := &healthcheck{
		opts:        opts,
		run:         check,
		status:      GetOrMustRegisterGauge(&opts),
		runInterval: runInterval,
	}

	durationOpts := prometheus.GaugeOpts{
		Namespace:   opts.Namespace,
		Subsystem:   opts.Subsystem,
		Name:        fmt.Sprintf("%s_duration_seconds", opts.Name),
		Help:        "The healthckeck run duration in seconds",
		ConstLabels: addLabels(healthcheckLabelsDuration, opts.ConstLabels),
	}
	// check for metric collision
	if Registered(GaugeFQName(&durationOpts)) {
		logger.Panic().Err(ErrMetricAlreadyRegistered).Msg("")
	}

	h.runDuration = GetOrMustRegisterGauge(&durationOpts)
	h.StartTicker()
	a.healthchecks[key] = h

	return h
}

// NewHealthCheckVector creates a new HealthCheck.
// check is required - panics if nil.
// The healthcheck metrics are registered. Failing to registering the metrics will trigger a panic.
func (a *HealthCheckRegistry) NewHealthCheckVector(opts *GaugeVecOpts, runInterval time.Duration, check RunHealthCheck, labelValues []string) HealthCheck {
	a.Lock()
	defer a.Unlock()

	key := strings.Join([]string{GaugeFQName(opts.GaugeOpts), strings.Join(labelValues, "")}, "")
	if _, exists := a.healthchecks[key]; exists {
		logger.Panic().Msgf("HealthCheck vector is already registered for : %q -> %q", GaugeFQName(opts.GaugeOpts), labelValues)
	}

	if check == nil {
		logger.Panic().Msg("check is required")
	}

	if len(opts.Labels) == 0 {
		logger.Panic().Msgf("Labels are required : %v", GaugeFQName(opts.GaugeOpts))
	}

	if len(labelValues) != len(opts.Labels) {
		logger.Panic().Msgf("The number of label values must match the number of labels defined in the GaugeVecOpts : %v : %v", opts.Labels, labelValues)
	}

	opts.ConstLabels = addLabels(healthcheckLabelsStatus, opts.ConstLabels)

	h := &healthcheck{
		opts:        *opts.GaugeOpts,
		labels:      opts.Labels,
		labelValues: labelValues,
		run:         check,
		status:      GetOrMustRegisterGaugeVec(opts).WithLabelValues(labelValues...),
		runInterval: runInterval,
	}

	durationOpts := &GaugeVecOpts{
		&prometheus.GaugeOpts{
			Namespace:   opts.Namespace,
			Subsystem:   opts.Subsystem,
			Name:        fmt.Sprintf("%s_duration_seconds", opts.Name),
			Help:        "The healthckeck run duration in seconds",
			ConstLabels: addLabels(healthcheckLabelsDuration, opts.ConstLabels),
		}, opts.Labels,
	}

	h.runDuration = GetOrMustRegisterGaugeVec(durationOpts).WithLabelValues(labelValues...)
	h.StartTicker()
	a.healthchecks[key] = h
	return h
}

func addLabels(from prometheus.Labels, to prometheus.Labels) prometheus.Labels {
	if len(to) == 0 {
		return from
	}

	labels := prometheus.Labels{}
	for k, v := range to {
		labels[k] = v
	}
	for k, v := range from {
		labels[k] = v
	}
	return labels
}

// HealthChecks returns all current registered healthchecks
func (a *HealthCheckRegistry) HealthChecks() []HealthCheck {
	a.RLock()
	defer a.RUnlock()
	healthchecks := make([]HealthCheck, len(a.healthchecks))
	i := 0
	for _, healthcheck := range a.healthchecks {
		healthchecks[i] = healthcheck
		i++
	}
	return healthchecks
}

// FailedHealthChecks returns all failed health checks based on the last result.
// If a HealthCheck has not yet been run, then it will be run.
func (a *HealthCheckRegistry) FailedHealthChecks() []HealthCheck {
	a.RLock()
	defer a.RUnlock()
	return a.failedHealthChecks()
}

func (a *HealthCheckRegistry) failedHealthChecks() []HealthCheck {
	failed := []HealthCheck{}
	for _, healthcheck := range a.healthchecks {
		if healthcheck.LastResult() != nil && !healthcheck.LastResult().Success() {
			failed = append(failed, healthcheck)
		}
	}
	return failed
}

// SucceededHealthChecks returns all health checks that succeeded based on the last result.
// If a HealthCheck has not yet been run, then it will be run.
func (a *HealthCheckRegistry) SucceededHealthChecks() []HealthCheck {
	a.RLock()
	defer a.RUnlock()
	succeeded := []HealthCheck{}
	for _, healthcheck := range a.healthchecks {
		if healthcheck.LastResult() != nil && healthcheck.LastResult().Success() {
			succeeded = append(succeeded, healthcheck)
		}
	}
	return succeeded
}

// RunAllHealthChecks runs all registered health checks.
// After a health check is run it is delivered on the channel.
func (a *HealthCheckRegistry) RunAllHealthChecks() <-chan HealthCheck {
	a.RLock()
	defer a.RUnlock()
	count := len(a.healthchecks)
	c := make(chan HealthCheck, count)
	wait := sync.WaitGroup{}
	wait.Add(count)

	for _, healthcheck := range a.healthchecks {
		go func(healthcheck HealthCheck) {
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
func (a *HealthCheckRegistry) RunAllFailedHealthChecks() <-chan HealthCheck {
	a.RLock()
	defer a.RUnlock()
	failedHealthChecks := a.failedHealthChecks()
	c := make(chan HealthCheck, len(failedHealthChecks))
	wait := sync.WaitGroup{}
	wait.Add(len(failedHealthChecks))

	for _, healthcheck := range failedHealthChecks {
		go func(healthcheck HealthCheck) {
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

func (a *HealthCheckRegistry) StopAllHealthCheckTickers() {
	a.RLock()
	defer a.RUnlock()
	for _, c := range a.healthchecks {
		c.StopTicker()
	}
}

func (a *HealthCheckRegistry) StartAllHealthCheckTickers() {
	a.RLock()
	defer a.RUnlock()
	for _, c := range a.healthchecks {
		c.StartTicker()
	}
}

// Clear clears the registry - this is exposed for testing purposes
func (a *HealthCheckRegistry) Clear() {
	a.Lock()
	defer a.Unlock()
	for _, c := range a.healthchecks {
		c.StopTicker()
	}
	a.healthchecks = make(map[string]HealthCheck)
}
