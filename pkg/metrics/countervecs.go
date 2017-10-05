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

// GetOrMustRegisterCounterVec first checks if a counterVec with the same name is already registered.
// If the counterVec is already registered, and was registered with the same opts, then the cached counterVec is returned.
// If the counterVec is already registered, and was registered with the different opts, then a panic is triggered.
// If not such counterVec exists, then it is registered and cached along with its opts.
func GetOrMustRegisterCounterVec(opts *CounterVecOpts) *prometheus.CounterVec {
	const FUNC = "GetOrMustRegisterCounterVec"
	mutex.Lock()
	defer mutex.Unlock()
	name := CounterFQName(opts.CounterOpts)
	if counterVec := counterVecsMap[name]; counterVec != nil {
		if CounterVecOptsMatch(opts, counterVec.CounterVecOpts) {
			return counterVec.CounterVec

		}
		logger.Panic().Str(logging.FUNC, FUNC).
			Str("registered", fmt.Sprintf("%v", counterVec.CounterVecOpts)).
			Str("dup", fmt.Sprintf("%v", opts)).
			Err(ErrMetricAlreadyRegisteredWithDifferentOpts).
			Msg("")
	}

	if registered(name) {
		logger.Panic().Str(logging.FUNC, FUNC).
			Str("name", name).
			Int("type", COUNTERVEC.Value()).
			Int("registered_type", COUNTER.Value()).
			Err(ErrMetricNameUsedByDifferentMetricType).
			Msg("")
	}

	counterVec := prometheus.NewCounterVec(*opts.CounterOpts, opts.Labels)
	Registry.MustRegister(counterVec)
	counterVecsMap[name] = &CounterVec{counterVec, opts}
	return counterVec
}

// CounterVecNames returns names of all registered counterVecs
func CounterVecNames() []string {
	mutex.RLock()
	defer mutex.RUnlock()
	names := make([]string, len(counterVecsMap))
	i := 0
	for k := range counterVecsMap {
		names[i] = k
		i++
	}
	return names
}

// CounterVecs returns all registered counterVecs
func CounterVecs() []*CounterVec {
	mutex.RLock()
	defer mutex.RUnlock()
	c := make([]*CounterVec, len(counterVecsMap))
	i := 0
	for _, v := range counterVecsMap {
		c[i] = v
		i++
	}
	return c
}

// GetCounterVec looks up the counterVec by its fully qualified name
func GetCounterVec(name string) *CounterVec {
	mutex.RLock()
	defer mutex.RUnlock()
	return counterVecsMap[name]
}
