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
	"time"

	"sync"

	"bytes"
	"sort"

	"github.com/oysterpack/oysterpack.go/pkg/commons"
	"github.com/oysterpack/oysterpack.go/pkg/logging"
	"github.com/prometheus/client_golang/prometheus"
)

var (
	// HealthChecks is the global HealthCheck registry
	HealthChecks = &HealthCheckRegistry{
		healthchecks: make(map[string]HealthCheck),
	}
)

// HealthCheck represents a health check metric - it maps to a prometheus status, where
// 	0 = FAIL
// 	1 = PASS
//
// The unique identifier for a HealthCheck is the combination of name and labels.
//
// The health check run duration is also recorded as a gauge metric.
// This can help identify health checks that are taking too long to execute.
// The run duration may also be used to configure alerts. For example, health checks that are passing but taking longer
// to run may be an early warning sign.
//
// The naming convention for the run duration gauge metric is : {health_check_name}_duration_seconds
type HealthCheck interface {
	// Name is the base health check name
	Name() string

	// LabelSet are used to label health check instances.
	// For example, a database health check could have the following named labels : 'db', 'schema'
	Labels() map[string]string

	Key() HealthCheckKey

	// Help provides information about this health check
	Help() string

	// Run executes the health check and updates the health check status
	// It is safe to run it concurrently - it is protected by a mutex.
	Run() *HealthCheckResult

	// LastResult returns the result from the latest run
	LastResult() *HealthCheckResult

	// RunInterval
	RunInterval() time.Duration

	// StopTicker stops the ticker used to run the healthcheck periodically - applies only if run interval > 0
	StopTicker()
	// StartTicker starts the ticker (if it is not already started) used to run the healthcheck periodically - applies only if run interval > 0
	StartTicker()
	// Scheduled returns true if the healthcheck is scheduled to run
	Scheduled() bool
}

// HealthCheckKey health check unique metric key
type HealthCheckKey struct {
	Name   string
	Labels map[string]string

	// used to cache the key
	key string
}

