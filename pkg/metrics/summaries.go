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

// GetOrMustRegisterSummary first checks if a summary with the same name is already registered.
// If the summary is already registered, and was registered with the same opts, then the cached summary is returned.
// If the summary is already registered, and was registered with the different opts, then a panic is triggered.
// If not such summary exists, then it is registered and cached along with its opts.
func GetOrMustRegisterSummary(opts *prometheus.SummaryOpts) prometheus.Summary {
	const FUNC = "GetOrMustRegisterSummary"
	mutex.RLock()
	defer mutex.RUnlock()
	name := SummaryFQName(opts)
	if summary := summariesMap[name]; summary != nil {
		if SummaryOptsMatch(opts, summary.SummaryOpts) {
			return summary

		}
		logger.Panic().Str(logging.FUNC, FUNC).
			Str("registered", fmt.Sprintf("%v", summary.SummaryOpts)).
			Str("dup", fmt.Sprintf("%v", opts)).
			Err(ErrMetricAlreadyRegisteredWithDifferentOpts).
			Msg("")
	}

	if registered(name) {
		logger.Panic().Str(logging.FUNC, FUNC).
			Str("name", name).
			Int("type", SUMMARY.Value()).
			Int("registered_type", SUMMARYVEC.Value()).
			Err(ErrMetricNameUsedByDifferentMetricType).
			Msg("")
	}

	summary := prometheus.NewSummary(*opts)
	Registry.MustRegister(summary)
	summariesMap[name] = &Summary{summary, opts}
	return summary
}

// SummaryNames returns names of all registered summarys
func SummaryNames() []string {
	mutex.RLock()
	defer mutex.RUnlock()
	names := make([]string, len(summariesMap))
	i := 0
	for k := range summariesMap {
		names[i] = k
		i++
	}
	return names
}

// Summaries returns all registered summarys
func Summaries() []*Summary {
	c := make([]*Summary, len(summariesMap))
	i := 0
	for _, v := range summariesMap {
		c[i] = v
		i++
	}
	return c
}

// GetSummary looks up the summary by its fully qualified name
func GetSummary(name string) *Summary {
	return summariesMap[name]
}
