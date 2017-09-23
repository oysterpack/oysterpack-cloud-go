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

func TestCounters_GetOrMustRegisterCounter(t *testing.T) {
	defer metrics.ResetRegistry()

	opts := &prometheus.CounterOpts{
		Name: "counter1",
		Help: "Counter #1",
	}

	counter := metrics.GetOrMustRegisterCounter(opts)
	counter.Inc()
	metrics.GetOrMustRegisterCounter(opts).Inc()
	metrics.GetCounter(metrics.CounterFQName(opts)).Inc()
	metrics.Counters()[0].Inc()

	// check the same counter was incremented
	c := make(chan prometheus.Metric, 1)
	counter.Collect(c)
	collectedMetric := <-c
	metric := &io_prometheus_client.Metric{}
	collectedMetric.Write(metric)
	if *metric.Counter.Value != 4 {
		t.Error("counter value should be 4")
	}

	if len(metrics.CounterNames()) != 1 || metrics.CounterNames()[0] != metrics.CounterFQName(opts) {
		t.Errorf("Counter name %q was not returned : %v", metrics.CounterFQName(opts), metrics.CounterNames())
	}

	func() {
		defer func() {
			if p := recover(); p == nil {
				t.Errorf("The counter is already registered but with different opts should have triggered a panic")
			} else {
				t.Logf("panic : %v", p)
			}

		}()
		metrics.GetOrMustRegisterCounter(&prometheus.CounterOpts{
			Name:        opts.Name,
			Help:        opts.Help,
			ConstLabels: map[string]string{"a": "b"},
		})
	}()

}
