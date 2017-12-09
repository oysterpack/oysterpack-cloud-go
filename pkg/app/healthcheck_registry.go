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
	"github.com/rs/zerolog"
	"gopkg.in/tomb.v2"
)

var (
	// used to ensure that healtchecks are concurrency safe
	healthchecksMutex      sync.RWMutex
	healthCheckSpecs       map[HealthCheckID]*HealthCheckSpec
	registeredHealthChecks map[HealthCheckID]*registeredHealthCheck

	// used to ensure that only 1 healthcheck at a time can be run
	healthcheckRunMutex sync.Mutex
)

func initHealthCheckService() {
	registeredHealthChecks = make(map[HealthCheckID]*registeredHealthCheck)
	registerHealthCheckGauges()
	registerHealthCheckService()
	initHealthCheckSpecs()
}

func registerHealthCheckService() {
	healthchecksMutex.Lock()
	defer healthchecksMutex.Unlock()
	healthCheckService, err := Services.Service(HEALTHCHECK_SERVICE_ID)
	if err != nil {
		if !IsError(err, ErrSpec_ServiceNotRegistered.ErrorID) {
			Logger().Panic().Err(err).Msg("HealthCheckService is already registered")
		}
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
		return IllegalArgumentError("HealthCheckID cannot be 0")
	}
	if healthCheckFunc == nil {
		return IllegalArgumentError("HealthCheck function is required")
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
			HealthCheckSpec:    spec,
			ResultGauge:        MetricRegistry.GaugeVector(HEALTHCHECK_SERVICE_ID, HEALTHCHECK_METRIC_ID).GaugeVec.WithLabelValues(id.Hex()),
			RunDurationGauge:   MetricRegistry.GaugeVector(HEALTHCHECK_SERVICE_ID, HEALTHCHECK_RUN_DURATION_METRIC_ID).GaugeVec.WithLabelValues(id.Hex()),
			HealthCheck:        healthCheckFunc,
			RunOnDemandChan:    make(chan chan HealthCheckResult),
			HealthCheckService: healthCheckService,
		}
		registeredHealthChecks[id] = healthcheckEntry
		HEALTHCHECK_REGISTERED.Log(healthCheckService.Logger().Info()).
			Uint64(HEALTHCHECK_ID_LOG_FIELD, uint64(id)).
			Dict("spec", zerolog.Dict().Dur("run-interval", spec.RunInterval).Dur("timeout", spec.Timeout)).
			Msg("registered")
	}

	scheduleHealthCheck(healthcheckEntry)

	return nil
}

func scheduleHealthCheck(healthcheck *registeredHealthCheck) {
	healthcheck.Go(func() error {
		ticker := time.NewTicker(healthcheck.HealthCheckSpec.RunInterval)
		defer ticker.Stop()

		for {
			select {
			case <-healthcheck.Dying():
				return nil
			case <-healthcheck.HealthCheckService.Dying():
				return nil
			case <-ticker.C:
				healthcheck.run()
			case resultChan := <-healthcheck.RunOnDemandChan:
				healthcheck.run()
				resultChan <- healthcheck.HealthCheckResult
			}
		}
	})
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
		return HealthCheckResult{}, IllegalArgumentError("HealthCheckIS cannot be 0")
	}

	healthchecksMutex.Lock()
	defer healthchecksMutex.Unlock()

	healthCheckService, err := Services.Service(HEALTHCHECK_SERVICE_ID)
	if err != nil {
		return HealthCheckResult{}, err
	}

	healthcheck := registeredHealthChecks[id]
	if healthcheck == nil {
		return HealthCheckResult{}, HealthCheckNotRegisteredError(id)
	}

	c := make(chan HealthCheckResult, 1)
	select {
	case healthcheck.RunOnDemandChan <- c:
	case <-healthcheck.Dying():
		return HealthCheckResult{}, HealthCheckNotAliveError(id)
	case <-healthCheckService.Dying():
		return HealthCheckResult{}, ServiceNotAliveError(HEALTHCHECK_SERVICE_ID)
	}

	select {
	case result := <-c:
		return result, nil
	case <-healthcheck.Dying():
		return HealthCheckResult{}, HealthCheckNotAliveError(id)
	case <-healthCheckService.Dying():
		return HealthCheckResult{}, ServiceNotAliveError(HEALTHCHECK_SERVICE_ID)
	}
}

