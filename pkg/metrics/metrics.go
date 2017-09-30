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
	"github.com/prometheus/client_golang/prometheus"
)

// MetricType is an enum for metric types
type MetricType int

const (
	// UNKNOWN unknown metric type
	UNKNOWN MetricType = iota

	// COUNTER -> prometheus.Counter
	COUNTER
	// GAUGE -> prometheus.Gauge
	GAUGE
	// HISTOGRAM -> prometheus.Histogram
	HISTOGRAM
	// SUMMARY -> prometheus.Summary
	SUMMARY

	// COUNTERVEC -> prometheus.CounterVec
	COUNTERVEC
	// GAUGEVEC -> prometheus.GaugeVec
	GAUGEVEC
	// HISTOGRAMVEC -> prometheus.HistogramVec
	HISTOGRAMVEC
	// SUMMARYVEC -> prometheus.SummaryVec
	SUMMARYVEC
)

// Value returns the value as an int
func (a MetricType) Value() int {
	return int(a)
}

func (a MetricType) String() string {
	switch a {
	case COUNTER:
		return "Counter"
	case GAUGE:
		return "Gauge"
	case HISTOGRAM:
		return "Histogram"
	case SUMMARY:
		return "Summary"
	case COUNTERVEC:
		return "CounterVec"
	case GAUGEVEC:
		return "GaugeVec"
	case HISTOGRAMVEC:
		return "HistogramVec"
	case SUMMARYVEC:
		return "SummaryVec"
	default:
		return "UNKNOWN"
	}
}

// Counter aggregates the metric along with its opts
type Counter struct {
	prometheus.Counter
	*prometheus.CounterOpts
}

// CounterVec aggregates the metric along with its opts
type CounterVec struct {
	*prometheus.CounterVec
	*CounterVecOpts
}

// Gauge aggregates the metric along with its opts
type Gauge struct {
	prometheus.Gauge
	*prometheus.GaugeOpts
}

// GaugeVec aggregates the metric along with its opts
type GaugeVec struct {
	*prometheus.GaugeVec
	*GaugeVecOpts
}

// Gauge aggregates the metric along with its opts
type GaugeFunc struct {
	prometheus.GaugeFunc
	*prometheus.GaugeOpts
}

// Histogram aggregates the metric along with its opts
type Histogram struct {
	prometheus.Histogram
	*prometheus.HistogramOpts
}

// HistogramVec aggregates the metric along with its opts
type HistogramVec struct {
	*prometheus.HistogramVec
	*HistogramVecOpts
}

// Summary aggregates the metric along with its opts
type Summary struct {
	prometheus.Summary
	*prometheus.SummaryOpts
}

// SummaryVec aggregates the metric along with its opts
type SummaryVec struct {
	*prometheus.SummaryVec
	*SummaryVecOpts
}
