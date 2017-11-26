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

// Gauge looks up the registered metric.
func (a AppMetricRegistry) Gauge(serviceId ServiceID, metricID MetricID) *GaugeMetric {
	metrics := gauges[serviceId]
	if metrics == nil {
		return nil
	}
	return metrics[metricID]
}

// GaugeVector looks up the registered metric.
func (a AppMetricRegistry) GaugeVector(serviceId ServiceID, metricID MetricID) *GaugeVectorMetric {
	metrics := gaugeVectors[serviceId]
	if metrics == nil {
		return nil
	}
	return metrics[metricID]
}

// GaugeMetricIds returns all gauge metric ids registered for the service
func (a AppMetricRegistry) GaugeMetricIds(serviceId ServiceID) []MetricID {
	metrics := gauges[serviceId]
	metricIds := make([]MetricID, len(metrics))
	i := 0
	for id := range metrics {
		metricIds[i] = id
		i++
	}
	return metricIds
}

// GaugeVectorMetricIds returns all gauge vector metric ids registered for the service
func (a AppMetricRegistry) GaugeVectorMetricIds(serviceId ServiceID) []MetricID {
	metrics := gaugeVectors[serviceId]
	metricIds := make([]MetricID, len(metrics))
	i := 0
	for id := range metrics {
		metricIds[i] = id
		i++
	}
	return metricIds
}

// GaugeMetricsByService returns all gauge metric ids grouped by service
func (a AppMetricRegistry) GaugeMetricsByService() map[ServiceID][]MetricID {
	m := map[ServiceID][]MetricID{}
	i := 0
	for serviceId, metrics := range gauges {
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

// GaugeVectorMetricsByService returns all gauge vector metric ids grouped by service
func (a AppMetricRegistry) GaugeVectorMetricsByService() map[ServiceID][]MetricID {
	m := map[ServiceID][]MetricID{}
	i := 0
	for serviceId, metrics := range gaugeVectors {
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
