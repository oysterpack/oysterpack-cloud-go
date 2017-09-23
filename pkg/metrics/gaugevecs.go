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

// GetOrMustRegisterGaugeVec first checks if a gaugeVec with the same name is already registered.
// If the gaugeVec is already registered, and was registered with the same opts, then the cached gaugeVec is returned.
// If the gaugeVec is already registered, and was registered with the different opts, then a panic is triggered.
// If not such gaugeVec exists, then it is registered and cached along with its opts.
func GetOrMustRegisterGaugeVec(opts *GaugeVecOpts) *prometheus.GaugeVec {
	const FUNC = "GetOrMustRegisterGaugeVec"
	mutex.RLock()
	defer mutex.RUnlock()
	name := GaugeFQName(opts.GaugeOpts)
	if gaugeVec := gaugeVecsMap[name]; gaugeVec != nil {
		if GaugeVecOptsMatch(opts, gaugeVec.GaugeVecOpts) {
			return gaugeVec.GaugeVec

		}
		logger.Panic().Str(logging.FUNC, FUNC).
			Str("registered", fmt.Sprintf("%v", gaugeVec.GaugeVecOpts)).
			Str("dup", fmt.Sprintf("%v", opts)).
			Err(ErrMetricAlreadyRegisteredWithDifferentOpts).
			Msg("")
	}

	if registered(name) {
		logger.Panic().Str(logging.FUNC, FUNC).
			Str("name", name).
			Int("type", GAUGEVEC.Value()).
			Int("registered_type", GAUGE.Value()).
			Err(ErrMetricNameUsedByDifferentMetricType).
			Msg("")
	}

	gaugeVec := prometheus.NewGaugeVec(*opts.GaugeOpts, opts.Labels)
	Registry.MustRegister(gaugeVec)
	gaugeVecsMap[name] = &GaugeVec{gaugeVec, opts}
	return gaugeVec
}

// GaugeVecNames returns names of all registered gaugeVecs
func GaugeVecNames() []string {
	mutex.RLock()
	defer mutex.RUnlock()
	names := make([]string, len(gaugeVecsMap))
	i := 0
	for k := range gaugeVecsMap {
		names[i] = k
		i++
	}
	return names
}

// GaugeVecs returns all registered gaugeVecs
func GaugeVecs() []*GaugeVec {
	c := make([]*GaugeVec, len(gaugeVecsMap))
	i := 0
	for _, v := range gaugeVecsMap {
		c[i] = v
		i++
	}
	return c
}

// GetGaugeVec looks up the gaugeVec by its fully qualified name
func GetGaugeVec(name string) *GaugeVec {
	return gaugeVecsMap[name]
}
