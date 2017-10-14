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

import "github.com/prometheus/client_golang/prometheus"

// CounterDesc maps a counter Opts to its Desc
type CounterDesc struct {
	Opts *prometheus.CounterOpts
	Desc *prometheus.Desc
}

// CounterVecDesc maps a counter vector Opts to its Desc
type CounterVecDesc struct {
	Opts *CounterVecOpts
	Desc *prometheus.Desc
}

// GaugeDesc maps a gauge Opts to its Desc
type GaugeDesc struct {
	Opts *prometheus.GaugeOpts
	Desc *prometheus.Desc
}

// GaugeVecDesc maps a gauge vector Opts to its Desc
type GaugeVecDesc struct {
	Opts *GaugeVecOpts
	Desc *prometheus.Desc
}

// HistogramDesc maps a histogram Opts to its Desc
type HistogramDesc struct {
	Opts *prometheus.HistogramOpts
	Desc *prometheus.Desc
}

// HistogramVecDesc maps a histogram vector Opts to its Desc
type HistogramVecDesc struct {
	Opts *HistogramVecOpts
	Desc *prometheus.Desc
}

// SummaryDesc maps a summary Opts to its Desc
type SummaryDesc struct {
	Opts *prometheus.SummaryOpts
	Desc *prometheus.Desc
}

// SummaryVecDesc maps a summary vector Opts to its Desc
type SummaryVecDesc struct {
	Opts *SummaryVecOpts
	Desc *prometheus.Desc
}
