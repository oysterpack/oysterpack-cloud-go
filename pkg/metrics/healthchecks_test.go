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

package metrics_test

import (
	"testing"
	"time"

	"errors"

	"fmt"

	"github.com/oysterpack/oysterpack.go/pkg/metrics"
	"github.com/prometheus/client_golang/prometheus"
)

func TestRegisterHealthCheck_WithNoLabels(t *testing.T) {
	defer metrics.ResetRegistry()
	var ping metrics.RunHealthCheck = func() error {
		return nil
	}

	opts := prometheus.GaugeOpts{
		Name: "ping",
		Help: "ping always succeeds",
	}

	pingCheck := metrics.NewHealthCheck(opts, 0, ping)
	t.Log(pingCheck)

	t.Log(pingCheck)
	if pingCheck.Name() != prometheus.BuildFQName(opts.Namespace, opts.Subsystem, opts.Name) {
		t.Errorf("%v != %v", pingCheck.Name(), prometheus.BuildFQName(opts.Namespace, opts.Subsystem, opts.Name))
	}
	result := pingCheck.Run()
	if result.Time != pingCheck.LastResult().Time {
		t.Errorf("Last result did not match : %v != %v", result, pingCheck.LastResult())
	}

}

func TestRegisterHealthCheck_CheckMetricsAreRegistered(t *testing.T) {
	defer metrics.ResetRegistry()
	var ping metrics.RunHealthCheck = func() error {
		return nil
	}

	opts := prometheus.GaugeOpts{
		Name: "ping",
		Help: "ping always succeeds",
	}

	pingCheck := metrics.NewHealthCheck(opts, 0, ping)

	// registering the same metrics again should fail
	if metrics.Registry.Register(prometheus.NewGauge(opts)) == nil {
		t.Error("Registration should have failed")
	}

	durationOpts := prometheus.GaugeOpts{
		Namespace:   opts.Namespace,
		Subsystem:   opts.Subsystem,
		Name:        fmt.Sprintf("%s_duration_seconds", opts.Name),
		Help:        "healthcheck run duration",
		ConstLabels: opts.ConstLabels,
	}
	if metrics.Registry.Register(prometheus.NewGauge(durationOpts)) == nil {
		t.Error("Registration should have failed")
	}

	t.Logf("result : %v", pingCheck.Run())
	gatheredMetrics, err := metrics.Registry.Gather()
	if err != nil {
		t.Errorf("Gathering metrics failed : %v", err)
	}

	if m := metrics.FindMetricFamilyByName(gatheredMetrics, opts.Name); m == nil {
		t.Errorf("Metric was not found : %v", opts.Name)
	} else {
		labels := prometheus.Labels{}
		for _, label := range m.Metric[0].Label {
			labels[*label.Name] = *label.Value
		}
		if labels[metrics.HEALTHCHECK_LABEL] != "status" {
			t.Errorf("label was not found (%v=status) in (%v) ", metrics.HEALTHCHECK_LABEL, labels)
		}
	}
	if m := metrics.FindMetricFamilyByName(gatheredMetrics, durationOpts.Name); m == nil {
		t.Errorf("Metric was not found : %v", durationOpts.Name)
		labels := prometheus.Labels{}
		for _, label := range m.Metric[0].Label {
			labels[*label.Name] = *label.Value
		}
		if labels[metrics.HEALTHCHECK_LABEL] != "duration" {
			t.Errorf("label was not found (%v=duration) in (%v) ", metrics.HEALTHCHECK_LABEL, labels)
		}
	}

}

func TestNewHealthCheck_WithNilRunFunc(t *testing.T) {
	defer metrics.ResetRegistry()
	var ping metrics.RunHealthCheck

	opts := prometheus.GaugeOpts{
		Name: "ping",
		Help: "ping always succeeds",
	}

	func() {
		defer func() {
			if p := recover(); p == nil {
				t.Errorf("creating new heatlhcheck should have failed because the run fuc was nil")
			}
		}()
		metrics.NewHealthCheck(opts, 0, ping)
	}()
}

func TestRegisterHealthCheck_WithLabels(t *testing.T) {
	defer metrics.ResetRegistry()
	var ping metrics.RunHealthCheck = func() error {
		return nil
	}

	opts := prometheus.GaugeOpts{
		Name:        "ping",
		Help:        "ping always succeeds",
		ConstLabels: map[string]string{"service": "Foo"},
	}

	pingCheck := metrics.NewHealthCheck(opts, 0, ping)
	t.Log(pingCheck)

	if pingCheck.Name() != prometheus.BuildFQName(opts.Namespace, opts.Subsystem, opts.Name) {
		t.Errorf("%v != %v", pingCheck.Name(), prometheus.BuildFQName(opts.Namespace, opts.Subsystem, opts.Name))
	}
	if len(pingCheck.Labels()) != 1 && pingCheck.Labels()["service"] != "Foo" {
		t.Errorf("ERROR: labels are not matching", pingCheck.Labels())
	}
	pingCheck.Run()
}

