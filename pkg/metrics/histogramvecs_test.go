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

func TestHistogramVecs_GetOrMustRegisterHistogramVec(t *testing.T) {
	defer metrics.ResetRegistry()

	opts := &metrics.HistogramVecOpts{
		HistogramOpts: &prometheus.HistogramOpts{
			Name:    "histogramVec1",
			Help:    "HistogramVec #1",
			Buckets: []float64{1, 5, 10},
		},
		Labels: []string{"request"},
	}

	histogramVec := metrics.GetOrMustRegisterHistogramVec(opts)
	histogramVec.WithLabelValues("process").Observe(1)
	metrics.GetOrMustRegisterHistogramVec(opts).WithLabelValues("process").Observe(2)
	metrics.GetHistogramVec(metrics.HistogramFQName(opts.HistogramOpts)).WithLabelValues("process").Observe(3)
	metrics.HistogramVecs()[0].WithLabelValues("process").Observe(4)

	// check the same histogramVec was incremented
	c := make(chan prometheus.Metric, 1)
	histogramVec.Collect(c)
	collectedMetric := <-c
	metric := &io_prometheus_client.Metric{}
	collectedMetric.Write(metric)
	if *metric.Histogram.SampleCount != 4 {
		t.Error("there should be 4 samples")
	}

	if len(metrics.HistogramVecNames()) != 1 || metrics.HistogramVecNames()[0] != metrics.HistogramFQName(opts.HistogramOpts) {
		t.Errorf("HistogramVec name %q was not returned : %v", metrics.HistogramFQName(opts.HistogramOpts), metrics.HistogramVecNames())
	}

	func() {
		defer func() {
			if p := recover(); p == nil {
				t.Errorf("The histogramVec is already registered but with different opts should have triggered a panic")
			} else {
				t.Logf("panic : %v", p)
			}

		}()
		metrics.GetOrMustRegisterHistogramVec(&metrics.HistogramVecOpts{
			HistogramOpts: &prometheus.HistogramOpts{
				Name: "histogramVec1",
				Help: "HistogramVec #1",
			},
			Labels: []string{"request", "errors"},
		})
	}()

}

func TestHistogramVecs_GetOrMustRegisterHistogramVec_NameCollisionWithHistogramVecVec(t *testing.T) {
	defer metrics.ResetRegistry()

	opts := &prometheus.HistogramOpts{
		Name: "histogramVec1",
		Help: "HistogramVec #1",
	}
	metrics.GetOrMustRegisterHistogram(opts)
	metrics.GetOrMustRegisterHistogram(opts)

	func() {
		defer func() {
			if p := recover(); p == nil {
				t.Errorf("The histogramVec is already registered but with different opts should have triggered a panic")
			} else {
				t.Logf("panic : %v", p)
			}

		}()
		metrics.GetOrMustRegisterHistogramVec(&metrics.HistogramVecOpts{
			HistogramOpts: &prometheus.HistogramOpts{
				Name: "histogramVec1",
				Help: "HistogramVec #1",
			},
			Labels: []string{"request"},
		})
	}()

}
