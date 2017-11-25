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
	"os"

	"sync"

	"github.com/oysterpack/oysterpack.go/pkg/app/config"
	"github.com/prometheus/client_golang/prometheus"
)

const (
	METRICS_SERVICE_ID               = ServiceID(0xe3054017c1b1d214)
	DEFAULT_METRICS_HTTP_PORT uint16 = 4444
)

// metrics
var (
	metricsServiceMutex        sync.Mutex
	metricsServiceBootstrapped = false

	// global metrics registry
	metricsRegistry = NewMetricsRegistry(true)

	counters       = make(map[ServiceID]map[MetricID]*CounterMetric)
	counterVectors = make(map[ServiceID]map[MetricID]*CounterVectorMetric)

	gauges       = make(map[ServiceID]map[MetricID]*GaugeMetric)
	gaugeVectors = make(map[ServiceID]map[MetricID]*GaugeVectorMetric)

	histograms       = make(map[ServiceID]map[MetricID]*HistogramMetric)
	histogramVectors = make(map[ServiceID]map[MetricID]*HistogramVectorMetric)
)

// NewRegistry creates a new registry.
// If collectProcessMetrics = true, then the prometheus GoCollector and ProcessCollectors are registered.
func NewMetricsRegistry(collectProcessMetrics bool) *prometheus.Registry {
	registry := prometheus.NewRegistry()
	if collectProcessMetrics {
		registry.MustRegister(
			prometheus.NewGoCollector(),
			prometheus.NewProcessCollector(os.Getpid(), ""),
		)
	}
	return registry
}

type MetricSpec struct {
	ServiceID
	MetricID
	Help string
}

type MetricVectorSpec struct {
	MetricSpec
	DynamicLabels []string
}

func NewCounterMetricSpec(spec config.CounterMetricSpec) (CounterMetricSpec, error) {
	help, err := spec.Help()
	if err != nil {
		return CounterMetricSpec{}, err
	}
	return CounterMetricSpec{
		ServiceID: ServiceID(spec.ServiceId()),
		MetricID:  MetricID(spec.MetricId()),
		Help:      help,
	}, nil
}

type CounterMetricSpec MetricSpec

func (a CounterMetricSpec) CounterOpts() prometheus.CounterOpts {
	return prometheus.CounterOpts{
		Name:        a.MetricID.Hex(),
		Help:        a.Help,
		ConstLabels: a.ServiceID.MetricSpecLabels(),
	}
}

func NewCounterVectorMetricSpec(spec config.CounterVectorMetricSpec) (*CounterVectorMetricSpec, error) {
	counterSpec, err := spec.MetricSpec()
	if err != nil {
		return nil, err
	}
	counterMetricSpec, err := NewCounterMetricSpec(counterSpec)
	if err != nil {
		return nil, err
	}
	labelNamesList, err := spec.LabelNames()
	if err != nil {
		return nil, err
	}
	labels := make([]string, labelNamesList.Len())
	for i := 0; i < len(labels); i++ {
		labels[i], err = labelNamesList.At(i)
		if err != nil {
			METRICS_SERVICE_CONFIG_ERROR.Log(Logger().Fatal()).Err(err).Msgf("Failed to read CounterVectorMetricSpec.LabelNames[%d] field", i)
		}
	}
	return &CounterVectorMetricSpec{MetricSpec: MetricSpec(counterMetricSpec), DynamicLabels: labels}, nil
}

type CounterVectorMetricSpec MetricVectorSpec

func (a CounterVectorMetricSpec) CounterOpts() prometheus.CounterOpts {
	return prometheus.CounterOpts{
		Name:        a.MetricID.Hex(),
		Help:        a.Help,
		ConstLabels: a.ServiceID.MetricSpecLabels(),
	}
}

func NewGaugeMetricSpec(spec config.GaugeMetricSpec) (GaugeMetricSpec, error) {
	help, err := spec.Help()
	if err != nil {
		return GaugeMetricSpec{}, err
	}
	return GaugeMetricSpec{
		ServiceID: ServiceID(spec.ServiceId()),
		MetricID:  MetricID(spec.MetricId()),
		Help:      help,
	}, nil
}

type GaugeMetricSpec MetricSpec

func (a GaugeMetricSpec) GaugeOpts() prometheus.GaugeOpts {
	return prometheus.GaugeOpts{
		Name:        a.MetricID.Hex(),
		Help:        a.Help,
		ConstLabels: a.ServiceID.MetricSpecLabels(),
	}
}