func (a HealthCheckKey) String() string {
	if a.key != "" {
		return a.key
	}
	if len(a.Labels) > 0 {
		keys := []string{}
		for k := range a.Labels {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		buf := bytes.Buffer{}
		for _, key := range keys {
			buf.WriteString(key)
			buf.WriteString("=")
			buf.WriteString(a.Labels[key])
			buf.WriteString(" ")
		}
		labels := string(buf.Bytes()[:len(buf.Bytes())-1])
		a.key = fmt.Sprintf("%v[%v]", a.Name, labels)
		return a.key
	}
	return a.Name
}

// RunHealthCheck is used to run the health check.
// If the health check fails, then an error is returned.
type RunHealthCheck func() error

// HealthCheckResult is the result of running the health check
type HealthCheckResult struct {
	// why the health check failed
	Err error

	// when the health check started run
	time.Time

	// how long it took to run the health check
	time.Duration
}

// HealthCheckGaugeValue represents an enum for HealthCheck status values
type HealthCheckGaugeValue float64

const (
	// HEALTHCHECK_FAILURE = 0
	HEALTHCHECK_FAILURE HealthCheckGaugeValue = 0
	// HEALTHCHECK_SUCCESS = 1
	HEALTHCHECK_SUCCESS HealthCheckGaugeValue = 1

	// HEALTHCHECK_LABEL is used to tag metrics as healthchecks. It is added as a constant label to healthcheck related
	// metrics. The label value indicates what type of healthcheck metric. Valid values are :
	//
	//	status -> heathcheck status that records the healthcheck status
	//	count -> healthcheck run counter
	//	duration -> healthchecl run duration
	HEALTHCHECK_LABEL = "healthcheck"
)

var (
	// marks the healthcheck status metric, which reports the status for the latest healthcheck run
	healthcheckLabelsStatus = prometheus.Labels{HEALTHCHECK_LABEL: "status"}
	// marks the healthcheck run duration metric
	healthcheckLabelsDuration = prometheus.Labels{HEALTHCHECK_LABEL: "duration"}
)

// Success returns true if there was no error
func (a *HealthCheckResult) Success() bool {
	return a.Err == nil
}

func (a *HealthCheckResult) String() string {
	if a.Success() {
		return fmt.Sprintf("PASS : %v : %v ", a.Time, a.Duration)
	}
	return fmt.Sprintf("FAIL : %v : %v : %v", a.Time, a.Duration, a.Err)
}

// Value maps the HealthCheckResult to a status value
func (a *HealthCheckResult) Value() HealthCheckGaugeValue {
	if a.Success() {
		return HEALTHCHECK_SUCCESS
	}
	return HEALTHCHECK_FAILURE
}

type healthcheck struct {
	opts prometheus.GaugeOpts

	// for vector based healthcheck metrics
	labels      []string
	labelValues []string

	sync.RWMutex
	run RunHealthCheck

	status      prometheus.Gauge
	runDuration prometheus.Gauge

	runInterval time.Duration
	ticker      *time.Ticker
	stopTrigger chan struct{}

	lastResult *HealthCheckResult
}

func (a *healthcheck) Key() HealthCheckKey {
	return HealthCheckKey{Name: a.Name(), Labels: a.Labels()}
}

func (a *healthcheck) Name() string {
	return prometheus.BuildFQName(a.opts.Namespace, a.opts.Subsystem, a.opts.Name)
}

func (a *healthcheck) Labels() map[string]string {
	labels := map[string]string{}
	for k, v := range a.opts.ConstLabels {
		labels[k] = v
	}
	for i, label := range a.labels {
		labels[label] = a.labelValues[i]
	}
	return labels
}

func (a *healthcheck) Help() string {
	return a.opts.Help
}

func (a *healthcheck) Run() (result *HealthCheckResult) {
	a.Lock()
	defer a.Unlock()

	start := time.Now()

	defer func() {
		if p := recover(); p != nil {
			result = &HealthCheckResult{
				Err:      fmt.Errorf("HealthCheck panic : %v", p),
				Time:     start,
				Duration: time.Since(start),
			}
		}

		if result.Success() {
			logger.Info().Str(logging.HEALTHCHECK, a.Key().String()).Dur("duration", result.Duration).Msg("")
		} else {
			logger.Error().Str(logging.HEALTHCHECK, a.Key().String()).Dur("duration", result.Duration).Err(result.Err).Msg("")
		}

		a.status.Set(float64(result.Value()))
		a.runDuration.Set(result.Duration.Seconds())

		a.lastResult = result
	}()

	result = &HealthCheckResult{
		Err:      a.run(),
		Time:     start,
		Duration: time.Since(start),
	}
	return
}

func (a *healthcheck) LastResult() *HealthCheckResult {
	a.RLock()
	defer a.RUnlock()
	return a.lastResult
}

func (a *healthcheck) String() string {
	a.RLock()
	defer a.RUnlock()
	if a.lastResult != nil {
		return fmt.Sprintf("%v : %v : %v", a.Key(), a.Help(), a.lastResult.Success())
	}
	return fmt.Sprintf("%v : %v", a.Key(), a.Help())
}

// RunInterval if 0, then the healthcheck is not run on an interval
func (a *healthcheck) RunInterval() time.Duration {
	return a.runInterval
}

func (a *healthcheck) StopTicker() {
	a.Lock()
	defer a.Unlock()
	if a.ticker != nil {
		a.ticker.Stop()
		commons.CloseQuietly(a.stopTrigger)
	}
	a.ticker = nil
	a.stopTrigger = nil
}

func (a *healthcheck) StartTicker() {
	if a.RunInterval() <= 0 {
		return
	}
	a.Lock()
	defer a.Unlock()
	if a.ticker == nil {
		a.ticker = time.NewTicker(a.runInterval)
		a.stopTrigger = make(chan struct{})
		go func(ticker *time.Ticker, stopTrigger <-chan struct{}) {
			for {
				select {
				case <-ticker.C:
					a.Run()
				case <-stopTrigger:
					return
				}
			}
		}(a.ticker, a.stopTrigger)
	}
}

func (a *healthcheck) Scheduled() bool {
	a.RLock()
	defer a.RUnlock()
	return a.ticker != nil
}

// NewHealthCheck creates a new HealthCheck.
// check is required - panics if nil.
// The healthcheck metrics are registered. Failing to registering the metrics will trigger a panic.
func NewHealthCheck(opts prometheus.GaugeOpts, runInterval time.Duration, check RunHealthCheck) HealthCheck {
	return HealthChecks.NewHealthCheck(opts, runInterval, check)
}

// NewHealthCheckVector creates a new HealthCheck.
// check is required - panics if nil.
// The healthcheck metrics are registered. Failing to registering the metrics will trigger a panic.
func NewHealthCheckVector(opts *GaugeVecOpts, runInterval time.Duration, check RunHealthCheck, labelValues []string) HealthCheck {
	return HealthChecks.NewHealthCheckVector(opts, runInterval, check, labelValues)
}
