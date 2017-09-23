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

type MetricType int

const (
	MetricType_UNKNOWN MetricType = iota

	MetricType_COUNTER
	MetricType_GAUGE
	MetricType_HISTOGRAM
	MetricType_SUMMARY

	MetricType_COUNTER_VEC
	MetricType_GAUGE_VEC
	MetricType_HISTOGRAM_VEC
	MetricType_SUMMARY_VEC
)

func (a MetricType) Value() int {
	return int(a)
}

func (a MetricType) String() string {
	switch a {
	case MetricType_COUNTER:
		return "Counter"
	case MetricType_GAUGE:
		return "Gauge"
	case MetricType_HISTOGRAM:
		return "Histogram"
	case MetricType_SUMMARY:
		return "Summary"
	case MetricType_COUNTER_VEC:
		return "CounterVec"
	case MetricType_GAUGE_VEC:
		return "GaugeVec"
	case MetricType_HISTOGRAM_VEC:
		return "HistogramVec"
	case MetricType_SUMMARY_VEC:
		return "SummaryVec"
	default:
		return "UNKNOWN"
	}
}

type Counter struct {
	prometheus.Counter
	*prometheus.CounterOpts
}

func NewCounter(opts prometheus.CounterOpts) *Counter {
	return &Counter{prometheus.NewCounter(opts), &opts}
}

type CounterVec struct {
	*prometheus.CounterVec
	*CounterVecOpts
}

func NewCounterVec(opts CounterVecOpts) *CounterVec {
	return &CounterVec{prometheus.NewCounterVec(opts.CounterOpts, opts.Labels), &opts}
}

type Gauge struct {
	prometheus.Gauge
	*prometheus.GaugeOpts
}

func NewGauge(opts prometheus.GaugeOpts) *Gauge {
	return &Gauge{prometheus.NewGauge(opts), &opts}
}

type GaugeVec struct {
	*prometheus.GaugeVec
	*GaugeVecOpts
}

func NewGaugeVec(opts GaugeVecOpts) *GaugeVec {
	return &GaugeVec{prometheus.NewGaugeVec(opts.GaugeOpts, opts.Labels), &opts}
}

type Histogram struct {
	prometheus.Histogram
	*prometheus.HistogramOpts
}

func NewHistogram(opts prometheus.HistogramOpts) *Histogram {
	return &Histogram{prometheus.NewHistogram(opts), &opts}
}

type HistogramVec struct {
	*prometheus.HistogramVec
	*HistogramVecOpts
}

func NewHistogramVec(opts HistogramVecOpts) *HistogramVec {
	return &HistogramVec{prometheus.NewHistogramVec(opts.HistogramOpts, opts.Labels), &opts}
}

type Summary struct {
	prometheus.Summary
	*prometheus.SummaryOpts
}

func NewSummary(opts prometheus.SummaryOpts) *Summary {
	return &Summary{prometheus.NewSummary(opts), &opts}
}

type SummaryVec struct {
	*prometheus.SummaryVec
	*SummaryVecOpts
}

func NewSummaryVec(opts SummaryVecOpts) *SummaryVec {
	return &SummaryVec{prometheus.NewSummaryVec(opts.SummaryOpts, opts.Labels), &opts}
}