func NewGaugeVectorMetricSpec(spec config.GaugeVectorMetricSpec) (*GaugeVectorMetricSpec, error) {
	gaugeSpec, err := spec.MetricSpec()
	if err != nil {
		return nil, err
	}
	gaugeMetricSpec, err := NewGaugeMetricSpec(gaugeSpec)
	if err != nil {
		return nil, err
	}
	labelNamesList, err := spec.LabelNames()
	if err != nil {
		return nil, err
	}
	labels := make([]string, labelNamesList.Len())
	for i := 0; i < len(labels); i++ {
		labels[i], err = labelNamesList.At(i)
		if err != nil {
			METRICS_SERVICE_CONFIG_ERROR.Log(Logger().Fatal()).Err(err).Msgf("Failed to read GaugeVectorMetricSpec.LabelNames[%d] field", i)
		}
	}
	return &GaugeVectorMetricSpec{MetricSpec: MetricSpec(gaugeMetricSpec), DynamicLabels: labels}, nil
}

type GaugeVectorMetricSpec MetricVectorSpec

func (a GaugeVectorMetricSpec) GaugeOpts() prometheus.GaugeOpts {
	return prometheus.GaugeOpts{
		Name:        a.MetricID.Hex(),
		Help:        a.Help,
		ConstLabels: a.ServiceID.MetricSpecLabels(),
	}
}

func NewHistogramMetricSpec(spec config.HistogramMetricSpec) (HistogramMetricSpec, error) {
	metricSpec := func() (MetricSpec, error) {
		help, err := spec.Help()
		if err != nil {
			return MetricSpec{}, err
		}
		return MetricSpec{
			ServiceID: ServiceID(spec.ServiceId()),
			MetricID:  MetricID(spec.MetricId()),
			Help:      help,
		}, nil
	}

	bucketList, err := spec.Buckets()
	if err != nil {
		return HistogramMetricSpec{}, err
	}

	buckets := make([]float64, bucketList.Len())
	for i := 0; i < len(buckets); i++ {
		buckets[i] = bucketList.At(i)
	}
	histogramMetricSpec, err := metricSpec()
	if err != nil {
		return HistogramMetricSpec{}, err
	}
	return HistogramMetricSpec{
		MetricSpec: histogramMetricSpec,
		Buckets:    buckets,
	}, nil
}

type HistogramMetricSpec struct {
	MetricSpec
	Buckets []float64
}

func (a HistogramMetricSpec) HistogramOpts() prometheus.HistogramOpts {
	return prometheus.HistogramOpts{
		Name:        a.MetricID.Hex(),
		Help:        a.Help,
		ConstLabels: a.ServiceID.MetricSpecLabels(),
		Buckets:     a.Buckets,
	}
}

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
	return &HistogramVectorMetricSpec{
		MetricVectorSpec: &MetricVectorSpec{metricSpec.MetricSpec, labels},
		Buckets:          metricSpec.Buckets,
	}, nil
}

type HistogramVectorMetricSpec struct {
	*MetricVectorSpec
	Buckets []float64
}

func (a HistogramVectorMetricSpec) HistogramOpts() prometheus.HistogramOpts {
	return prometheus.HistogramOpts{
		Name:        a.MetricID.Hex(),
		Help:        a.Help,
		ConstLabels: a.ServiceID.MetricSpecLabels(),
		Buckets:     a.Buckets,
	}
}

type CounterMetric struct {
	*CounterMetricSpec
	prometheus.Counter
}

func (a *CounterMetric) Register() {
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

type CounterVectorMetric struct {
	*CounterVectorMetricSpec
	*prometheus.CounterVec
}

func (a *CounterVectorMetric) Register() {
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

type GaugeMetric struct {
	*GaugeMetricSpec
	prometheus.Gauge
}

func (a *GaugeMetric) Register() {
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

type GaugeVectorMetric struct {
	*GaugeVectorMetricSpec
	*prometheus.GaugeVec
}

func (a *GaugeVectorMetric) Register() {
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

type HistogramMetric struct {
	*HistogramMetricSpec
	prometheus.Histogram
}

func (a *HistogramMetric) Register() {
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

type HistogramVectorMetric struct {
	*HistogramVectorMetricSpec
	*prometheus.HistogramVec
}

func (a *HistogramVectorMetric) Register() {
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
