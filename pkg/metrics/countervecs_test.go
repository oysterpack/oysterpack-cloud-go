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

func TestCounterVecs_GetOrMustRegisterCounterVec(t *testing.T) {
	defer metrics.ResetRegistry()

	opts := &metrics.CounterVecOpts{
		CounterOpts: &prometheus.CounterOpts{
			Name: "counterVec1",
			Help: "CounterVec #1",
		},
		Labels: []string{"request"},
	}

	counterVec := metrics.GetOrMustRegisterCounterVec(opts)
	counterVec.WithLabelValues("process").Inc()
	metrics.GetOrMustRegisterCounterVec(opts).WithLabelValues("process").Inc()
	metrics.GetCounterVec(metrics.CounterFQName(opts.CounterOpts)).WithLabelValues("process").Inc()
	metrics.CounterVecs()[0].WithLabelValues("process").Inc()

	// check the same counterVec was incremented
	c := make(chan prometheus.Metric, 1)
	counterVec.Collect(c)
	collectedMetric := <-c
	metric := &io_prometheus_client.Metric{}
	collectedMetric.Write(metric)
	if *metric.Counter.Value != 4 {
		t.Error("counterVec value should be 4")
	}

	if len(metrics.CounterVecNames()) != 1 || metrics.CounterVecNames()[0] != metrics.CounterFQName(opts.CounterOpts) {
		t.Errorf("CounterVec name %q was not returned : %v", metrics.CounterFQName(opts.CounterOpts), metrics.CounterVecNames())
	}

	func() {
		defer func() {
			if p := recover(); p == nil {
				t.Errorf("The counterVec is already registered but with different opts should have triggered a panic")
			} else {
				t.Logf("panic : %v", p)
			}

		}()
		metrics.GetOrMustRegisterCounterVec(&metrics.CounterVecOpts{
			CounterOpts: &prometheus.CounterOpts{
				Name: "counterVec1",
				Help: "CounterVec #1",
			},
			Labels: []string{"request", "errors"},
		})
	}()

}

func TestCounterVecs_GetOrMustRegisterCounterVec_NameCollisionWithCounterVecVec(t *testing.T) {
	defer metrics.ResetRegistry()

	opts := &prometheus.CounterOpts{
		Name: "counterVec1",
		Help: "CounterVec #1",
	}
	metrics.GetOrMustRegisterCounter(opts)
	metrics.GetOrMustRegisterCounter(opts)

	func() {
		defer func() {
			if p := recover(); p == nil {
				t.Errorf("The counterVec is already registered but with different opts should have triggered a panic")
			} else {
				t.Logf("panic : %v", p)
			}

		}()
		metrics.GetOrMustRegisterCounterVec(&metrics.CounterVecOpts{
			CounterOpts: &prometheus.CounterOpts{
				Name: "counterVec1",
				Help: "CounterVec #1",
			},
			Labels: []string{"request"},
		})
	}()

}
