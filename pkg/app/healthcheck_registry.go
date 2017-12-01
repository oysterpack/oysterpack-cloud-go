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

package app

import (
	"sync"

	"time"

	"fmt"

	"github.com/prometheus/client_golang/prometheus"
	"gopkg.in/tomb.v2"
)

var (
	// used to ensure that healtchecks are concurrency safe
	healthchecksMutex      sync.RWMutex
	healthCheckSpecs       map[HealthCheckID]*HealthCheckSpec
	registeredHealthChecks = make(map[HealthCheckID]*registeredHealthCheck)

	// used to ensure that only 1 healthcheck at a time can be run
	healthcheckRunMutex sync.Mutex
)

func init() {
	registerHealthCheckGauges()
	registerHealthCheckService()
	initHealthCheckSpecs()
}

func registerHealthCheckService() {
	healthchecksMutex.Lock()
	defer healthchecksMutex.Unlock()
	healthCheckService, err := Services.Service(HEALTHCHECK_SERVICE_ID)
	if err != nil && err != ErrServiceNotRegistered {
		Logger().Panic().Err(err).Msg("HealthCheckService lookup failed")
	}
	if healthCheckService == nil {
		healthCheckService = NewService(HEALTHCHECK_SERVICE_ID)
		if err := Services.Register(healthCheckService); err != nil {
			panic(fmt.Sprintf("Failed to register HealthCheckService: %v", err))
		}
	}
}

func registerHealthCheckGauges() {
	metricsServiceMutex.Lock()
	defer metricsServiceMutex.Unlock()
	metricSpecs := []*GaugeVectorMetricSpec{
		&GaugeVectorMetricSpec{
			MetricSpec: MetricSpec{
				ServiceID: HEALTHCHECK_SERVICE_ID,
				MetricID:  HEALTHCHECK_METRIC_ID,
				Help:      "HealthChecks",
			},
			DynamicLabels: []string{HEALTHCHECK_METRIC_LABEL},
		},
		&GaugeVectorMetricSpec{
			MetricSpec: MetricSpec{
				ServiceID: HEALTHCHECK_SERVICE_ID,
				MetricID:  HEALTHCHECK_RUN_DURATION_METRIC_ID,
				Help:      "HealthCheck run durations in nanoseconds",
			},
			DynamicLabels: []string{HEALTHCHECK_METRIC_LABEL},
		},
	}

	for _, metricSpec := range metricSpecs {
		if MetricRegistry.GaugeVector(metricSpec.ServiceID, metricSpec.MetricID) == nil {
			metric := &GaugeVectorMetric{metricSpec, prometheus.NewGaugeVec(metricSpec.GaugeOpts(), metricSpec.DynamicLabels)}
			metric.register()
		}
	}
}

func initHealthCheckSpecs() {
	healthchecksMutex.Lock()
	defer healthchecksMutex.Unlock()
	serviceSpec := loadHealthCheckServiceSpec()
	specs, err := serviceSpec.HealthCheckSpecs()
	if err != nil {
		CONFIG_LOADING_ERR.Log(Logger().Panic()).Err(err).Msg("Failed on HealthCheckSpecs")
	}
	if specs.Len() == 0 {
		ZERO_HEALTHCHECKS.Log(Logger().Warn()).Msg("No health checks")
		return
	}

	healthCheckSpecs = make(map[HealthCheckID]*HealthCheckSpec, specs.Len())
	for i := 0; i < specs.Len(); i++ {
		spec := specs.At(i)
		healthCheckSpecs[HealthCheckID(spec.HealthCheckID())] = &HealthCheckSpec{
			HealthCheckID: HealthCheckID(spec.HealthCheckID()),
			RunInterval:   time.Duration(spec.RunIntervalSeconds()) * time.Second,
			Timeout:       time.Duration(spec.TimeoutSeconds()) * time.Second,
		}
		Logger().Info()
	}
}

// Registered returns true if a HealthCheck is registered for the specified HealthCheckID
//
// errors:
//	- ErrServiceNotAlive
func (a AppHealthChecks) Registered(id HealthCheckID) bool {
	healthchecksMutex.RLock()
	defer healthchecksMutex.RUnlock()
	_, registered := registeredHealthChecks[id]
	return registered
}

// HealthCheckSpec returns the HealthCheckSpec for the specified HealthCheckID
//
// errors:
//  - ErrHealthCheckNotRegistered
//	- ErrServiceNotAlive
func (a AppHealthChecks) HealthCheckSpec(id HealthCheckID) *HealthCheckSpec {
	healthchecksMutex.RLock()
	defer healthchecksMutex.RUnlock()
	return healthCheckSpecs[id]
}

