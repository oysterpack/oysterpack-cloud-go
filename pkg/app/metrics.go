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

	"errors"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
	"zombiezen.com/go/capnproto2"
)

const (
	METRICS_SERVICE_ID               = ServiceID(0xe3054017c1b1d214)
	DEFAULT_METRICS_HTTP_PORT uint16 = 4444
)

// metrics
var (
	metricsServiceMutex        sync.Mutex
	registerMetricsHandlerOnce sync.Once

	// global metrics registry
	metricsRegistry = newMetricsRegistry(true)

	counters       = make(map[ServiceID]map[MetricID]*CounterMetric)
	counterVectors = make(map[ServiceID]map[MetricID]*CounterVectorMetric)

	gauges       = make(map[ServiceID]map[MetricID]*GaugeMetric)
	gaugeVectors = make(map[ServiceID]map[MetricID]*GaugeVectorMetric)

	histograms       = make(map[ServiceID]map[MetricID]*HistogramMetric)
	histogramVectors = make(map[ServiceID]map[MetricID]*HistogramVectorMetric)
)

type AppMetricRegistry struct{}

// NewRegistry creates a new registry.
// If collectProcessMetrics = true, then the prometheus GoCollector and ProcessCollectors are registered.
func newMetricsRegistry(collectProcessMetrics bool) *prometheus.Registry {
	registry := prometheus.NewRegistry()
	if collectProcessMetrics {
		registry.MustRegister(
			prometheus.NewGoCollector(),
			prometheus.NewProcessCollector(os.Getpid(), ""),
		)
	}
	return registry
}

// MetricSpec specifies a service metric.
// Metrics are scoped to services.
//
// Metric naming convention : op_{ServiceID.Hex()}_{MetricID.Hex()}
//                            op_e49214fa20b35ba8_db722ec1bb1b5766
// where op stands for OysterPack
//
// Think of ServiceID and MetricID together as the unique key for a Service metric.
//
// Each metric will have the following constant labels
//	- domain : DomainID.Hex()
//	- app 	 : AppID.Hex
//	- svc 	 : ServiceID.Hex()
type MetricSpec struct {
	ServiceID
	MetricID
	Help string
}

// PrometheusName naming convention : op_{ServiceID.Hex()}_{MetricID.Hex()}
// e.g., op_e49214fa20b35ba8_db722ec1bb1b5766
func PrometheusName(metricSpec *MetricSpec) string {
	return metricSpec.MetricID.PrometheusName(metricSpec.ServiceID)
}

// MetricSpecLabels returns the following labels:
//	- domain : DomainID.Hex()
//	- app 	 : AppID.Hex
//	- svc 	 : ServiceID.Hex()
func MetricSpecLabels(serviceID ServiceID) prometheus.Labels {
	return prometheus.Labels{
		"domain": Domain().Hex(),
		"app":    ID().Hex(),
		"svc":    serviceID.Hex(),
	}
}

// MetricVectorSpec specifies a service metric vector.
// The difference between a simple metric and a metric vector is that a metric vector can have dimensions, which are defined
// by dynamic labels. The labels are called dynamic because the the label value is not constant.
//
// When a metric is recorded for a metric vector, all label values must be specified.
type MetricVectorSpec struct {
	MetricSpec
	DynamicLabels []string
}

// used to duck type metrics to enable centralizing common generic metric functions
type metricSpec interface {
	HasHelp() bool

	ServiceId() uint64

	MetricId() uint64

	Help() (string, error)
}

func validateMetricSpec(spec metricSpec) (help string, err error) {
	if !spec.HasHelp() {
		return "", ConfigError(APP_SERVICE, errors.New("Help is required"), "")
	}
	if spec.ServiceId() == 0 {
		return "", ConfigError(APP_SERVICE, errors.New("ServiceID must be > 0"), "")
	}
	if spec.MetricId() == 0 {
		return "", ConfigError(APP_SERVICE, errors.New("MetricID must be > 0"), "")
	}

	help, err = spec.Help()
	if err != nil {
		return "", err
	}
	help = strings.TrimSpace(help)
	if help == "" {
		return "", ConfigError(APP_SERVICE, errors.New("Help is required"), "")
	}

	return help, nil
}

// used to duck type metrics to enable centralizing common generic metric functions
type metricVectorSpec interface {
	HasMetricSpec() bool
	HasLabelNames() bool
	LabelNames() (capnp.TextList, error)
}

func validateMetricVectorSpec(spec metricVectorSpec) (labels []string, err error) {
	if !spec.HasMetricSpec() {
		return nil, ConfigError(APP_SERVICE, errors.New("MetricSpec is required"), "")
	}
	if !spec.HasLabelNames() {
		return nil, ConfigError(APP_SERVICE, errors.New("LabelNames is required"), "")
	}
	labelNamesList, err := spec.LabelNames()
	if err != nil {
		return nil, err
	}
	if labelNamesList.Len() == 0 {
		return nil, ConfigError(APP_SERVICE, errors.New("At least 1 label name is required"), "")
	}
	labels = make([]string, labelNamesList.Len())
	for i := 0; i < len(labels); i++ {
		labels[i], err = labelNamesList.At(i)
		if err != nil {
			return
		}
	}
	return labels, nil
}
