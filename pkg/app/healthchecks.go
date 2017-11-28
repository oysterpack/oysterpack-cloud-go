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
	"time"

	"sync"

	"github.com/oysterpack/oysterpack.go/pkg/app/config"
	"github.com/prometheus/client_golang/prometheus"
)

// healthcheck constants
const (
	HEALTHCHECK_SERVICE_ID = ServiceID(0xe105d0909fcc5ec5)

	HEALTHCHECK_METRIC_LABEL           = "healthcheck" // the label value will be the HealthCheckID in hex format
	HEALTHCHECK_METRIC_ID              = MetricID(0x844d7830332bffd3)
	HEALTHCHECK_RUN_DURATION_METRIC_ID = MetricID(0xcd4260d6e89ad9c6)
)

var (
	healthchecks = &healthcheckService{}
)

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
				Help:      "HealthCheck run durations",
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

// HealthCheckSpec is used to configure the healthchecks. Configuration that is loaded will override any runtime configuration.
// Consider runtime settings specified in the code as the default settings.
//
// All healthchecks are recorded under a GaugeVector (HEALTHCHECK_METRIC_ID).
// The healthcheck id is used as the healthcheck label value.
//
// The health gauge value indicates the number of consecutive failures.
// Thus, a gauge value of zero means the last time the health check ran, it succeeded.
// A gauge value of 3 means the healthcheck has failed the last 3 times it ran.
// A gauge value of -1 means the healthcheck has not yet been run.
//
// The health check is run via the specified interval and the latest health check result is kept.
// When metrics are collected, the latest health result will be used to report the metric.
// It is the application's responsibility to register HealthCheck functions.
//
// The health check run duration is also recorded as a gauge metric (HEALTHCHECK_RUN_DURATION_METRIC_ID).
// This can help identify health checks that are taking too long to execute.
// The run duration may also be used to configure alerts. For example, health checks that are passing but taking longer
// to run may be an early warning sign.
//
type HealthCheckSpec struct {
	HealthCheckID
	RunInterval time.Duration
	Timeout     time.Duration
}

type registeredHealthCheck struct {
	*HealthCheckSpec
	HealthCheck
	prometheus.Gauge
	HealthCheckResult
}

// HealthCheck is used to run the health check.
// params:
// 	- result : the healthcheck will close the channel to signal success. If the healthcheck fails, then an error is sent on the channel.
//	- cancel : when the cancel channel is closed, then it signals to the healthcheck that it has timed out.
type HealthCheck func(result chan<- error, cancel <-chan struct{})

// HealthCheckResult is the result of running the health check
type HealthCheckResult struct {
	// why the health check failed
	Err error

	// when the health check start time
	time.Time

	// how long it took to run the health check
	time.Duration
}

type healthcheckService struct {
	mutex        sync.Mutex
	server       *CommandServer
	config       map[HealthCheckID]*HealthCheckSpec
	healthchecks map[HealthCheckID]*registeredHealthCheck
}

func (a *healthcheckService) init() {
	a.mutex.Lock()
	defer a.mutex.Unlock()
	if a.server != nil && a.server.Alive() {
		return
	}
	service := NewService(HEALTHCHECK_SERVICE_ID)
	if err := Services.Register(service); err != nil {
		SERVICE_STARTING.Log(Logger().Panic()).Err(err).Msg("Failed to register the HealthCheck service")
	}
	if server, err := NewCommandServer(service, 1, "healthchecks", nil, nil); err != nil {
		SERVICE_STARTING.Log(Logger().Panic()).Err(err).Msg("Failed to create CommandServer")
	} else {
		a.server = server
	}

	a.healthchecks = make(map[HealthCheckID]*registeredHealthCheck)
	registerHealthCheckGauges()
	a.server.Submit(a.loadConfig)
}

func (a *healthcheckService) loadConfig() {
	cfg, err := Configs.Config(HEALTHCHECK_SERVICE_ID)
	if err != nil {
		CONFIG_LOADING_ERR.Log(a.server.Logger().Fatal()).Err(err).Msg("")
	}
	if cfg == nil {
		return
	}
	serviceSpec, err := config.ReadRootHealthCheckServiceSpec(cfg)
	if err != nil {
		CONFIG_LOADING_ERR.Log(a.server.Logger().Panic()).Err(err).Msg("config.ReadRootHealthCheckServiceSpec() failed")
	}

	healthcheckSpecs, err := serviceSpec.HealthCheckSpecs()
	if err != nil {
		CONFIG_LOADING_ERR.Log(a.server.Logger().Panic()).Err(err).Msg("Failed on HealthCheckSpecs")
	}
	if healthcheckSpecs.Len() == 0 {
		ZERO_HEALTHCHECKS.Log(a.server.Logger().Warn()).Msg("No health checks")
		return
	}

	a.config = make(map[HealthCheckID]*HealthCheckSpec, healthcheckSpecs.Len())
	for i := 0; i < healthcheckSpecs.Len(); i++ {
		spec := healthcheckSpecs.At(i)
		a.config[HealthCheckID(spec.HealthCheckID())] = &HealthCheckSpec{
			HealthCheckID: HealthCheckID(spec.HealthCheckID()),
			RunInterval:   time.Duration(spec.RunIntervalSeconds()) * time.Second,
			Timeout:       time.Duration(spec.TimeoutSeconds()) * time.Second,
		}
		a.server.Logger().Info()
	}
}

func (a *healthcheckService) Register(id HealthCheckID, healthcheck HealthCheck) error {
	// TODO
	return nil
}

func (a *healthcheckService) RegisteredHealthCheckIDs() []HealthCheckID {
	// TODO
	return nil
}

func (a *healthcheckService) Run(id HealthCheckID) (HealthCheckResult, bool) {
	//TODO
	return HealthCheckResult{}, false
}

func (a *healthcheckService) LastResult(id HealthCheckID) (HealthCheckResult, bool) {
	return HealthCheckResult{}, false
}
