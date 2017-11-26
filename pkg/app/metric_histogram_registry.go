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

// Histogram looks up a registered HistogramMetric
func Histogram(serviceId ServiceID, metricID MetricID) *HistogramMetric {
	metrics := histograms[serviceId]
	if metrics == nil {
		return nil
	}
	return metrics[metricID]
}

// HistogramVector looks up a registered HistogramVectorMetric
func HistogramVector(serviceId ServiceID, metricID MetricID) *HistogramVectorMetric {
	metrics := histogramVectors[serviceId]
	if metrics == nil {
		return nil
	}
	return metrics[metricID]
}

// HistogramMetricIds returns all histogram metric ids for the specified service
func HistogramMetricIds(serviceId ServiceID) []MetricID {
	metrics := histograms[serviceId]
	metricIds := make([]MetricID, len(metrics))
	i := 0
	for id := range metrics {
		metricIds[i] = id
		i++
	}
	return metricIds
}

// HistogramVectorMetricIds returns all histogram vector metric ids for the specified service
func HistogramVectorMetricIds(serviceId ServiceID) []MetricID {
	metrics := histogramVectors[serviceId]
	metricIds := make([]MetricID, len(metrics))
	i := 0
	for id := range metrics {
		metricIds[i] = id
		i++
	}
	return metricIds
}

// HistogramMetricsByService returns all gauge metric ids grouped by service
func HistogramMetricsByService() map[ServiceID][]MetricID {
	m := map[ServiceID][]MetricID{}
	i := 0
	for serviceId, metrics := range histograms {
		metricIds := make([]MetricID, len(metrics))
		i = 0
		for metricId := range metrics {
			metricIds[i] = metricId
			i++
		}
		m[serviceId] = metricIds
	}
	return m
}

// HistogramVectorMetricsByService returns all gauge vector metric ids grouped by service
func HistogramVectorMetricsByService() map[ServiceID][]MetricID {
	m := map[ServiceID][]MetricID{}
	i := 0
	for serviceId, metrics := range histogramVectors {
		metricIds := make([]MetricID, len(metrics))
		i = 0
		for metricId := range metrics {
			metricIds[i] = metricId
			i++
		}
		m[serviceId] = metricIds
	}
	return m
}
