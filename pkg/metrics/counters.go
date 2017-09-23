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

// GetOrMustRegisterCounter first checks if a counter with the same name is already registered.
// If the counter is already registered, and was registered with the same opts, then the cached counter is returned.
// If the counter is already registered, and was registered with the different opts, then a panic is triggered.
// If not such counter exists, then it is registered and cached along with its opts.
func GetOrMustRegisterCounter(opts *prometheus.CounterOpts) prometheus.Counter {
	const FUNC = "GetOrMustRegisterCounter"
	mutex.RLock()
	defer mutex.RUnlock()
	name := CounterFQName(opts)
	if counter := countersMap[name]; counter != nil {
		if CounterOptsMatch(opts, counter.CounterOpts) {
			return counter

		}
		logger.Panic().Str(logging.FUNC, FUNC).
			Str("registered", fmt.Sprintf("%v", counter.CounterOpts)).
			Str("dup", fmt.Sprintf("%v", opts)).
			Err(MetricAlreadyRegisteredWithDifferentOpts).
			Msg("")
	}

	if registered(name) {
		logger.Panic().Str(logging.FUNC, FUNC).
			Str("name", name).
			Int("type", COUNTER.Value()).
			Int("registered_type", COUNTERVEC.Value()).
			Err(MetricNameUsedByDifferentMetricType).
			Msg("")
	}

	counter := prometheus.NewCounter(*opts)
	Registry.MustRegister(counter)
	countersMap[name] = &Counter{counter, opts}
	return counter
}

// CounterNames returns names of all registered counters
func CounterNames() []string {
	mutex.RLock()
	defer mutex.RUnlock()
	names := make([]string, len(countersMap))
	i := 0
	for k := range countersMap {
		names[i] = k
		i++
	}
	return names
}

// Counters returns all registered counters
func Counters() []*Counter {
	c := make([]*Counter, len(countersMap))
	i := 0
	for _, v := range countersMap {
		c[i] = v
		i++
	}
	return c
}

// GetCounter looks up the counter by its fully qualified name
func GetCounter(name string) *Counter {
	return countersMap[name]
}
