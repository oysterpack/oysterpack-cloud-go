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

package app

import (
	"github.com/oysterpack/oysterpack.go/pkg/app/config"
	"github.com/prometheus/client_golang/prometheus"
)

// NewCounterMetricSpec is a CounterMetricSpec factory method
//
// errors
// - ConfigError
func NewCounterMetricSpec(spec config.CounterMetricSpec) (CounterMetricSpec, error) {
	help, err := validateMetricSpec(spec)
	if err != nil {
		return CounterMetricSpec{}, err
	}

	return CounterMetricSpec{
		ServiceID: ServiceID(spec.ServiceId()),
		MetricID:  MetricID(spec.MetricId()),
		Help:      help,
	}, nil
}

// CounterMetricSpec type alias for a Counter MetricSpec
type CounterMetricSpec MetricSpec

// CounterOpts maps the spec to a prometheus CounterOpts
func (a *CounterMetricSpec) CounterOpts() prometheus.CounterOpts {
	return prometheus.CounterOpts{
		Name:        a.MetricID.PrometheusName(a.ServiceID),
		Help:        a.Help,
		ConstLabels: MetricSpecLabels(a.ServiceID),
	}
}

// NewCounterVectorMetricSpec is the CounterVectorMetricSpec factory method
func NewCounterVectorMetricSpec(spec config.CounterVectorMetricSpec) (*CounterVectorMetricSpec, error) {
	labels, err := validateMetricVectorSpec(spec)
	if err != nil {
		return nil, err
	}
	counterSpec, err := spec.MetricSpec()
	if err != nil {
		return nil, err
	}
	counterMetricSpec, err := NewCounterMetricSpec(counterSpec)
	if err != nil {
		return nil, err
	}
	return &CounterVectorMetricSpec{MetricSpec: MetricSpec(counterMetricSpec), DynamicLabels: labels}, nil
}

// CounterVectorMetricSpec is a type alias for a Counter MetricVectorSpec
type CounterVectorMetricSpec MetricVectorSpec

func (a *CounterVectorMetricSpec) CounterOpts() prometheus.CounterOpts {
	return prometheus.CounterOpts{
		Name:        a.MetricID.PrometheusName(a.ServiceID),
		Help:        a.Help,
		ConstLabels: MetricSpecLabels(a.ServiceID),
	}
}

// CounterMetric associates Counter with its spec.
// This is cached after the metric is registered.
type CounterMetric struct {
	*CounterMetricSpec
	prometheus.Counter
}

// critical section that must be synchronized via metricsServiceMutex
func (a *CounterMetric) register() {
	metrics := counters[a.ServiceID]
	if metrics == nil {
		metrics = make(map[MetricID]*CounterMetric)
		counters[a.ServiceID] = metrics
	}
	if _, exists := metrics[a.MetricID]; !exists {
		metricsRegistry.MustRegister(a.Counter)
		metrics[a.MetricID] = a
	}
}

// CounterVectorMetric associates the CounterVec with its spec.
// This is cached after the metric is registered.
type CounterVectorMetric struct {
	*CounterVectorMetricSpec
	*prometheus.CounterVec
}

// critical section that must be synchronized via metricsServiceMutex
func (a *CounterVectorMetric) register() {
	metrics := counterVectors[a.ServiceID]
	if metrics == nil {
		metrics = make(map[MetricID]*CounterVectorMetric)
		counterVectors[a.ServiceID] = metrics
	}
	if _, exists := metrics[a.MetricID]; !exists {
		metricsRegistry.MustRegister(a.CounterVec)
		metrics[a.MetricID] = a
	}
}
