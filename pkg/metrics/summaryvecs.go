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

// GetOrMustRegisterSummaryVec first checks if a summaryVec with the same name is already registered.
// If the summaryVec is already registered, and was registered with the same opts, then the cached summaryVec is returned.
// If the summaryVec is already registered, and was registered with the different opts, then a panic is triggered.
// If not such summaryVec exists, then it is registered and cached along with its opts.
func GetOrMustRegisterSummaryVec(opts *SummaryVecOpts) *prometheus.SummaryVec {
	const FUNC = "GetOrMustRegisterSummaryVec"
	mutex.Lock()
	defer mutex.Unlock()
	name := SummaryFQName(opts.SummaryOpts)
	if summaryVec := summaryVecsMap[name]; summaryVec != nil {
		if SummaryVecOptsMatch(opts, summaryVec.SummaryVecOpts) {
			return summaryVec.SummaryVec

		}
		logger.Panic().Str(logging.FUNC, FUNC).
			Str("registered", fmt.Sprintf("%v", summaryVec.SummaryVecOpts)).
			Str("dup", fmt.Sprintf("%v", opts)).
			Err(ErrMetricAlreadyRegisteredWithDifferentOpts).
			Msg("")
	}

	if registered(name) {
		logger.Panic().Str(logging.FUNC, FUNC).
			Str("name", name).
			Int("type", SUMMARYVEC.Value()).
			Int("registered_type", SUMMARY.Value()).
			Err(ErrMetricNameUsedByDifferentMetricType).
			Msg("")
	}

	summaryVec := prometheus.NewSummaryVec(*opts.SummaryOpts, opts.Labels)
	Registry.MustRegister(summaryVec)
	summaryVecsMap[name] = &SummaryVec{summaryVec, opts}
	return summaryVec
}

// SummaryVecNames returns names of all registered summaryVecs
func SummaryVecNames() []string {
	mutex.RLock()
	defer mutex.RUnlock()
	names := make([]string, len(summaryVecsMap))
	i := 0
	for k := range summaryVecsMap {
		names[i] = k
		i++
	}
	return names
}

// SummaryVecs returns all registered summaryVecs
func SummaryVecs() []*SummaryVec {
	mutex.RLock()
	defer mutex.RUnlock()
	c := make([]*SummaryVec, len(summaryVecsMap))
	i := 0
	for _, v := range summaryVecsMap {
		c[i] = v
		i++
	}
	return c
}

// GetSummaryVec looks up the summaryVec by its fully qualified name
func GetSummaryVec(name string) *SummaryVec {
	mutex.RLock()
	defer mutex.RUnlock()
	return summaryVecsMap[name]
}
