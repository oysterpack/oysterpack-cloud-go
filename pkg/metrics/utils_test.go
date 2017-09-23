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

package metrics_test

import (
	"testing"

	"time"

	"github.com/oysterpack/oysterpack.go/pkg/metrics"
	"github.com/prometheus/client_golang/prometheus"
)

func TestCounterFQName(t *testing.T) {
	opts := &prometheus.CounterOpts{
		Namespace:   "Namespace",
		Subsystem:   "Subsystem",
		Name:        "Name",
		ConstLabels: map[string]string{"a": "b"},
	}

	expected := "Namespace_Subsystem_Name"
	if actual := metrics.CounterFQName(opts); actual != expected {
		t.Errorf("%v != %v", actual, expected)
	}
}

func TestCounterOptsMatch(t *testing.T) {
	opts1 := &prometheus.CounterOpts{
		Namespace:   "Namespace",
		Subsystem:   "Subsystem",
		Name:        "Name",
		ConstLabels: map[string]string{"a": "b"},
	}

	opts2 := &prometheus.CounterOpts{
		Namespace:   "Namespace",
		Subsystem:   "Subsystem",
		Name:        "Name",
		ConstLabels: map[string]string{"a": "b"},
	}

	if !metrics.CounterOptsMatch(opts1, opts2) {
		t.Error("opts should have matched")
	}
	opts1.Namespace = ""
	if metrics.CounterOptsMatch(opts1, opts2) {
		t.Error("opts should not have matched")
	}
	opts1.Namespace = "Namespace"
	opts1.Subsystem = ""
	if metrics.CounterOptsMatch(opts1, opts2) {
		t.Error("opts should not have matched")
	}
	opts1.Namespace = "Namespace"
	opts1.Subsystem = "Subsystem"
	opts1.Name = ""
	if metrics.CounterOptsMatch(opts1, opts2) {
		t.Error("opts should not have matched")
	}

	opts1.Namespace = "Namespace"
	opts1.Subsystem = "Subsystem"
	opts1.Name = "Name"
	opts1.ConstLabels = nil
	if metrics.CounterOptsMatch(opts1, opts2) {
		t.Error("opts should not have matched")
	}

	opts1.Namespace = "Namespace"
	opts1.Subsystem = "Subsystem"
	opts1.Name = "Name"
	opts1.ConstLabels = opts2.ConstLabels
	opts1.Help = "HELP"
	if metrics.CounterOptsMatch(opts1, opts2) {
		t.Error("opts should not have matched")
	}
}

func TestHistogramOptsMatch(t *testing.T) {
	opts1 := &prometheus.HistogramOpts{
		Name: "Name",
	}
	opts2 := &prometheus.HistogramOpts{
		Name: "Name",
	}
	if !metrics.HistogramOptsMatch(opts1, opts2) {
		t.Error("opts should have matched")
	}

	opts1.Name = "AAA"
	if metrics.HistogramOptsMatch(opts1, opts2) {
		t.Error("opts should not have matched")
	}

	opts2.Name = "AAA"
	opts2.Help = "AAA"
	if metrics.HistogramOptsMatch(opts1, opts2) {
		t.Error("opts should not have matched")
	}

	opts1.Buckets = []float64{1, 2, 3}
	opts1.Help = "AAA"
	if metrics.HistogramOptsMatch(opts1, opts2) {
		t.Error("opts should not have matched")
	}

	opts2.Buckets = []float64{1, 2, 3}
	if !metrics.HistogramOptsMatch(opts1, opts2) {
		t.Error("opts should have matched")
	}

	opts2.Buckets = []float64{3, 2, 1}
	if !metrics.HistogramOptsMatch(opts1, opts2) {
		t.Error("opts should have matched")
	}

}

func TestCounterVecOptsMatch(t *testing.T) {
	var opts1, opts2 *metrics.CounterVecOpts
	if !metrics.CounterVecOptsMatch(opts1, opts2) {
		t.Error("opts should have matched")
	}

	opts1 = &metrics.CounterVecOpts{
		CounterOpts: &prometheus.CounterOpts{
			Namespace:   "Namespace",
			Subsystem:   "Subsystem",
			Name:        "Name",
			ConstLabels: map[string]string{"a": "b"},
		},
	}

	if metrics.CounterVecOptsMatch(opts1, opts2) {
		t.Error("opts should not have matched")
	}

	opts2 = &metrics.CounterVecOpts{
		CounterOpts: &prometheus.CounterOpts{
			Namespace: "Namespace",
			Subsystem: "Subsystem",
			Name:      "Name",
		},
	}

	if metrics.CounterVecOptsMatch(opts1, opts2) {
		t.Error("opts should not have matched")
	}
}

