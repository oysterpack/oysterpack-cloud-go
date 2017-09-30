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

	"github.com/prometheus/client_model/go"
)

func TestGauges_GetOrMustRegisterGauge(t *testing.T) {
	defer metrics.ResetRegistry()

	opts := &prometheus.GaugeOpts{
		Name: "gauge1",
		Help: "Gauge #1",
	}

	gauge := metrics.GetOrMustRegisterGauge(opts)
	gauge.Inc()
	metrics.GetOrMustRegisterGauge(opts).Inc()
	metrics.GetGauge(metrics.GaugeFQName(opts)).Inc()
	metrics.Gauges()[0].Inc()

	// check the same gauge was incremented
	c := make(chan prometheus.Metric, 1)
	gauge.Collect(c)
	collectedMetric := <-c
	metric := &io_prometheus_client.Metric{}
	collectedMetric.Write(metric)
	if *metric.Gauge.Value != 4 {
		t.Error("gauge value should be 4")
	}

	if len(metrics.GaugeNames()) != 1 || metrics.GaugeNames()[0] != metrics.GaugeFQName(opts) {
		t.Errorf("Gauge name %q was not returned : %v", metrics.GaugeFQName(opts), metrics.GaugeNames())
	}

	func() {
		defer func() {
			if p := recover(); p == nil {
				t.Errorf("The gauge is already registered but with different opts should have triggered a panic")
			} else {
				t.Logf("panic : %v", p)
			}

		}()
		metrics.GetOrMustRegisterGauge(&prometheus.GaugeOpts{
			Name:        opts.Name,
			Help:        opts.Help,
			ConstLabels: map[string]string{"a": "b"},
		})
	}()

	func() {
		defer func() {
			if p := recover(); p == nil {
				t.Errorf("The gauge is already registered but with different opts should have triggered a panic")
			} else {
				t.Logf("panic : %v", p)
			}

		}()
		metrics.GetOrMustRegisterGaugeVec(&metrics.GaugeVecOpts{
			GaugeOpts: &prometheus.GaugeOpts{
				Name:        opts.Name,
				Help:        opts.Help,
				ConstLabels: map[string]string{"a": "b"},
			},
		})
	}()

}

func TestGauges_GetOrMustRegisterGauge_NameCollisionWithGaugeVec(t *testing.T) {
	defer metrics.ResetRegistry()

	opts := &prometheus.GaugeOpts{
		Name: "gauge1",
		Help: "Gauge #1",
	}
	metrics.GetOrMustRegisterGaugeVec(&metrics.GaugeVecOpts{GaugeOpts: opts})
	metrics.GetOrMustRegisterGaugeVec(&metrics.GaugeVecOpts{GaugeOpts: opts})

	func() {
		defer func() {
			if p := recover(); p == nil {
				t.Errorf("The gauge is already registered but with different opts should have triggered a panic")
			} else {
				t.Logf("panic : %v", p)
			}

		}()
		metrics.GetOrMustRegisterGauge(opts)
	}()

}

func TestGauges_GetOrMustRegisterGaugeFunc(t *testing.T) {
	defer metrics.ResetRegistry()

	opts := &prometheus.GaugeOpts{
		Name: "gauge1",
		Help: "Gauge #1",
	}

	metrics.GetOrMustRegisterGaugeFunc(opts, func() float64 {
		return 1
	})
	metrics.GetOrMustRegisterGaugeFunc(opts, func() float64 {
		return 1
	})
}

func TestGauges_GetOrMustRegisterGaugeFunc_withNameCollision(t *testing.T) {
	defer metrics.ResetRegistry()

	opts := &prometheus.GaugeOpts{
		Name: "gauge1",
		Help: "Gauge #1",
	}

	metrics.GetOrMustRegisterGauge(opts)
	metrics.GetOrMustRegisterGauge(opts)

	func() {
		defer func() {
			if p := recover(); p == nil {
				t.Errorf("name collision should ahve triggered panic")
			} else {
				t.Logf("panic : %v", p)
			}

		}()
		metrics.GetOrMustRegisterGaugeFunc(opts, func() float64 {
			return 1
		})
	}()
}

func TestGauges_GetOrMustRegisterGaugeFunc_WithDifferentOpts(t *testing.T) {
	defer metrics.ResetRegistry()

	opts := &prometheus.GaugeOpts{
		Name: "gauge1",
		Help: "Gauge #1",
	}

	metrics.GetOrMustRegisterGaugeFunc(opts, func() float64 {
		return 1
	})

	func() {
		defer func() {
			if p := recover(); p == nil {
				t.Errorf("panic should have been triggered because the opts don't match")
			} else {
				t.Logf("panic : %v", p)
			}

		}()
		opts := &prometheus.GaugeOpts{
			Name: "gauge1",
			Help: "sdfsdfsdf",
		}
		metrics.GetOrMustRegisterGaugeFunc(opts, func() float64 {
			return 1
		})
	}()

}
