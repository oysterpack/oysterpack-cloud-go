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

func TestHistograms_GetOrMustRegisterHistogram(t *testing.T) {
	metrics.ResetRegistry()
	defer metrics.ResetRegistry()

	opts := &prometheus.HistogramOpts{
		Name:    "histogram1",
		Help:    "Histogram #1",
		Buckets: []float64{1, 5, 10},
	}

	histogram := metrics.GetOrMustRegisterHistogram(opts)
	histogram.Observe(1.4)
	metrics.GetOrMustRegisterHistogram(opts).Observe(2.4)
	metrics.GetHistogram(metrics.HistogramFQName(opts)).Observe(3.4)
	metrics.Histograms()[0].Observe(4.4)

	// check the same histogram was incremented
	c := make(chan prometheus.Metric, 1)
	histogram.Collect(c)
	collectedMetric := <-c
	metric := &io_prometheus_client.Metric{}
	collectedMetric.Write(metric)
	if *metric.Histogram.SampleCount != 4 {
		t.Error("4 samples should have been observed")
	}

	if len(metrics.HistogramNames()) != 1 || metrics.HistogramNames()[0] != metrics.HistogramFQName(opts) {
		t.Errorf("Histogram name %q was not returned : %v", metrics.HistogramFQName(opts), metrics.HistogramNames())
	}

	func() {
		defer func() {
			if p := recover(); p == nil {
				t.Errorf("The histogram is already registered but with different opts should have triggered a panic")
			} else {
				t.Logf("panic : %v", p)
			}

		}()
		metrics.GetOrMustRegisterHistogram(&prometheus.HistogramOpts{
			Name:        opts.Name,
			Help:        opts.Help,
			ConstLabels: map[string]string{"a": "b"},
		})
	}()

	func() {
		defer func() {
			if p := recover(); p == nil {
				t.Errorf("The histogram is already registered but with different opts should have triggered a panic")
			} else {
				t.Logf("panic : %v", p)
			}

		}()
		metrics.GetOrMustRegisterHistogramVec(&metrics.HistogramVecOpts{
			HistogramOpts: &prometheus.HistogramOpts{
				Name:        opts.Name,
				Help:        opts.Help,
				ConstLabels: map[string]string{"a": "b"},
			},
		})
	}()

}

func TestHistograms_GetOrMustRegisterHistogram_NameCollisionWithHistogramVec(t *testing.T) {
	defer metrics.ResetRegistry()

	opts := &prometheus.HistogramOpts{
		Name: "histogram1",
		Help: "Histogram #1",
	}
	metrics.GetOrMustRegisterHistogramVec(&metrics.HistogramVecOpts{HistogramOpts: opts})
	metrics.GetOrMustRegisterHistogramVec(&metrics.HistogramVecOpts{HistogramOpts: opts})

	func() {
		defer func() {
			if p := recover(); p == nil {
				t.Errorf("The histogram is already registered but with different opts should have triggered a panic")
			} else {
				t.Logf("panic : %v", p)
			}

		}()
		metrics.GetOrMustRegisterHistogram(opts)
	}()

}
