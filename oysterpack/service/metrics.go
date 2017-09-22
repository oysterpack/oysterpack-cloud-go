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

package service

import (
	"sync"

	"github.com/oysterpack/oysterpack.go/oysterpack/metrics"
	"github.com/prometheus/client_golang/prometheus"
)

type Metrics interface {
	Counters() []prometheus.CounterOpts

	CounterVecs() []metrics.CounterVecOpts
}

type MetricsRegistry struct {
	mutex       sync.RWMutex
	counters    []prometheus.CounterOpts
	counterVecs []metrics.CounterVecOpts
}

func (a *MetricsRegistry) Counters() []prometheus.CounterOpts {
	a.mutex.RLock()
	defer a.mutex.RUnlock()
	return a.counters
}

func (a *MetricsRegistry) CounterVecs() []metrics.CounterVecOpts {
	a.mutex.RLock()
	defer a.mutex.RUnlock()
	return a.counterVecs
}

func (a *MetricsRegistry) MustRegisterCounter(opts prometheus.CounterOpts) prometheus.Counter {
	a.mutex.Lock()
	defer a.mutex.Unlock()
	counter := prometheus.NewCounter(opts)
	metrics.Registry.MustRegister(counter)
	a.counters = append(a.counters, opts)
	return counter
}

func (a *MetricsRegistry) MustRegisterCounterVec(opts metrics.CounterVecOpts) *prometheus.CounterVec {
	a.mutex.Lock()
	defer a.mutex.Unlock()
	counter := prometheus.NewCounterVec(opts.CounterOpts, opts.Labels)
	metrics.Registry.MustRegister(counter)
	a.counterVecs = append(a.counterVecs, opts)
	return counter
}
