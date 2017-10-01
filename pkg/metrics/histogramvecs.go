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

// GetOrMustRegisterHistogramVec first checks if a histogramVec with the same name is already registered.
// If the histogramVec is already registered, and was registered with the same opts, then the cached histogramVec is returned.
// If the histogramVec is already registered, and was registered with the different opts, then a panic is triggered.
// If not such histogramVec exists, then it is registered and cached along with its opts.
func GetOrMustRegisterHistogramVec(opts *HistogramVecOpts) *prometheus.HistogramVec {
	const FUNC = "GetOrMustRegisterHistogramVec"
	mutex.RLock()
	defer mutex.RUnlock()
	name := HistogramFQName(opts.HistogramOpts)
	if histogramVec := histogramVecsMap[name]; histogramVec != nil {
		if HistogramVecOptsMatch(opts, histogramVec.HistogramVecOpts) {
			return histogramVec.HistogramVec

		}
		logger.Panic().Str(logging.FUNC, FUNC).
			Str("registered", fmt.Sprintf("%v", histogramVec.HistogramVecOpts)).
			Str("dup", fmt.Sprintf("%v", opts)).
			Err(ErrMetricAlreadyRegisteredWithDifferentOpts).
			Msg("")
	}

	if registered(name) {
		logger.Panic().Str(logging.FUNC, FUNC).
			Str("name", name).
			Int("type", HISTOGRAMVEC.Value()).
			Int("registered_type", HISTOGRAM.Value()).
			Err(ErrMetricNameUsedByDifferentMetricType).
			Msg("")
	}

	histogramVec := prometheus.NewHistogramVec(*opts.HistogramOpts, opts.Labels)
	Registry.MustRegister(histogramVec)
	histogramVecsMap[name] = &HistogramVec{histogramVec, opts}
	return histogramVec
}

// HistogramVecNames returns names of all registered histogramVecs
func HistogramVecNames() []string {
	mutex.RLock()
	defer mutex.RUnlock()
	names := make([]string, len(histogramVecsMap))
	i := 0
	for k := range histogramVecsMap {
		names[i] = k
		i++
	}
	return names
}

// HistogramVecs returns all registered histogramVecs
func HistogramVecs() []*HistogramVec {
	mutex.RLock()
	defer mutex.RUnlock()
	c := make([]*HistogramVec, len(histogramVecsMap))
	i := 0
	for _, v := range histogramVecsMap {
		c[i] = v
		i++
	}
	return c
}

// GetHistogramVec looks up the histogramVec by its fully qualified name
func GetHistogramVec(name string) *HistogramVec {
	mutex.RLock()
	defer mutex.RUnlock()
	return histogramVecsMap[name]
}
