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

func TestSummarys_GetOrMustRegisterSummary(t *testing.T) {
	defer metrics.ResetRegistry()

	opts := &prometheus.SummaryOpts{
		Name: "summary1",
		Help: "Summary #1",
	}

	summary := metrics.GetOrMustRegisterSummary(opts)
	summary.Observe(1)
	metrics.GetOrMustRegisterSummary(opts).Observe(1)
	metrics.GetSummary(metrics.SummaryFQName(opts)).Observe(1)
	metrics.Summaries()[0].Observe(1)

	// check the same summary was incremented
	c := make(chan prometheus.Metric, 1)
	summary.Collect(c)
	collectedMetric := <-c
	metric := &io_prometheus_client.Metric{}
	collectedMetric.Write(metric)
	if *metric.Summary.SampleCount != 4 {
		t.Error("summary value should be 4")
	}

	if len(metrics.SummaryNames()) != 1 || metrics.SummaryNames()[0] != metrics.SummaryFQName(opts) {
		t.Errorf("Summary name %q was not returned : %v", metrics.SummaryFQName(opts), metrics.SummaryNames())
	}

	func() {
		defer func() {
			if p := recover(); p == nil {
				t.Errorf("The summary is already registered but with different opts should have triggered a panic")
			} else {
				t.Logf("panic : %v", p)
			}

		}()
		metrics.GetOrMustRegisterSummary(&prometheus.SummaryOpts{
			Name:        opts.Name,
			Help:        opts.Help,
			ConstLabels: map[string]string{"a": "b"},
		})
	}()

	func() {
		defer func() {
			if p := recover(); p == nil {
				t.Errorf("The summary is already registered but with different opts should have triggered a panic")
			} else {
				t.Logf("panic : %v", p)
			}

		}()
		metrics.GetOrMustRegisterSummaryVec(&metrics.SummaryVecOpts{
			SummaryOpts: &prometheus.SummaryOpts{
				Name:        opts.Name,
				Help:        opts.Help,
				ConstLabels: map[string]string{"a": "b"},
			},
		})
	}()

}

func TestSummarys_GetOrMustRegisterSummary_NameCollisionWithSummaryVec(t *testing.T) {
	defer metrics.ResetRegistry()

	opts := &prometheus.SummaryOpts{
		Name: "summary1",
		Help: "Summary #1",
	}
	metrics.GetOrMustRegisterSummaryVec(&metrics.SummaryVecOpts{SummaryOpts: opts})
	metrics.GetOrMustRegisterSummaryVec(&metrics.SummaryVecOpts{SummaryOpts: opts})

	func() {
		defer func() {
			if p := recover(); p == nil {
				t.Errorf("The summary is already registered but with different opts should have triggered a panic")
			} else {
				t.Logf("panic : %v", p)
			}

		}()
		metrics.GetOrMustRegisterSummary(opts)
	}()

}
