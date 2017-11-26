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
	"testing"

	"zombiezen.com/go/capnproto2"

	"github.com/oysterpack/oysterpack.go/pkg/app/config"
)

func TestMetricsService(t *testing.T) {
	previousConfigDir := configDir
	configDir = "testdata/TestMetricsService"

	// clean the config dir
	os.RemoveAll(configDir)
	os.MkdirAll(configDir, 0755)

	// Reset the app
	Reset()
	defer func() {
		configDir = previousConfigDir
		Reset()
	}()

	const (
		SERVICE_ID_1 = ServiceID(1 + iota)
		SERVICE_ID_2
	)
	// metric ids
	const (
		METRIC_ID_COUNTER = MetricID(1 + iota)
		METRIC_ID_COUNTER_VEC
		METRIC_ID_GAUGE
		METRIC_ID_GAUGE_VEC
		METRIC_ID_HISTOGRAM
		METRIC_ID_HISTOGRAM_VEC
	)

	t.Run("No Metrics Registered", func(t *testing.T) {
		if MetricRegistry.Counter(SERVICE_ID_1, METRIC_ID_COUNTER) != nil {
			t.Error("No metric should be registered")
		}
		if MetricRegistry.Counter(SERVICE_ID_1, METRIC_ID_COUNTER_VEC) != nil {
			t.Error("No metric should be registered")
		}
		if MetricRegistry.Counter(SERVICE_ID_1, METRIC_ID_GAUGE) != nil {
			t.Error("No metric should be registered")
		}
		if MetricRegistry.Counter(SERVICE_ID_1, METRIC_ID_GAUGE_VEC) != nil {
			t.Error("No metric should be registered")
		}
		if MetricRegistry.Counter(SERVICE_ID_1, METRIC_ID_HISTOGRAM) != nil {
			t.Error("No metric should be registered")
		}
		if MetricRegistry.Counter(SERVICE_ID_1, METRIC_ID_HISTOGRAM_VEC) != nil {
			t.Error("No metric should be registered")
		}

		if len(MetricRegistry.CounterMetricIds(SERVICE_ID_1)) > 0 {
			t.Error("No metric should be registered")
		}
		if len(MetricRegistry.CounterVectorMetricIds(SERVICE_ID_1)) > 0 {
			t.Error("No metric should be registered")
		}
		if len(MetricRegistry.GaugeMetricIds(SERVICE_ID_1)) > 0 {
			t.Error("No metric should be registered")
		}
		if len(MetricRegistry.GaugeVectorMetricIds(SERVICE_ID_1)) > 0 {
			t.Error("No metric should be registered")
		}
		if len(MetricRegistry.HistogramMetricIds(SERVICE_ID_1)) > 0 {
			t.Error("No metric should be registered")
		}
		if len(MetricRegistry.HistogramVectorMetricIds(SERVICE_ID_1)) > 0 {
			t.Error("No metric should be registered")
		}
	})

	t.Run("1 service with 1 metric for each metric type", func(t *testing.T) {
		service := NewService(SERVICE_ID_1)
		Services.Register(service)

		msg, s, err := capnp.NewMessage(capnp.SingleSegment(nil))
		if err != nil {
			t.Fatal(err)
		}

		metricsServiceSpec, err := config.NewRootMetricsServiceSpec(s)
		if err != nil {
			t.Fatal(err)
		}
		metricsServiceSpec.SetHttpPort(4455)
		metricsSpec, err := config.NewMetricSpecs(s)
		if err != nil {
			t.Fatal(err)
		}
		metricsServiceSpec.SetMetricSpecs(metricsSpec)

		setCounterSpecs := func() {
			metricSpec, err := config.NewCounterMetricSpec(s)
			if err != nil {
				t.Fatal(err)
			}
			metricSpec.SetServiceId(uint64(SERVICE_ID_1))
			metricSpec.SetMetricId(uint64(METRIC_ID_COUNTER))
			if err := metricSpec.SetHelp("SERVICE_ID_1 METRIC_ID_COUNTER"); err != nil {
				t.Fatal(err)
			}
			counterSpecs, err := metricsSpec.NewCounterSpecs(1)
			counterSpecs.Set(0, metricSpec)
		}

		setCounterVectorSpecs := func() {
			vectorMetricSpec, err := config.NewCounterVectorMetricSpec(s)

			labelNames, err := vectorMetricSpec.NewLabelNames(2)
			if err != nil {
				t.Fatal(err)
			}
			labelNames.Set(0, "a")
			labelNames.Set(1, "b")
			metricSpec, err := config.NewCounterMetricSpec(s)
			if err != nil {
				t.Fatal(err)
			}
			metricSpec.SetServiceId(uint64(SERVICE_ID_1))
			metricSpec.SetMetricId(uint64(METRIC_ID_COUNTER_VEC))
			if err := metricSpec.SetHelp("SERVICE_ID_1 METRIC_ID_COUNTER_VEC"); err != nil {
				t.Fatal(err)
			}
			vectorMetricSpec.SetMetricSpec(metricSpec)

			vectorSpecs, err := metricsSpec.NewCounterVectorSpecs(1)
			vectorSpecs.Set(0, vectorMetricSpec)
		}

		setGaugeSpecs := func() {
			metricSpec, err := config.NewGaugeMetricSpec(s)
			if err != nil {
				t.Fatal(err)
			}
			metricSpec.SetServiceId(uint64(SERVICE_ID_1))
			metricSpec.SetMetricId(uint64(METRIC_ID_GAUGE))
			if err := metricSpec.SetHelp("SERVICE_ID_1 METRIC_ID_GAUGE"); err != nil {
				t.Fatal(err)
			}
			specs, err := metricsSpec.NewGaugeSpecs(1)
			specs.Set(0, metricSpec)
		}

		setGaugeVectorSpecs := func() {
			vectorMetricSpec, err := config.NewGaugeVectorMetricSpec(s)

			labelNames, err := vectorMetricSpec.NewLabelNames(2)
			if err != nil {
				t.Fatal(err)
			}
			labelNames.Set(0, "a")
			labelNames.Set(1, "b")
			metricSpec, err := config.NewGaugeMetricSpec(s)
			if err != nil {
				t.Fatal(err)
			}
			metricSpec.SetServiceId(uint64(SERVICE_ID_1))
			metricSpec.SetMetricId(uint64(METRIC_ID_GAUGE_VEC))
			if err := metricSpec.SetHelp("SERVICE_ID_1 METRIC_ID_GAUGE_VEC"); err != nil {
				t.Fatal(err)
			}
			vectorMetricSpec.SetMetricSpec(metricSpec)

			vectorSpecs, err := metricsSpec.NewGaugeVectorSpecs(1)
			vectorSpecs.Set(0, vectorMetricSpec)
		}

		setHistogramSpecs := func() {
			metricSpec, err := config.NewHistogramMetricSpec(s)
			if err != nil {
				t.Fatal(err)
			}
			metricSpec.SetServiceId(uint64(SERVICE_ID_1))
			metricSpec.SetMetricId(uint64(METRIC_ID_HISTOGRAM))
			if err := metricSpec.SetHelp("SERVICE_ID_1 METRIC_ID_HISTOGRAM"); err != nil {
				t.Fatal(err)
			}
			buckets, err := metricSpec.NewBuckets(3)
			if err != nil {
				t.Fatal(err)
			}
			buckets.Set(0, 100)
			buckets.Set(1, 200)
			buckets.Set(2, 300)
			specs, err := metricsSpec.NewHistogramSpecs(1)
			specs.Set(0, metricSpec)
		}

		setHistogramVectorSpecs := func() {
			vectorMetricSpec, err := config.NewHistogramVectorMetricSpec(s)

			labelNames, err := vectorMetricSpec.NewLabelNames(2)
			if err != nil {
				t.Fatal(err)
			}
			labelNames.Set(0, "a")
			labelNames.Set(1, "b")
			metricSpec, err := config.NewHistogramMetricSpec(s)
			if err != nil {
				t.Fatal(err)
			}
			metricSpec.SetServiceId(uint64(SERVICE_ID_1))
			metricSpec.SetMetricId(uint64(METRIC_ID_HISTOGRAM_VEC))
			if err := metricSpec.SetHelp("SERVICE_ID_1 METRIC_ID_GAUGE_VEC"); err != nil {
				t.Fatal(err)
			}
			buckets, err := metricSpec.NewBuckets(3)
			if err != nil {
				t.Fatal(err)
			}
			buckets.Set(0, 100)
			buckets.Set(1, 200)
			buckets.Set(2, 300)
			vectorMetricSpec.SetMetricSpec(metricSpec)

			vectorSpecs, err := metricsSpec.NewHistogramVectorSpecs(1)
			vectorSpecs.Set(0, vectorMetricSpec)
		}

		setCounterSpecs()
		setCounterVectorSpecs()
		setGaugeSpecs()
		setGaugeVectorSpecs()
		setHistogramSpecs()
		setHistogramVectorSpecs()

		configFile, err := os.Create(Configs.ServiceConfigPath(METRICS_SERVICE_ID))
		MarshalCapnpMessage(msg, configFile)
		configFile.Close()

		Reset()

		if metric := MetricRegistry.Counter(SERVICE_ID_1, METRIC_ID_COUNTER); metric == nil {
			t.Error("Metric should have been registered")
		} else {
			t.Logf("ServiceID = %v, MetricID = %q, Help = %q", metric.ServiceID.Hex(), metric.PrometheusName(metric.ServiceID), metric.Help)
		}
		if len(MetricRegistry.CounterMetricIds(SERVICE_ID_1)) != 1 {
			t.Errorf("1 Metric should have been registered : %v", MetricRegistry.CounterMetricIds(SERVICE_ID_1))
		}

		if metric := MetricRegistry.CounterVector(SERVICE_ID_1, METRIC_ID_COUNTER_VEC); metric == nil {
			t.Error("Metric should have been registered")
		} else {
			t.Logf("ServiceID = %v, MetricID = %q, Help = %q", metric.ServiceID.Hex(), metric.MetricID.PrometheusName(metric.ServiceID), metric.Help)
		}
		if len(MetricRegistry.CounterVectorMetricIds(SERVICE_ID_1)) != 1 {
			t.Errorf("1 Metric should have been registered : %v", MetricRegistry.CounterVectorMetricIds(SERVICE_ID_1))
		}

		if metric := MetricRegistry.Gauge(SERVICE_ID_1, METRIC_ID_GAUGE); metric == nil {
			t.Error("Metric should have been registered")
		} else {
			t.Logf("ServiceID = %v, MetricID = %q, Help = %q", metric.ServiceID.Hex(), metric.PrometheusName(metric.ServiceID), metric.Help)
		}
		if len(MetricRegistry.GaugeMetricIds(SERVICE_ID_1)) != 1 {
			t.Errorf("1 Metric should have been registered : %v", MetricRegistry.GaugeMetricIds(SERVICE_ID_1))
		}

		if metric := MetricRegistry.GaugeVector(SERVICE_ID_1, METRIC_ID_GAUGE_VEC); metric == nil {
			t.Error("Metric should have been registered")
		} else {
			t.Logf("ServiceID = %v, MetricID = %q, Help = %q", metric.ServiceID.Hex(), metric.MetricID.PrometheusName(metric.ServiceID), metric.Help)
		}
		if len(MetricRegistry.GaugeVectorMetricIds(SERVICE_ID_1)) != 1 {
			t.Errorf("1 Metric should have been registered : %v", MetricRegistry.GaugeVectorMetricsByService())
		}

		if metric := MetricRegistry.Histogram(SERVICE_ID_1, METRIC_ID_HISTOGRAM); metric == nil {
			t.Error("Metric should have been registered")
		} else {
			t.Logf("ServiceID = %v, MetricID = %q, Help = %q", metric.ServiceID.Hex(), metric.MetricID.PrometheusName(metric.ServiceID), metric.Help)
		}
		if len(MetricRegistry.HistogramMetricIds(SERVICE_ID_1)) != 1 {
			t.Errorf("1 Metric should have been registered : %v", MetricRegistry.HistogramMetricIds(SERVICE_ID_1))
		}

		if metric := MetricRegistry.HistogramVector(SERVICE_ID_1, METRIC_ID_HISTOGRAM_VEC); metric == nil {
			t.Error("Metric should have been registered")
		} else {
			t.Logf("ServiceID = %v, MetricID = %q, Help = %q", metric.ServiceID.Hex(), metric.MetricID.PrometheusName(metric.ServiceID), metric.Help)
		}
		if len(MetricRegistry.HistogramVectorMetricIds(SERVICE_ID_1)) != 1 {
			t.Errorf("1 Metric should have been registered : %v", MetricRegistry.HistogramVectorMetricIds(SERVICE_ID_1))
		}
	})

}
