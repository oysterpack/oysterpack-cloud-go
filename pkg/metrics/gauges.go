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
	"fmt"

	"github.com/oysterpack/oysterpack.go/pkg/logging"
	"github.com/prometheus/client_golang/prometheus"
)

// GetOrMustRegisterGauge first checks if a gauge with the same name is already registered.
// If the gauge is already registered, and was registered with the same opts, then the cached metric is returned.
// If the gauge is already registered, and was registered with the different opts, then a panic is triggered.
// If no such gauge exists, then it is registered and cached along with its opts.
func GetOrMustRegisterGauge(opts *prometheus.GaugeOpts) prometheus.Gauge {
	const FUNC = "GetOrMustRegisterGauge"
	mutex.RLock()
	defer mutex.RUnlock()
	name := GaugeFQName(opts)
	if gauge := gaugesMap[name]; gauge != nil {
		if GaugeOptsMatch(opts, gauge.GaugeOpts) {
			return gauge
		}
		logger.Panic().Str(logging.FUNC, FUNC).
			Str("registered", fmt.Sprintf("%v", gauge.GaugeOpts)).
			Str("dup", fmt.Sprintf("%v", opts)).
			Err(ErrMetricAlreadyRegisteredWithDifferentOpts).
			Msg("")
	}

	if registered(name) {
		logger.Panic().Str(logging.FUNC, FUNC).
			Str("name", name).
			Int("type", GAUGE.Value()).
			Int("registered_type", GAUGEVEC.Value()).
			Err(ErrMetricNameUsedByDifferentMetricType).
			Msg("")
	}

	gauge := prometheus.NewGauge(*opts)
	Registry.MustRegister(gauge)
	gaugesMap[name] = &Gauge{gauge, opts}
	return gauge
}

// GetOrMustRegisterGaugeFunc registers a GaugeFunc, but first checks if a gauge with the same name is already registered.
// If the gauge is already registered, and was registered with the same opts, then the cached metric is returned.
// If the gauge is already registered, and was registered with the different opts, then a panic is triggered.
// If no such gauge exists, then it is registered and cached along with its opts.
func GetOrMustRegisterGaugeFunc(opts *prometheus.GaugeOpts, f func() float64) prometheus.GaugeFunc {
	const FUNC = "GetOrMustRegisterGauge"
	mutex.RLock()
	defer mutex.RUnlock()
	name := GaugeFQName(opts)
	if gaugeFunc := gaugeFuncsMap[name]; gaugeFunc != nil {
		if GaugeOptsMatch(opts, gaugeFunc.GaugeOpts) {
			return gaugeFunc
		}
		logger.Panic().Str(logging.FUNC, FUNC).
			Str("registered", fmt.Sprintf("%v", gaugeFunc.GaugeOpts)).
			Str("dup", fmt.Sprintf("%v", opts)).
			Err(ErrMetricAlreadyRegisteredWithDifferentOpts).
			Msg("")
	}

	if registered(name) {
		logger.Panic().Str(logging.FUNC, FUNC).
			Str("name", name).
			Int("type", GAUGE.Value()).
			Int("registered_type", GAUGEVEC.Value()).
			Err(ErrMetricNameUsedByDifferentMetricType).
			Msg("")
	}

	gaugeFunc := prometheus.NewGaugeFunc(*opts, f)
	Registry.MustRegister(gaugeFunc)
	gaugeFuncsMap[name] = &GaugeFunc{gaugeFunc, opts}
	return gaugeFunc
}

// GaugeNames returns names of all registered gauges
func GaugeNames() []string {
	mutex.RLock()
	defer mutex.RUnlock()
	names := make([]string, len(gaugesMap))
	i := 0
	for k := range gaugesMap {
		names[i] = k
		i++
	}
	return names
}

// Gauges returns all registered gauges
func Gauges() []*Gauge {
	mutex.RLock()
	defer mutex.RUnlock()
	c := make([]*Gauge, len(gaugesMap))
	i := 0
	for _, v := range gaugesMap {
		c[i] = v
		i++
	}
	return c
}

// GetGauge looks up the status by its fully qualified name
func GetGauge(name string) *Gauge {
	mutex.RLock()
	defer mutex.RUnlock()
	return gaugesMap[name]
}
