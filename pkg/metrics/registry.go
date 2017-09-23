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
	"github.com/prometheus/client_golang/prometheus"

	"sync"

	"github.com/oysterpack/oysterpack.go/pkg/commons/os"
)

var (
	mutex sync.RWMutex

	// Registry is the global registry
	Registry *prometheus.Registry = NewRegistry(true)

	countersMap    = map[string]*Counter{}
	counterVecsMap = map[string]*CounterVec{}

	gaugesMap    = map[string]*Gauge{}
	gaugeVecsMap = map[string]*GaugeVec{}

	histogramsMap    = map[string]*Histogram{}
	histogramVecsMap = map[string]*HistogramVec{}

	summariesMap   = map[string]*Summary{}
	summaryVecsMap = map[string]*SummaryVec{}
)

// NewRegistry creates a new registry.
// If collectProcessMetrics = true, then the prometheus GoCollector and ProcessCollectors are registered.
func NewRegistry(collectProcessMetrics bool) *prometheus.Registry {
	registry := prometheus.NewRegistry()
	if collectProcessMetrics {
		registry.MustRegister(
			prometheus.NewGoCollector(),
			prometheus.NewProcessCollector(os.PID(), ""),
		)
	}
	return registry
}

// ResetRegistry resets the prometheus Registry and clears all cached metrics
func ResetRegistry() {
	mutex.Lock()
	defer mutex.Unlock()
	Registry = NewRegistry(true)
	countersMap = map[string]*Counter{}
	counterVecsMap = map[string]*CounterVec{}
	gaugesMap = map[string]*Gauge{}
	gaugeVecsMap = map[string]*GaugeVec{}
	histogramsMap = map[string]*Histogram{}
	histogramVecsMap = map[string]*HistogramVec{}
	summariesMap = map[string]*Summary{}
	summaryVecsMap = map[string]*SummaryVec{}
}

// Registered returns true if a metric is registered with the same name
func Registered(name string) bool {
	mutex.RLock()
	defer mutex.RUnlock()
	return registered(name)
}

func registered(name string) bool {
	if _, exists := countersMap[name]; exists {
		return true
	}
	if _, exists := counterVecsMap[name]; exists {
		return true
	}

	if _, exists := gaugesMap[name]; exists {
		return true
	}
	if _, exists := gaugeVecsMap[name]; exists {
		return true
	}

	if _, exists := histogramsMap[name]; exists {
		return true
	}
	if _, exists := histogramVecsMap[name]; exists {
		return true
	}

	if _, exists := summariesMap[name]; exists {
		return true
	}
	if _, exists := summaryVecsMap[name]; exists {
		return true
	}

	return false
}
