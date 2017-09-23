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

	"github.com/oysterpack/oysterpack.go/pkg/metrics"
	"github.com/prometheus/client_golang/prometheus"
)

func TestRegisteringMetricsOfDifferentTypesWithSameFullyQualifiedNamesShouldFail(t *testing.T) {
	var metric1Opts = prometheus.Opts{
		Name: "metric1",
		Help: "Metric 1 Help",
	}
	counter := prometheus.NewCounter(prometheus.CounterOpts(metric1Opts))

	registry := prometheus.NewPedanticRegistry()
	registry.MustRegister(counter)
	counter.Inc()

	gatheredMetrics, _ := registry.Gather()
	for _, m := range gatheredMetrics {
		t.Logf("%v", m)
	}

	// different metric types with the same name are not allowed
	counterVec1Opts := metrics.NewCounterVecOpts(prometheus.CounterOpts(metric1Opts), "a", "b")
	if err := registry.Register(prometheus.NewCounterVec(counterVec1Opts.CounterOpts, counterVec1Opts.Labels)); err == nil {
		t.Error("should have failed because the counter metric has the same fully-qualified name")
	} else {
		t.Logf("Failed to register counter vector : %v", err)
	}
	if err := registry.Register(prometheus.NewGauge(prometheus.GaugeOpts(metric1Opts))); err == nil {
		t.Error("should have failed because the previous metric has the same fully-qualified name")
	} else {
		t.Logf("Failed to register gauge : %v", err)
	}
	gaugeVec1Opts := metrics.NewGaugeVecOpts(prometheus.GaugeOpts(metric1Opts), "a", "b", "c")
	if err := registry.Register(prometheus.NewGaugeVec(gaugeVec1Opts.GaugeOpts, gaugeVec1Opts.Labels)); err == nil {
		t.Error("should have failed because the previous metric has the same fully-qualified name")
	} else {
		t.Logf("Failed to register gauge vector : %v", err)
	}

	if err := registry.Register(prometheus.NewCounter(prometheus.CounterOpts{
		Name:        metric1Opts.Name,
		Help:        metric1Opts.Help,
		ConstLabels: prometheus.Labels{"a": "1"},
	})); err == nil {
		t.Error("should have failed because the previous metric has the same fully-qualified name")
	} else {
		t.Logf("Failed to register gauge : %v", err)
	}

	if err := registry.Register(prometheus.NewGauge(prometheus.GaugeOpts{
		Name:        metric1Opts.Name,
		Help:        metric1Opts.Help,
		ConstLabels: prometheus.Labels{"a": "1"},
	})); err == nil {
		t.Error("should have failed because the previous metric has the same fully-qualified name")
	} else {
		t.Logf("Failed to register gauge : %v", err)
	}
}
