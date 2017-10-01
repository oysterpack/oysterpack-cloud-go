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

// GetOrMustRegisterHistogram first checks if a histogram with the same name is already registered.
// If the histogram is already registered, and was registered with the same opts, then the cached histogram is returned.
// If the histogram is already registered, and was registered with the different opts, then a panic is triggered.
// If not such histogram exists, then it is registered and cached along with its opts.
func GetOrMustRegisterHistogram(opts *prometheus.HistogramOpts) prometheus.Histogram {
	const FUNC = "GetOrMustRegisterHistogram"
	mutex.RLock()
	defer mutex.RUnlock()
	name := HistogramFQName(opts)
	if histogram := histogramsMap[name]; histogram != nil {
		if HistogramOptsMatch(opts, histogram.HistogramOpts) {
			return histogram

		}
		logger.Panic().Str(logging.FUNC, FUNC).
			Str("registered", fmt.Sprintf("%v", histogram.HistogramOpts)).
			Str("dup", fmt.Sprintf("%v", opts)).
			Err(ErrMetricAlreadyRegisteredWithDifferentOpts).
			Msg("")
	}

	if registered(name) {
		logger.Panic().Str(logging.FUNC, FUNC).
			Str("name", name).
			Int("type", HISTOGRAM.Value()).
			Int("registered_type", HISTOGRAMVEC.Value()).
			Err(ErrMetricNameUsedByDifferentMetricType).
			Msg("")
	}

	histogram := prometheus.NewHistogram(*opts)
	Registry.MustRegister(histogram)
	histogramsMap[name] = &Histogram{histogram, opts}
	return histogram
}

// HistogramNames returns names of all registered histograms
func HistogramNames() []string {
	mutex.RLock()
	defer mutex.RUnlock()
	names := make([]string, len(histogramsMap))
	i := 0
	for k := range histogramsMap {
		names[i] = k
		i++
	}
	return names
}

// Histograms returns all registered histograms
func Histograms() []*Histogram {
	mutex.RLock()
	defer mutex.RUnlock()
	c := make([]*Histogram, len(histogramsMap))
	i := 0
	for _, v := range histogramsMap {
		c[i] = v
		i++
	}
	return c
}

// GetHistogram looks up the histogram by its fully qualified name
func GetHistogram(name string) *Histogram {
	mutex.RLock()
	defer mutex.RUnlock()
	return histogramsMap[name]
}
