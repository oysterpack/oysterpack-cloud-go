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

// Metrics represents service specific metrics.
// It provides a means to discover what metrics have been registered for the service.
type Metrics interface {
	// Counters returns the service counters that have been registered
	Counters() []prometheus.CounterOpts

	// CounterVecs returns the service counter vectors that have been registered
	CounterVecs() []metrics.CounterVecOpts
}

// MetricsRegistry implements the Metrics interface and provides functionality to create and register metrics with prometheus
type MetricsRegistry struct {
	mutex       sync.RWMutex
	counters    []prometheus.CounterOpts
	counterVecs []metrics.CounterVecOpts
}

// Counters returns the service counters that have been registered
func (a *MetricsRegistry) Counters() []prometheus.CounterOpts {
	a.mutex.RLock()
	defer a.mutex.RUnlock()
	return a.counters
}

// CounterVecs returns the service counter vectors that have been registered
func (a *MetricsRegistry) CounterVecs() []metrics.CounterVecOpts {
	a.mutex.RLock()
	defer a.mutex.RUnlock()
	return a.counterVecs
}

// MustRegisterCounter creates and registers a counter based on the supplied opts.
// If registration fails, then the func panic.
func (a *MetricsRegistry) MustRegisterCounter(opts prometheus.CounterOpts) prometheus.Counter {
	a.mutex.Lock()
	defer a.mutex.Unlock()
	counter := prometheus.NewCounter(opts)
	metrics.Registry.MustRegister(counter)
	a.counters = append(a.counters, opts)
	return counter
}

// MustRegisterCounterVec creates and registers a counter vector based on the supplied opts.
// If registration fails, then the func panic.
func (a *MetricsRegistry) MustRegisterCounterVec(opts metrics.CounterVecOpts) *prometheus.CounterVec {
	a.mutex.Lock()
	defer a.mutex.Unlock()
	counter := prometheus.NewCounterVec(opts.CounterOpts, opts.Labels)
	metrics.Registry.MustRegister(counter)
	a.counterVecs = append(a.counterVecs, opts)
	return counter
}
