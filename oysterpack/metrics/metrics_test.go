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

package metrics

import (
	"testing"

	"github.com/prometheus/client_golang/prometheus"
)

func TestRegisteringMetricsOfDifferentTypesWithSAmeFullyQualifiedNamesShouldFail(t *testing.T) {
	var metric1Opts = prometheus.Opts{
		Name: "metric1",
		Help: "Metric 1",
	}
	var counter1 = prometheus.NewCounter(prometheus.CounterOpts(metric1Opts))
	var counterVec1Opts = prometheus.NewCounterVec(prometheus.CounterOpts(metric1Opts), []string{"a", "b"})

	registry := prometheus.NewPedanticRegistry()
	registry.MustRegister(counter1)

	if err := registry.Register(counterVec1Opts); err == nil {
		t.Error("should have failed because the previous metric has the same fully-qualified name")
	} else {
		t.Log(err)
	}
}
