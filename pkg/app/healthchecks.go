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

	"github.com/oysterpack/oysterpack.go/pkg/app/config"
	"github.com/prometheus/client_golang/prometheus"
	"gopkg.in/tomb.v2"
	"zombiezen.com/go/capnproto2"
)

// healthcheck constants
const (
	HEALTHCHECK_SERVICE_ID = ServiceID(0xe105d0909fcc5ec5)

	HEALTHCHECK_METRIC_LABEL           = "healthcheck" // the label value will be the HealthCheckID in hex format
	HEALTHCHECK_METRIC_ID              = MetricID(0x844d7830332bffd3)
	HEALTHCHECK_RUN_DURATION_METRIC_ID = MetricID(0xcd4260d6e89ad9c6)

	DEFAULT_HEALTHCHECK_RUN_INTERVAL = 5 * time.Minute
	DEFAULT_HEALTHCHECK_RUN_TIMEOUT  = 5 * time.Second

	HEALTHCHECK_ID_LOG_FIELD = "healthcheck"
)

type AppHealthChecks struct{}

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

	HealthCheckResult

	ResultGauge      prometheus.Gauge
	RunDurationGauge prometheus.Gauge

	HealthCheckService *Service
	// each healthcheck is scheduled to run in it own separate goroutine
	// the tomb is used to stop an individual healthcheck
	tomb.Tomb

	// used to run the healthcheck on demand. The result is returned on the specified chan
	RunOnDemandChan chan chan HealthCheckResult
}

func (a *registeredHealthCheck) run() {
	healthcheckRunMutex.Lock()
	defer healthcheckRunMutex.Unlock()

	if !a.Alive() || !a.HealthCheckService.Alive() {
		return
	}

	cancel := make(chan struct{})
	result := make(chan error, 1)

	start := time.Now()
	go a.HealthCheck(result, cancel)

	timeoutTimer := time.NewTicker(a.Timeout)
	defer timeoutTimer.Stop()

	select {
	case <-timeoutTimer.C:
		a.HealthCheckResult.Err = HealthCheckTimeoutError(a.HealthCheckSpec.HealthCheckID)
		a.HealthCheckResult.Duration = time.Now().Sub(start)
		a.HealthCheckResult.ErrCount++
		a.ResultGauge.Inc()
		a.RunDurationGauge.Set(float64(a.HealthCheckResult.Duration))
		close(cancel) // notifies the HealthCheck func that it has been cancelled
		a.logHealthCheckResult()
	case err := <-result:
		a.HealthCheckResult.Err = err
		a.HealthCheckResult.Time = start
		a.HealthCheckResult.Duration = time.Now().Sub(start)
		if err != nil {
			a.HealthCheckResult.ErrCount++
		} else {
			a.HealthCheckResult.ErrCount = 0
		}
		a.ResultGauge.Set(float64(a.HealthCheckResult.ErrCount))
		a.RunDurationGauge.Set(float64(a.HealthCheckResult.Duration))
		a.logHealthCheckResult()
	case <-a.Dying():
		close(cancel) // notifies the HealthCheck func that it has been cancelled
		return
	case <-a.HealthCheckService.Dying():
		close(cancel) // notifies the HealthCheck func that it has been cancelled
		return
	}
}

func (a *registeredHealthCheck) logHealthCheckResult() {
	if a.HealthCheckResult.Err != nil {
		HEALTHCHECK_RESULT.Log(a.HealthCheckService.Logger().Error()).
			Err(a.HealthCheckResult.Err).
			Uint64(HEALTHCHECK_ID_LOG_FIELD, uint64(a.HealthCheckID)).
			Time("start", a.Time).
			Dur("duration", a.Duration).
			Uint("err-count", a.ErrCount).
			Msg("")
	} else {
		HEALTHCHECK_RESULT.Log(a.HealthCheckService.Logger().Info()).
			Uint64(HEALTHCHECK_ID_LOG_FIELD, uint64(a.HealthCheckID)).
			Time("start", a.Time).
			Dur("duration", a.Duration).
			Msg("")
	}
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

	// when the health check started running
	time.Time

	// how long it took to run the health check
	time.Duration

	// how many times the health check has failed consecutively
	ErrCount uint
}

func loadHealthCheckServiceSpec() config.HealthCheckServiceSpec {
	cfg, err := Configs.Config(HEALTHCHECK_SERVICE_ID)
	if err != nil {
		CONFIG_LOADING_ERR.Log(Logger().Fatal()).Err(err).Msg("config.HealthCheckServiceSpec")
	}
	if cfg == nil {
		_, seg, err := capnp.NewMessage(capnp.SingleSegment(nil))
		if err != nil {
			CONFIG_LOADING_ERR.Log(Logger().Panic()).Err(err).Msg("capnp.NewMessage(capnp.SingleSegment(nil)) failed")
		}
		spec, err := config.NewRootHealthCheckServiceSpec(seg)
		spec.SetCommandServerChanSize(1)
		return spec
	}
	serviceSpec, err := config.ReadRootHealthCheckServiceSpec(cfg)
	if err != nil {
		CONFIG_LOADING_ERR.Log(Logger().Panic()).Err(err).Msg("config.ReadRootHealthCheckServiceSpec() failed")
	}
	return serviceSpec
}
