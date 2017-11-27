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
)

// healthcheck constants
const (
	HEALTHCHECK_SERVICE_ID = ServiceID(0xe105d0909fcc5ec5)

	HEALTHCHECK_METRIC_LABEL = "healthcheck"

	HEALTHCHECK_METRIC_ID              = MetricID(0x844d7830332bffd3)
	HEALTHCHECK_RUN_DURATION_METRIC_ID = MetricID(0xcd4260d6e89ad9c6)
)

// HealthcheckSpec maps a healthcheck to a metric gauge vector.
// The healthcheck id is used as the healthcheck label value.
//
// If the health gauge value indicates number of consecutive failures.
// Thus, a gauge value of zero means the least time the health check ran, it succeeded.
// A gauge value of 3 means the healthcheck has failed the last 3 times it ran.
// A gauge value of -1 means the healthcheck has not yet been run.
//
// The health check is run via the specified interval and the latest health check result is kept.
// When metrics are collected, the latest health result will be used to report the metric.
// It is the application's responsibility to register HealthCheck functions for the registered HealthCheck(s).
//
// The health check run duration is also recorded as a gauge metric.
// This can help identify health checks that are taking too long to execute.
// The run duration may also be used to configure alerts. For example, health checks that are passing but taking longer
// to run may be an early warning sign.
//
type HealthcheckSpec struct {
	HealthCheckID
	*GaugeVectorMetricSpec
	RunInterval time.Duration
}

type registeredHealthCheck struct {
	*HealthcheckSpec
	HealthCheck
}

// HealthCheck is used to run the health check.
// If the health check fails, then an error is returned.
type HealthCheck func() error

// HealthCheckResult is the result of running the health check
type HealthCheckResult struct {
	// why the health check failed
	Err error

	// when the health check started run
	time.Time

	// how long it took to run the health check
	time.Duration
}

type HealthCheckService struct {
	*CommandServer
	healthchecks map[HealthCheckID]*registeredHealthCheck
}

func (a *HealthCheckService) init() {
	cfg, err := Configs.Config(HEALTHCHECK_SERVICE_ID)
	if err != nil {
		CONFIG_LOADING_ERR.Log(a.Logger().Fatal()).Err(err).Msg("")
	}
	if cfg == nil {
		return
	}
	serviceSpec, err := config.ReadRootHealthCheckServiceSpec(cfg)
	if err != nil {
		CONFIG_LOADING_ERR.Log(a.Logger().Fatal()).Err(err).Msg("config.ReadRootHealthCheckServiceSpec() failed")
	}

	a.healthchecks = make(map[HealthCheckID]*registeredHealthCheck)
	healthcheckSpecs, err := serviceSpec.HealthCheckSpecs()
	if err != nil {
		CONFIG_LOADING_ERR.Log(a.Logger().Fatal()).Err(err).Msg("Failed on HealthCheckSpecs")
	}
	if healthcheckSpecs.Len() == 0 {
		ZERO_HEALTHCHECKS.Log(a.Logger().Warn()).Msg("No health checks")
		return
	}
	// TODO
}
