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

	"github.com/oysterpack/oysterpack.go/oysterpack/metrics"
	"github.com/prometheus/client_golang/prometheus"
)

func TestRegisterHealthCheck_WithNoLabels(t *testing.T) {
	var ping metrics.RunHealthCheck = func() error {
		return nil
	}

	opts := prometheus.GaugeOpts{
		Name: "ping",
		Help: "ping always succeeds",
	}

	registry := prometheus.NewPedanticRegistry()
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

	if err := pingCheck.Register(registry); err != nil {
		t.Errorf("ERROR: %v", err)
	}

	if err := pingCheck.Register(registry); err != nil {
		t.Log(err)
	} else if pingCheck != nil {
		t.Error("ERROR: healthcheck should be nil")
	}
}

func TestNewHealthCheck_WithNilRunFunc(t *testing.T) {
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
	var ping metrics.RunHealthCheck = func() error {
		return nil
	}

	opts := prometheus.GaugeOpts{
		Name:        "ping",
		Help:        "ping always succeeds",
		ConstLabels: map[string]string{"service": "Foo"},
	}

	registry := prometheus.NewPedanticRegistry()
	pingCheck := metrics.NewHealthCheck(opts, 0, ping)
	t.Log(pingCheck)

	t.Log(pingCheck)
	if pingCheck.Name() != prometheus.BuildFQName(opts.Namespace, opts.Subsystem, opts.Name) {
		t.Errorf("%v != %v", pingCheck.Name(), prometheus.BuildFQName(opts.Namespace, opts.Subsystem, opts.Name))
	}
	if len(pingCheck.Labels()) != 1 && pingCheck.Labels()["service"] != "Foo" {
		t.Errorf("ERROR: labels are not matching", pingCheck.Labels())
	}
	pingCheck.Run()

	if err := pingCheck.Register(registry); err != nil {
		t.Errorf("ERROR: %v", err)
	}

	if err := pingCheck.Register(registry); err != nil {
		t.Log(err)
	} else if pingCheck != nil {
		t.Error("ERROR: healthcheck should be nil")
	}
}

func TestRegisterHealthCheck_WithRunInterval(t *testing.T) {
	var ping metrics.RunHealthCheck = func() error {
		return nil
	}

	opts := prometheus.GaugeOpts{
		Name:        "ping",
		Help:        "ping always succeeds",
		ConstLabels: map[string]string{"service": "Foo"},
	}

	registry := prometheus.NewPedanticRegistry()
	now := time.Now()
	pingCheck := metrics.NewHealthCheck(opts, 5*time.Millisecond, ping)
	t.Log(pingCheck)
	pingCheck.MustRegister(registry)

	time.Sleep(pingCheck.RunInterval() + time.Millisecond)
	if pingCheck.LastResult() == nil {
		t.Errorf("ERROR: healthcheck should have run")
	} else if !pingCheck.LastResult().Time.After(now) {
		t.Errorf("the healthcheck should have run after : %v : %v", now, pingCheck.LastResult())
	}
	now = time.Now()
	time.Sleep(pingCheck.RunInterval() + time.Millisecond)
	if pingCheck.LastResult() == nil {
		t.Errorf("ERROR: healthcheck should have run")
	} else if !pingCheck.LastResult().Time.After(now) {
		t.Errorf("the healthcheck should have run after : %v : %v", now, pingCheck.LastResult())
	}
	t.Log(pingCheck)

}

func TestHealthcheck_Run_Error(t *testing.T) {
	var ping metrics.RunHealthCheck = func() error {
		return errors.New("BOOM !!!")
	}

	opts := prometheus.GaugeOpts{
		Name:        "ping",
		Help:        "ping always fails",
		ConstLabels: map[string]string{"service": "Foo"},
	}

	registry := prometheus.NewPedanticRegistry()
	pingCheck := metrics.NewHealthCheck(opts, 5*time.Millisecond, ping)
	t.Log(pingCheck)
	pingCheck.MustRegister(registry)
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

func TestHealthcheck_Run_Panic(t *testing.T) {
	var ping metrics.RunHealthCheck = func() error {
		panic("BOOM !!!")
	}

	opts := prometheus.GaugeOpts{
		Name:        "ping",
		Help:        "ping always fails",
		ConstLabels: map[string]string{"service": "Foo"},
	}

	registry := prometheus.NewPedanticRegistry()
	pingCheck := metrics.NewHealthCheck(opts, 5*time.Millisecond, ping)
	t.Log(pingCheck)
	pingCheck.MustRegister(registry)
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
	time.Sleep(pingCheck.RunInterval() + time.Millisecond)
	if pingCheck.LastResult() == nil {
		t.Error("the healthcheck should have been run")
	}

}
