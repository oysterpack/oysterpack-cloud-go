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

func TestGaugeVecs_GetOrMustRegisterGaugeVec(t *testing.T) {
	defer metrics.ResetRegistry()

	opts := &metrics.GaugeVecOpts{
		GaugeOpts: &prometheus.GaugeOpts{
			Name: "gaugeVec1",
			Help: "GaugeVec #1",
		},
		Labels: []string{"request"},
	}

	gaugeVec := metrics.GetOrMustRegisterGaugeVec(opts)
	gaugeVec.WithLabelValues("process").Inc()
	metrics.GetOrMustRegisterGaugeVec(opts).WithLabelValues("process").Inc()
	metrics.GetGaugeVec(metrics.GaugeFQName(opts.GaugeOpts)).WithLabelValues("process").Inc()
	metrics.GaugeVecs()[0].WithLabelValues("process").Inc()

	// check the same gaugeVec was incremented
	c := make(chan prometheus.Metric, 1)
	gaugeVec.Collect(c)
	collectedMetric := <-c
	metric := &io_prometheus_client.Metric{}
	collectedMetric.Write(metric)
	if *metric.Gauge.Value != 4 {
		t.Error("gaugeVec value should be 4")
	}

	if len(metrics.GaugeVecNames()) != 1 || metrics.GaugeVecNames()[0] != metrics.GaugeFQName(opts.GaugeOpts) {
		t.Errorf("GaugeVec name %q was not returned : %v", metrics.GaugeFQName(opts.GaugeOpts), metrics.GaugeVecNames())
	}

	func() {
		defer func() {
			if p := recover(); p == nil {
				t.Errorf("The gaugeVec is already registered but with different opts should have triggered a panic")
			} else {
				t.Logf("panic : %v", p)
			}

		}()
		metrics.GetOrMustRegisterGaugeVec(&metrics.GaugeVecOpts{
			GaugeOpts: &prometheus.GaugeOpts{
				Name: "gaugeVec1",
				Help: "GaugeVec #1",
			},
			Labels: []string{"request", "errors"},
		})
	}()

}

func TestGaugeVecs_GetOrMustRegisterGaugeVec_NameCollisionWithGaugeVecVec(t *testing.T) {
	defer metrics.ResetRegistry()

	opts := &prometheus.GaugeOpts{
		Name: "gaugeVec1",
		Help: "GaugeVec #1",
	}
	metrics.GetOrMustRegisterGauge(opts)
	metrics.GetOrMustRegisterGauge(opts)

	func() {
		defer func() {
			if p := recover(); p == nil {
				t.Errorf("The gaugeVec is already registered but with different opts should have triggered a panic")
			} else {
				t.Logf("panic : %v", p)
			}

		}()
		metrics.GetOrMustRegisterGaugeVec(&metrics.GaugeVecOpts{
			GaugeOpts: &prometheus.GaugeOpts{
				Name: "gaugeVec1",
				Help: "GaugeVec #1",
			},
			Labels: []string{"request"},
		})
	}()

}
