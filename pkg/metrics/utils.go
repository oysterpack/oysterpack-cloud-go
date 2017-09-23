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
	"sort"

	"github.com/oysterpack/oysterpack.go/pkg/commons/collections"
	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
)

// FindMetricFamilyByName finds a MetricFamily by name.
// nil is returned if no match is found
func FindMetricFamilyByName(gatheredMetrics []*dto.MetricFamily, name string) *dto.MetricFamily {
	for _, m := range gatheredMetrics {
		if *m.Name == name {
			return m
		}
	}
	return nil
}

// CounterFQName returns the fully qualified name for the counter.
func CounterFQName(opts *prometheus.CounterOpts) string {
	o := prometheus.Opts(*opts)
	return MetricFQName(&o)
}

// GaugeFQName returns the fully qualified name for the counter.
func GaugeFQName(opts *prometheus.GaugeOpts) string {
	o := prometheus.Opts(*opts)
	return MetricFQName(&o)
}

func MetricFQName(opts *prometheus.Opts) string {
	return prometheus.BuildFQName(opts.Namespace, opts.Subsystem, opts.Name)
}

// CounterOptsMatch return true if the 2 opts match
func CounterOptsMatch(opts1, opts2 *prometheus.CounterOpts) bool {
	if CounterFQName(opts1) != CounterFQName(opts2) {
		return false
	}

	if opts1.Help != opts2.Help {
		return false
	}

	if !collections.StringMapsAreEqual(opts1.ConstLabels, opts2.ConstLabels) {
		return false
	}

	return true
}

// CounterVecOptsMatch return true if the 2 opts match
func CounterVecOptsMatch(opts1, opts2 *CounterVecOpts) bool {
	if opts1 == nil && opts2 == nil {
		return true
	}
	if opts1 == nil || opts2 == nil {
		return false
	}
	if !CounterOptsMatch(opts1.CounterOpts, opts2.CounterOpts) {
		return false
	}
	sort.Strings(opts1.Labels)
	sort.Strings(opts2.Labels)
	return collections.StringSlicesAreEqual(opts1.Labels, opts2.Labels)
}

// GaugeOptsMatch return true if the 2 opts match
func GaugeOptsMatch(opts1, opts2 *prometheus.GaugeOpts) bool {
	if GaugeFQName(opts1) != GaugeFQName(opts2) {
		return false
	}

	if opts1.Help != opts2.Help {
		return false
	}

	if !collections.StringMapsAreEqual(opts1.ConstLabels, opts2.ConstLabels) {
		return false
	}

	return true
}

// GaugeVecOptsMatch return true if the 2 opts match
func GaugeVecOptsMatch(opts1, opts2 *GaugeVecOpts) bool {
	if opts1 == nil && opts2 == nil {
		return true
	}
	if opts1 == nil || opts2 == nil {
		return false
	}
	if !GaugeOptsMatch(opts1.GaugeOpts, opts2.GaugeOpts) {
		return false
	}
	sort.Strings(opts1.Labels)
	sort.Strings(opts2.Labels)
	return collections.StringSlicesAreEqual(opts1.Labels, opts2.Labels)
}
