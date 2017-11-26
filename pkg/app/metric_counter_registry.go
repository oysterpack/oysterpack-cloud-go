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

// Counter looks up the registered metric.
func (a AppMetricRegistry) Counter(serviceId ServiceID, metricID MetricID) *CounterMetric {
	metrics := counters[serviceId]
	if metrics == nil {
		return nil
	}
	return metrics[metricID]
}

// CounterVec looks up the registered metric.
func (a AppMetricRegistry) CounterVector(serviceId ServiceID, metricID MetricID) *CounterVectorMetric {
	metrics := counterVectors[serviceId]
	if metrics == nil {
		return nil
	}
	return metrics[metricID]
}

// CounterMetricIds returns the registered metric ids for the specified service id
func (a AppMetricRegistry) CounterMetricIds(serviceId ServiceID) []MetricID {
	metrics := counters[serviceId]
	metricIds := make([]MetricID, len(metrics))
	i := 0
	for id := range metrics {
		metricIds[i] = id
		i++
	}
	return metricIds
}

// CounterVectorMetricIds returns the registered metric ids for the specified service id
func (a AppMetricRegistry) CounterVectorMetricIds(serviceId ServiceID) []MetricID {
	metrics := counterVectors[serviceId]
	metricIds := make([]MetricID, len(metrics))
	i := 0
	for id := range metrics {
		metricIds[i] = id
		i++
	}
	return metricIds
}

// CounterMetricsIdsPerService returns all counter metric ids grouped by service
func (a AppMetricRegistry) CounterMetricsIdsPerService() map[ServiceID][]MetricID {
	m := map[ServiceID][]MetricID{}
	i := 0
	for serviceId, metrics := range counters {
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

// CounterMetricsIdsPerService returns all counter metric ids grouped by service
func (a AppMetricRegistry) CounterVectorMetricsIdsPerService() map[ServiceID][]MetricID {
	m := map[ServiceID][]MetricID{}
	i := 0
	for serviceId, metrics := range counterVectors {
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