func TestGaugeOptsMatch(t *testing.T) {
	opts1 := &prometheus.GaugeOpts{
		Name: "Name",
	}
	opts2 := &prometheus.GaugeOpts{
		Name: "Name",
	}
	if !metrics.GaugeOptsMatch(opts1, opts2) {
		t.Error("opts should have matched")
	}

	opts1.Name = "AAA"
	if metrics.GaugeOptsMatch(opts1, opts2) {
		t.Error("opts should not have matched")
	}

	opts2.Name = "AAA"
	opts2.Help = "AAA"
	if metrics.GaugeOptsMatch(opts1, opts2) {
		t.Error("opts should not have matched")
	}

}

func TestGaugeVecOptsMatch(t *testing.T) {
	var opts1, opts2 *metrics.GaugeVecOpts
	if !metrics.GaugeVecOptsMatch(opts1, opts2) {
		t.Error("opts should have matched")
	}

	opts1 = &metrics.GaugeVecOpts{
		GaugeOpts: &prometheus.GaugeOpts{
			Namespace:   "Namespace",
			Subsystem:   "Subsystem",
			Name:        "Name",
			ConstLabels: map[string]string{"a": "b"},
		},
	}

	if metrics.GaugeVecOptsMatch(opts1, opts2) {
		t.Error("opts should not have matched")
	}

	opts2 = &metrics.GaugeVecOpts{
		GaugeOpts: &prometheus.GaugeOpts{
			Namespace: "Namespace",
			Subsystem: "Subsystem",
			Name:      "Name",
		},
	}

	if metrics.GaugeVecOptsMatch(opts1, opts2) {
		t.Error("opts should not have matched")
	}
}

func TestSummaryOptsMatch(t *testing.T) {
	opts1 := &prometheus.SummaryOpts{
		Name: "Name",
	}
	opts2 := &prometheus.SummaryOpts{
		Name: "Name",
	}
	if !metrics.SummaryOptsMatch(opts1, opts2) {
		t.Error("opts should have matched")
	}

	opts1.Name = "AAA"
	if metrics.SummaryOptsMatch(opts1, opts2) {
		t.Error("opts should not have matched")
	}

	opts2.Name = "AAA"
	opts2.Help = "AAA"
	if metrics.SummaryOptsMatch(opts1, opts2) {
		t.Error("opts should not have matched")
	}

	opts1.Help = "AAA"
	opts1.MaxAge = time.Minute * 15
	if metrics.SummaryOptsMatch(opts1, opts2) {
		t.Error("opts should not have matched")
	}

	opts2.MaxAge = time.Minute * 15
	opts1.BufCap = 10
	if metrics.SummaryOptsMatch(opts1, opts2) {
		t.Error("opts should not have matched")
	}

	opts1.AgeBuckets = 1000
	opts2.BufCap = 10
	if metrics.SummaryOptsMatch(opts1, opts2) {
		t.Error("opts should not have matched")
	}

	opts2.AgeBuckets = 1000
	opts1.Objectives = map[float64]float64{0.9: 0.01}
	if metrics.SummaryOptsMatch(opts1, opts2) {
		t.Error("opts should not have matched")
	}

	opts2.Objectives = map[float64]float64{0.9: 0.01}
	opts1.Objectives = map[float64]float64{0.9: 0.01}
	if !metrics.SummaryOptsMatch(opts1, opts2) {
		t.Error("opts should have matched")
	}

}

func TestSummaryVecOptsMatch(t *testing.T) {
	var opts1, opts2 *metrics.SummaryVecOpts
	if !metrics.SummaryVecOptsMatch(opts1, opts2) {
		t.Error("opts should have matched")
	}

	opts1 = &metrics.SummaryVecOpts{
		SummaryOpts: &prometheus.SummaryOpts{
			Name: "Name",
		},
	}

	if metrics.SummaryVecOptsMatch(opts1, opts2) {
		t.Error("opts should not have matched")
	}

	opts2 = &metrics.SummaryVecOpts{
		SummaryOpts: &prometheus.SummaryOpts{
			Name: "Name",
			Help: "AAAA",
		},
	}

	if metrics.SummaryVecOptsMatch(opts1, opts2) {
		t.Error("opts should not have matched")
	}
}
