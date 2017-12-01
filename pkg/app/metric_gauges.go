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

func NewGaugeMetricSpec(spec config.GaugeMetricSpec) (GaugeMetricSpec, error) {
	help, err := validateMetricSpec(spec)
	if err != nil {
		return GaugeMetricSpec{}, err
	}

	return GaugeMetricSpec{
		ServiceID: ServiceID(spec.ServiceId()),
		MetricID:  MetricID(spec.MetricId()),
		Help:      help,
	}, nil
}

// GaugeMetricSpec is a type alias for a Gauge MetricSpec
type GaugeMetricSpec MetricSpec

// GaugeOpts maps the spec to a prometheus GaugeOpts
func (a *GaugeMetricSpec) GaugeOpts() prometheus.GaugeOpts {
	return prometheus.GaugeOpts{
		Name:        a.MetricID.PrometheusName(a.ServiceID),
		Help:        a.Help,
		ConstLabels: MetricSpecLabels(a.ServiceID),
	}
}

// NewGaugeVectorMetricSpec is the GaugeVectorMetricSpec factory method
//
// errors:
// - ConfigError
func NewGaugeVectorMetricSpec(spec config.GaugeVectorMetricSpec) (*GaugeVectorMetricSpec, error) {
	labels, err := validateMetricVectorSpec(spec)
	if err != nil {
		return nil, err
	}
	gaugeSpec, err := spec.MetricSpec()
	if err != nil {
		return nil, err
	}
	gaugeMetricSpec, err := NewGaugeMetricSpec(gaugeSpec)
	if err != nil {
		return nil, err
	}
	return &GaugeVectorMetricSpec{MetricSpec: MetricSpec(gaugeMetricSpec), DynamicLabels: labels}, nil
}

// GaugeVectorMetricSpec is a type alaias for the Gauge MetricVectorSpec
type GaugeVectorMetricSpec MetricVectorSpec

// GaugeOpts maps the spec to a prometheus GaugeOpts
func (a *GaugeVectorMetricSpec) GaugeOpts() prometheus.GaugeOpts {
	return prometheus.GaugeOpts{
		Name:        a.MetricID.PrometheusName(a.ServiceID),
		Help:        a.Help,
		ConstLabels: MetricSpecLabels(a.ServiceID),
	}
}

// GaugeMetric associates the Gauge with its spec.
// It is cached after it is registered with prometheus.
type GaugeMetric struct {
	*GaugeMetricSpec
	prometheus.Gauge
}

// critical section that must be synchronized via metricsServiceMutex
func (a *GaugeMetric) register() {
	metrics := gauges[a.ServiceID]
	if metrics == nil {
		metrics = make(map[MetricID]*GaugeMetric)
		gauges[a.ServiceID] = metrics
	}
	if _, exists := metrics[a.MetricID]; !exists {
		metricsRegistry.MustRegister(a.Gauge)
		metrics[a.MetricID] = a
	}
}

// GaugeVectorMetric associates the Gauge with its spec.
// It is cached after it is registered with prometheus.
type GaugeVectorMetric struct {
	*GaugeVectorMetricSpec
	*prometheus.GaugeVec
}

// critical section that must be synchronized via metricsServiceMutex
func (a *GaugeVectorMetric) register() {
	metrics := gaugeVectors[a.ServiceID]
	if metrics == nil {
		metrics = make(map[MetricID]*GaugeVectorMetric)
		gaugeVectors[a.ServiceID] = metrics
	}
	if _, exists := metrics[a.MetricID]; !exists {
		metricsRegistry.MustRegister(a.GaugeVec)
		metrics[a.MetricID] = a
	}
}
