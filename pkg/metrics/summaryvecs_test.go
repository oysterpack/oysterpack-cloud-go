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

func TestSummaryVecs_GetOrMustRegisterSummaryVec(t *testing.T) {
	defer metrics.ResetRegistry()

	opts := &metrics.SummaryVecOpts{
		SummaryOpts: &prometheus.SummaryOpts{
			Name: "summaryVec1",
			Help: "SummaryVec #1",
		},
		Labels: []string{"request"},
	}

	summaryVec := metrics.GetOrMustRegisterSummaryVec(opts)
	summaryVec.WithLabelValues("process").Observe(1)
	metrics.GetOrMustRegisterSummaryVec(opts).WithLabelValues("process").Observe(1)
	metrics.GetSummaryVec(metrics.SummaryFQName(opts.SummaryOpts)).WithLabelValues("process").Observe(1)
	metrics.SummaryVecs()[0].WithLabelValues("process").Observe(1)

	// check the same summaryVec was incremented
	c := make(chan prometheus.Metric, 1)
	summaryVec.Collect(c)
	collectedMetric := <-c
	metric := &io_prometheus_client.Metric{}
	collectedMetric.Write(metric)
	if *metric.Summary.SampleCount != 4 {
		t.Error("summaryVec value should be 4")
	}

	if len(metrics.SummaryVecNames()) != 1 || metrics.SummaryVecNames()[0] != metrics.SummaryFQName(opts.SummaryOpts) {
		t.Errorf("SummaryVec name %q was not returned : %v", metrics.SummaryFQName(opts.SummaryOpts), metrics.SummaryVecNames())
	}

	func() {
		defer func() {
			if p := recover(); p == nil {
				t.Errorf("The summaryVec is already registered but with different opts should have triggered a panic")
			} else {
				t.Logf("panic : %v", p)
			}

		}()
		metrics.GetOrMustRegisterSummaryVec(&metrics.SummaryVecOpts{
			SummaryOpts: &prometheus.SummaryOpts{
				Name: "summaryVec1",
				Help: "SummaryVec #1",
			},
			Labels: []string{"request", "errors"},
		})
	}()

}

func TestSummaryVecs_GetOrMustRegisterSummaryVec_NameCollisionWithSummaryVecVec(t *testing.T) {
	defer metrics.ResetRegistry()

	opts := &prometheus.SummaryOpts{
		Name: "summaryVec1",
		Help: "SummaryVec #1",
	}
	metrics.GetOrMustRegisterSummary(opts)
	metrics.GetOrMustRegisterSummary(opts)

	func() {
		defer func() {
			if p := recover(); p == nil {
				t.Errorf("The summaryVec is already registered but with different opts should have triggered a panic")
			} else {
				t.Logf("panic : %v", p)
			}

		}()
		metrics.GetOrMustRegisterSummaryVec(&metrics.SummaryVecOpts{
			SummaryOpts: &prometheus.SummaryOpts{
				Name: "summaryVec1",
				Help: "SummaryVec #1",
			},
			Labels: []string{"request"},
		})
	}()

}