func TestRegisterHealthCheckVector(t *testing.T) {
	defer metrics.ResetRegistry()
	var ping metrics.RunHealthCheck = func() error {
		return nil
	}

	opts := &metrics.GaugeVecOpts{&prometheus.GaugeOpts{
		Name: "ping",
		Help: "ping always succeeds",
	}, []string{"service"},
	}

	pingCheck := metrics.NewHealthCheckVector(opts, 0, ping, []string{"Foo"})
	t.Log(pingCheck)

	if pingCheck.Name() != prometheus.BuildFQName(opts.Namespace, opts.Subsystem, opts.Name) {
		t.Errorf("%v != %v", pingCheck.Name(), prometheus.BuildFQName(opts.Namespace, opts.Subsystem, opts.Name))
	}
	if len(pingCheck.Labels()) != 1 && pingCheck.Labels()["service"] != "Foo" {
		t.Errorf("ERROR: labels are not matching", pingCheck.Labels())
	}
	pingCheck.Run()

	pingCheck2 := metrics.NewHealthCheckVector(opts, 0, ping, []string{"Bar"})
	t.Log(pingCheck2)
	pingCheck2.Run()

	func() {
		defer func() {
			if p := recover(); p == nil {
				t.Error("registering a dup healthcheck vector with the same label values should have failed")
			} else {
				t.Logf("%v", p)
			}
		}()
		metrics.NewHealthCheckVector(opts, 0, ping, []string{"Bar"})
	}()

	if count := len(metrics.HealthChecks()); count != 2 {
		t.Errorf("registered healthcheck count did not match : %d", count)
	}

}

func TestRegisterHealthCheck_WithRunInterval(t *testing.T) {
	defer metrics.ResetRegistry()
	var ping metrics.RunHealthCheck = func() error {
		return nil
	}

	opts := prometheus.GaugeOpts{
		Name:        "ping",
		Help:        "ping always succeeds",
		ConstLabels: map[string]string{"service": "Foo"},
	}

	now := time.Now()
	pingCheck := metrics.NewHealthCheck(opts, 5*time.Millisecond, ping)
	t.Log(pingCheck)

	time.Sleep(pingCheck.RunInterval() + 5*time.Millisecond)
	if pingCheck.LastResult() == nil {
		t.Errorf("ERROR: healthcheck should have run")
	} else if !pingCheck.LastResult().Time.After(now) {
		t.Errorf("the healthcheck should have run after : %v : %v", now, pingCheck.LastResult())
	}
	now = time.Now()
	time.Sleep(pingCheck.RunInterval() + 5*time.Millisecond)
	if pingCheck.LastResult() == nil {
		t.Errorf("ERROR: healthcheck should have run")
	} else if !pingCheck.LastResult().Time.After(now) {
		t.Errorf("the healthcheck should have run after : %v : %v", now, pingCheck.LastResult())
	}
	t.Log(pingCheck)

}

func TestHealthcheck_Run_Error(t *testing.T) {
	defer metrics.ResetRegistry()
	var ping metrics.RunHealthCheck = func() error {
		return errors.New("BOOM !!!")
	}

	opts := prometheus.GaugeOpts{
		Name:        "ping",
		Help:        "ping always fails",
		ConstLabels: map[string]string{"service": "Foo"},
	}

	pingCheck := metrics.NewHealthCheck(opts, 5*time.Millisecond, ping)
	t.Log(pingCheck)

	now := time.Now()
	result := pingCheck.Run()
	t.Logf("result : %v", result)
	if result.Success() {
		t.Errorf("should have failed")
	}
	if pingCheck.LastResult().Success() {
		t.Errorf("should have failed")
	}

	time.Sleep(pingCheck.RunInterval() + time.Millisecond)
	if !pingCheck.LastResult().Time.After(now) {
		t.Errorf("the healthcheck should have run after : %v : %v", now, pingCheck.LastResult())
	}
}

func TestHealthcheck_Run_Panic(t *testing.T) {
	defer metrics.ResetRegistry()
	var ping metrics.RunHealthCheck = func() error {
		panic("BOOM !!!")
	}

	opts := prometheus.GaugeOpts{
		Name:        "ping",
		Help:        "ping always fails",
		ConstLabels: map[string]string{"service": "Foo"},
	}

	pingCheck := metrics.NewHealthCheck(opts, 5*time.Millisecond, ping)
	t.Log(pingCheck)

	now := time.Now()
	result := pingCheck.Run()
	if result.Success() {
		t.Errorf("should have failed")
	}
	if pingCheck.LastResult().Success() {
		t.Errorf("should have failed")
	}

	time.Sleep(pingCheck.RunInterval() + time.Millisecond)
	if !pingCheck.LastResult().Time.After(now) {
		t.Errorf("the healthcheck should have run after : %v : %v", now, pingCheck.LastResult())
	}
}

func TestHealthcheck_StopTicker_StartTicker(t *testing.T) {
	var ping metrics.RunHealthCheck = func() error {
		panic("BOOM !!!")
	}

	opts := prometheus.GaugeOpts{
		Name:        "ping",
		Help:        "ping always fails",
		ConstLabels: map[string]string{"service": "Foo"},
	}

	pingCheck := metrics.NewHealthCheck(opts, 5*time.Millisecond, ping)
	pingCheck.StopTicker()
	if pingCheck.Scheduled() {
		t.Errorf("healthcheck should not be scheduled")
	}
	time.Sleep(pingCheck.RunInterval() * 2)
	if pingCheck.LastResult() != nil {
		t.Error("the healthcheck should not have been run")
	}

	pingCheck.StartTicker()
	if !pingCheck.Scheduled() {
		t.Errorf("healthcheck should be scheduled")
	}
	time.Sleep(pingCheck.RunInterval() + 5*time.Millisecond)
	if pingCheck.LastResult() == nil {
		t.Error("the healthcheck should have been run")
	}

}
