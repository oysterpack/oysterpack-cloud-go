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
	"errors"
	"sort"

	"github.com/oysterpack/oysterpack.go/pkg/app/config"
	"github.com/prometheus/client_golang/prometheus"
)

// NewHistogramMetricSpec is the HistogramMetricSpec factory method
func NewHistogramMetricSpec(spec config.HistogramMetricSpec) (HistogramMetricSpec, error) {
	metricSpec := func() (MetricSpec, error) {
		help, err := validateMetricSpec(spec)
		if err != nil {
			return MetricSpec{}, err
		}
		return MetricSpec{
			ServiceID: ServiceID(spec.ServiceId()),
			MetricID:  MetricID(spec.MetricId()),
			Help:      help,
		}, nil
	}

	if !spec.HasBuckets() {
		return HistogramMetricSpec{}, ConfigError(METRICS_SERVICE_ID, errors.New("HistogramMetricSpec : Buckets is required"), "")
	}
	bucketList, err := spec.Buckets()
	if err != nil {
		return HistogramMetricSpec{}, err
	}
	if bucketList.Len() == 0 {
		return HistogramMetricSpec{}, ConfigError(METRICS_SERVICE_ID, errors.New("HistogramMetricSpec : At least 1 bucket is required"), "")
	}

	buckets := make([]float64, bucketList.Len())
	for i := 0; i < len(buckets); i++ {
		buckets[i] = bucketList.At(i)
	}
	sort.Float64s(buckets)
	histogramMetricSpec, err := metricSpec()
	if err != nil {
		return HistogramMetricSpec{}, err
	}
	return HistogramMetricSpec{
		MetricSpec: histogramMetricSpec,
		Buckets:    buckets,
	}, nil
}

// HistogramMetricSpec is a MetricSpec for a Histogram.
type HistogramMetricSpec struct {
	MetricSpec
	Buckets []float64 // required
}

// HistogramOpts maps the spec to a prometheus HistogramOpts
func (a *HistogramMetricSpec) HistogramOpts() prometheus.HistogramOpts {
	return prometheus.HistogramOpts{
		Name:        a.MetricID.PrometheusName(a.ServiceID),
		Help:        a.Help,
		ConstLabels: MetricSpecLabels(a.ServiceID),
		Buckets:     a.Buckets,
	}
}

// NewHistogramVectorMetricSpec is the HistogramVectorMetricSpec factory method
func NewHistogramVectorMetricSpec(spec config.HistogramVectorMetricSpec) (*HistogramVectorMetricSpec, error) {
	histogramMetricSpec, err := spec.MetricSpec()
	if err != nil {
		return nil, err
	}
	metricSpec, err := NewHistogramMetricSpec(histogramMetricSpec)
	if err != nil {
		return nil, err
	}
	labelNamesList, err := spec.LabelNames()
	if err != nil {
		return nil, err
	}
	labels := make([]string, labelNamesList.Len())
	for i := 0; i < labelNamesList.Len(); i++ {
		labels[i], err = labelNamesList.At(i)
	}
	return &HistogramVectorMetricSpec{
		MetricVectorSpec: &MetricVectorSpec{metricSpec.MetricSpec, labels},
		Buckets:          metricSpec.Buckets,
	}, nil
}

// HistogramVectorMetricSpec is a MetricVectorSpec for a histogram vector
type HistogramVectorMetricSpec struct {
	*MetricVectorSpec
	Buckets []float64
}

// HistogramOpts maps the spec to a prometheus HistogramOpts
func (a *HistogramVectorMetricSpec) HistogramOpts() prometheus.HistogramOpts {
	return prometheus.HistogramOpts{
		Name:        a.MetricID.PrometheusName(a.ServiceID),
		Help:        a.Help,
		ConstLabels: MetricSpecLabels(a.ServiceID),
		Buckets:     a.Buckets,
	}
}

// HistogramMetric associates a Histogram with its spetric spec
type HistogramMetric struct {
	*HistogramMetricSpec
	prometheus.Histogram
}

// critical section that must be synchronized via metricsServiceMutex
func (a *HistogramMetric) register() {
	metrics := histograms[a.ServiceID]
	if metrics == nil {
		metrics = make(map[MetricID]*HistogramMetric)
		histograms[a.ServiceID] = metrics
	}
	if _, exists := metrics[a.MetricID]; !exists {
		metricsRegistry.MustRegister(a.Histogram)
		metrics[a.MetricID] = a
	}
}

// HistogramVectorMetric associates a HistogramVec with its spetric spec
type HistogramVectorMetric struct {
	*HistogramVectorMetricSpec
	*prometheus.HistogramVec
}

// critical section that must be synchronized via metricsServiceMutex
func (a *HistogramVectorMetric) register() {
	metrics := histogramVectors[a.ServiceID]
	if metrics == nil {
		metrics = make(map[MetricID]*HistogramVectorMetric)
		histogramVectors[a.ServiceID] = metrics
	}
	if _, exists := metrics[a.MetricID]; !exists {
		metricsRegistry.MustRegister(a.HistogramVec)
		metrics[a.MetricID] = a
	}
}