// Registers the HealthCheck for the specified HealthCheckID. If a HealthCheck is already registered for the specified
// HealthCheckID, then it will be replaced. The bool result indicates if there was a pre-existing healthcheck that was replaced.
// If true is returned, then it indicates that a healthcheck was replaced, i.e., there was a pre-existing healthcheck registered.
//
// Design notes:
// - each healthcheck is scheduled to run on its own scheduled based on its HealthCheckSpec.RunInterval. The scheduling
//   is performed by a scheduling goroutine.
// - the healthcheck is run within the command server goroutine, which owns the healtchecks, i.e., the command server
//   goroutine ensures that healthchecks are run serially and in a concurrency safe manner.
//
//
// errors
//	- ErrHealthCheckNil
//	- ErrServiceNotAlive
func (a AppHealthChecks) Register(id HealthCheckID, healthCheckFunc HealthCheck) error {
	if id == HealthCheckID(0) {
		return ErrHealthCheckIDZero
	}
	if healthCheckFunc == nil {
		return ErrHealthCheckNil
	}

	healthCheckService, err := Services.Service(HEALTHCHECK_SERVICE_ID)
	if err != nil {
		return err
	}

	healthchecksMutex.Lock()
	defer healthchecksMutex.Unlock()

	// register the healthcheck
	healthcheckEntry := registeredHealthChecks[id]
	if healthcheckEntry != nil { // update the healthcheck entry
		// kill the current healthcheck goroutine because a new one needs to be started using the new HealthCheck func
		healthcheckEntry.Kill(nil)
		healthcheckEntry.Wait()

		healthcheckEntry.HealthCheck = healthCheckFunc
		healthcheckEntry.Tomb = tomb.Tomb{} // resurrect the entry
	} else { // register a new healthcheck entry
		spec := healthCheckSpecs[id]
		if spec == nil {
			spec = &HealthCheckSpec{
				HealthCheckID: id,
				RunInterval:   DEFAULT_HEALTHCHECK_RUN_INTERVAL,
				Timeout:       DEFAULT_HEALTHCHECK_RUN_TIMEOUT,
			}
		}
		healthcheckEntry = &registeredHealthCheck{
			HealthCheckSpec:  spec,
			ResultGauge:      MetricRegistry.GaugeVector(HEALTHCHECK_SERVICE_ID, HEALTHCHECK_METRIC_ID).GaugeVec.WithLabelValues(id.Hex()),
			RunDurationGauge: MetricRegistry.GaugeVector(HEALTHCHECK_SERVICE_ID, HEALTHCHECK_RUN_DURATION_METRIC_ID).GaugeVec.WithLabelValues(id.Hex()),
			HealthCheck:      healthCheckFunc,
			RunOnDemandChan:  make(chan chan HealthCheckResult),
		}
		registeredHealthChecks[id] = healthcheckEntry
	}

	// schedule the healthcheck to run
	healthcheckEntry.Go(func() error {
		ticker := time.NewTicker(healthcheckEntry.HealthCheckSpec.RunInterval)
		defer ticker.Stop()

		for {
			select {
			case <-healthcheckEntry.Dying():
				return nil
			case <-healthCheckService.Dying():
				return nil
			case <-ticker.C:
				healthcheckEntry.run()
			case resultChan := <-healthcheckEntry.RunOnDemandChan:
				healthcheckEntry.run()
				resultChan <- healthcheckEntry.HealthCheckResult
			}
		}
	})

	return nil
}

// HealthCheckIDs returns the HealthCheckID(s) for the healthchecks that are currently registered.
// nil is returned if there are no healthchecks currently registered.
//
// errors
//	- ErrServiceNotAlive
func (a AppHealthChecks) HealthCheckIDs() []HealthCheckID {
	healthchecksMutex.RLock()
	defer healthchecksMutex.RUnlock()
	ids := make([]HealthCheckID, len(registeredHealthChecks))
	i := 0
	for id := range registeredHealthChecks {
		ids[i] = id
		i++
	}
	return ids
}

// Run triggers the healthcheck to run on demand.
//
// errors
//	- ErrServiceNotAlive
//  - ErrHealthCheckNotRegistered
func (a AppHealthChecks) Run(id HealthCheckID) (HealthCheckResult, error) {
	if id == HealthCheckID(0) {
		return HealthCheckResult{}, ErrHealthCheckIDZero
	}

	healthchecksMutex.Lock()
	defer healthchecksMutex.Unlock()

	healthCheckService, err := Services.Service(HEALTHCHECK_SERVICE_ID)
	if err != nil {
		return HealthCheckResult{}, err
	}

	healthcheck := registeredHealthChecks[id]
	if healthcheck == nil {
		return HealthCheckResult{}, ErrHealthCheckNotRegistered
	}

	c := make(chan HealthCheckResult, 1)
	select {
	case healthcheck.RunOnDemandChan <- c:
	case <-healthcheck.Dying():
		return HealthCheckResult{}, ErrHealthCheckNotAlive
	case <-healthCheckService.Dying():
		return HealthCheckResult{}, ErrServiceNotAlive
	}

	select {
	case result := <-c:
		return result, nil
	case <-healthcheck.Dying():
		return HealthCheckResult{}, ErrHealthCheckNotAlive
	case <-healthCheckService.Dying():
		return HealthCheckResult{}, ErrServiceNotAlive
	}
}

// LastResult returns the latest HealthCheckResult from the last time the HealthCheck was run.
//
// errors
//  - ErrHealthCheckNotRegistered
func (a AppHealthChecks) HealthCheckResult(id HealthCheckID) (HealthCheckResult, error) {
	if id == HealthCheckID(0) {
		return HealthCheckResult{}, ErrHealthCheckIDZero
	}

	healthchecksMutex.RLock()
	defer healthchecksMutex.RUnlock()

	healthcheck := registeredHealthChecks[id]
	if healthcheck == nil {
		return HealthCheckResult{}, ErrHealthCheckNotRegistered
	}

	return healthcheck.HealthCheckResult, nil
}

// HealthCheckResults returns a snapshot of all current HealthCheckResult(s)
func (a AppHealthChecks) HealthCheckResults() map[HealthCheckID]HealthCheckResult {
	healthchecksMutex.RLock()
	defer healthchecksMutex.RUnlock()

	results := make(map[HealthCheckID]HealthCheckResult, len(registeredHealthChecks))
	for _, healthcheck := range registeredHealthChecks {
		results[healthcheck.HealthCheckID] = healthcheck.HealthCheckResult
	}

	return results
}
