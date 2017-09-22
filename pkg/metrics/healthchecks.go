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

// HealthCheck represents a health check metric - it maps to a prometheus gauge, where
// 	0 = FAIL
// 	1 = PASS
//
// The unique identifier for a HealthCheck is the combination of name and labels.
//
// Metrics are also collected for the healthchecks themselves :
// 1. health check run time -> gauge
// 2. health check run counter
type HealthCheck interface {
	// Name is the base health check name
	Name() string

	// LabelSet are used to label health check instances.
	// For example, a database health check could have the following named labels : 'db', 'schema'
	Labels() map[string]string

	Key() HealthCheckKey

	// Help provides information about this health check
	Help() string

	// Run executes the health check and updates the health check gauge
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

	// MustRegister panics if registration fails
	MustRegister(registerer prometheus.Registerer)
	// Register returns an error if registration fails
	Register(registerer prometheus.Registerer) error
}

// HealthCheckKey health check unique metric key
type HealthCheckKey struct {
	Name   string
	Labels map[string]string
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

// HealthCheckGaugeValue represents an enum for HealthCheck gauge values
type HealthCheckGaugeValue float64

const (
	// HEALTHCHECK_FAILURE = 0
	HEALTHCHECK_FAILURE HealthCheckGaugeValue = 0
	// HEALTHCHECK_SUCCESS = 1
	HEALTHCHECK_SUCCESS HealthCheckGaugeValue = 1
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

// Value maps the HealthCheckResult to a gauge value
func (a *HealthCheckResult) Value() HealthCheckGaugeValue {
	if a.Success() {
		return HEALTHCHECK_SUCCESS
	}
	return HEALTHCHECK_FAILURE
}

type healthcheck struct {
	opts prometheus.GaugeOpts

	sync.Mutex
	run RunHealthCheck

	gauge         prometheus.Gauge
	durationGauge prometheus.Gauge
	runCounter    prometheus.Counter

	runInterval time.Duration
	ticker      *time.Ticker
	stopTrigger chan struct{}

	lastResult *HealthCheckResult
}

func (a *healthcheck) Key() HealthCheckKey {
	return HealthCheckKey{a.Name(), a.opts.ConstLabels}
}

func (a HealthCheckKey) String() string {
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
		return fmt.Sprintf("%v[%v]", a.Name, labels)
	}
	return a.Name
}

func (a *healthcheck) Name() string {
	return prometheus.BuildFQName(a.opts.Namespace, a.opts.Subsystem, a.opts.Name)
}

func (a *healthcheck) Labels() map[string]string {
	return a.opts.ConstLabels
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

		a.gauge.Set(float64(result.Value()))
		a.durationGauge.Set(result.Duration.Seconds())
		a.runCounter.Inc()

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
	return a.lastResult
}

func (a *healthcheck) String() string {
	if a.lastResult != nil {
		return fmt.Sprintf("%v : %v : %v : %v", a.Key(), a.Help(), a.runCounter, a.lastResult.Success())
	}
	return fmt.Sprintf("%v : %v", a.Key(), a.Help())
}

// MustRegister panics if registration fails
func (a *healthcheck) MustRegister(registerer prometheus.Registerer) {
	registerer.MustRegister(a.gauge, a.durationGauge, a.runCounter)
}

// Register returns an error if registration fails
func (a *healthcheck) Register(registerer prometheus.Registerer) error {
	if err := registerer.Register(a.gauge); err != nil {
		return err
	}
	if err := registerer.Register(a.durationGauge); err != nil {
		return err
	}
	return registerer.Register(a.runCounter)
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
	if a.runInterval <= 0 {
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
	return a.ticker != nil
}

// NewHealthCheck creates a new HealthCheck.
// check is required - panics if nil.
func NewHealthCheck(opts prometheus.GaugeOpts, runInterval time.Duration, check RunHealthCheck) HealthCheck {
	if check == nil {
		panic("check is required")
	}

	counterOpts := prometheus.CounterOpts{
		Namespace:   opts.Namespace,
		Subsystem:   opts.Subsystem,
		Name:        fmt.Sprintf("%s_run_count", opts.Name),
		Help:        "healthcheck run count",
		ConstLabels: opts.ConstLabels,
	}

	a := &healthcheck{
		opts:        opts,
		run:         check,
		gauge:       prometheus.NewGauge(opts),
		runCounter:  prometheus.NewCounter(counterOpts),
		runInterval: runInterval,
	}
	opts.Name = fmt.Sprintf("%s_duration_seconds", opts.Name)
	opts.Help = "healthcheck run duration"
	a.durationGauge = prometheus.NewGauge(opts)
	a.StartTicker()
	return a
}