// LastResult returns the latest HealthCheckResult from the last time the HealthCheck was run.
//
// errors
//  - ErrHealthCheckNotRegistered
func (a AppHealthChecks) HealthCheckResult(id HealthCheckID) (HealthCheckResult, error) {
	if id == HealthCheckID(0) {
		return HealthCheckResult{}, IllegalArgumentError("HealthCheckIS cannot be 0")
	}

	healthchecksMutex.RLock()
	defer healthchecksMutex.RUnlock()

	healthcheck := registeredHealthChecks[id]
	if healthcheck == nil {
		return HealthCheckResult{}, HealthCheckNotRegisteredError(id)
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

// PausedHealthChecks returns HealthCheckID(s) for registered healthchecks that are not currently scheduled to run, i.e., not alive.
func (a AppHealthChecks) PausedHealthChecks() []HealthCheckID {
	healthchecksMutex.RLock()
	defer healthchecksMutex.RUnlock()
	ids := []HealthCheckID{}
	for _, healthcheck := range registeredHealthChecks {
		if !healthcheck.Alive() {
			ids = append(ids, healthcheck.HealthCheckID)
		}
	}
	return ids
}

// PauseHealthCheck will kill the healthcheck goroutine.
// false is returned if no HealthCheck is registered for the specified HealthCheckID
func (a AppHealthChecks) PauseHealthCheck(id HealthCheckID) bool {
	healthchecksMutex.Lock()
	defer healthchecksMutex.Unlock()
	healthCheck := registeredHealthChecks[id]
	if healthCheck == nil {
		return false
	}
	healthCheck.Kill(nil)
	HEALTHCHECK_PAUSED.Log(healthCheck.HealthCheckService.Logger().Info()).Uint64(HEALTHCHECK_ID_LOG_FIELD, uint64(id)).Msg("paused")
	return true
}

// ResumeHealthCheck will schedule the healthcheck to run, if it is currently pause.
//
// errors
//	- ErrHealthCheckNotRegistered
//	- ErrServiceNotAlive
//	- ErrAppNotAlive
//	- HealthCheckKillTimeoutError
func (a AppHealthChecks) ResumeHealthCheck(id HealthCheckID) error {
	healthchecksMutex.Lock()
	defer healthchecksMutex.Unlock()
	healthcheck := registeredHealthChecks[id]
	if healthcheck == nil {
		return HealthCheckNotRegisteredError(id)
	}
	if healthcheck.Alive() {
		return nil
	}

	if !healthcheck.HealthCheckService.Alive() {
		return ServiceNotAliveError(HEALTHCHECK_SERVICE_ID)
	}
	if !app.Alive() {
		return AppNotAliveError(HEALTHCHECK_SERVICE_ID)
	}

	ticker := time.NewTicker(time.Second * 10)
	defer ticker.Stop()
	for {
		select {
		case <-healthcheck.Dead():
			registeredHealthChecks[id] = &registeredHealthCheck{
				HealthCheckSpec:    healthcheck.HealthCheckSpec,
				ResultGauge:        healthcheck.ResultGauge,
				RunDurationGauge:   healthcheck.RunDurationGauge,
				HealthCheck:        healthcheck.HealthCheck,
				RunOnDemandChan:    healthcheck.RunOnDemandChan,
				HealthCheckService: healthcheck.HealthCheckService,
			}
			scheduleHealthCheck(registeredHealthChecks[id])
			HEALTHCHECK_RESUMED.Log(healthcheck.HealthCheckService.Logger().Info()).Uint64(HEALTHCHECK_ID_LOG_FIELD, uint64(id)).Msg("resumed")
			return nil
		case <-healthcheck.HealthCheckService.Dying():
			return ServiceNotAliveError(HEALTHCHECK_SERVICE_ID)
		case <-ticker.C:
			return HealthCheckKillTimeoutError(id)
		}
	}
}
