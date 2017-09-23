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

	"sort"

	"github.com/oysterpack/oysterpack.go/pkg/metrics"
	"github.com/prometheus/client_golang/prometheus"
)

func TestNewCounterVecOpts_TrimmingAndSortingLabel(t *testing.T) {
	opts := metrics.NewCounterVecOpts(&prometheus.CounterOpts{Name: " ABC ", Help: " XYZ"}, "  b  ", " a   ")
	if opts.Name != "ABC" {
		t.Errorf("Name have been trimmed, but was %q", opts.Name)
	}
	if opts.Help != "XYZ" {
		t.Errorf("Help should have been trimmed, but was %q", opts.Name)
	}
	if len(opts.Labels) != 2 || opts.Labels[0] != "a" || opts.Labels[1] != "b" {
		t.Errorf("Labels should have been trimmed and sorted :%v", opts.Labels)
	}
}

func TestNewCounterVecOpts_ChecksFail(t *testing.T) {
	func() {
		defer func() {
			if p := recover(); p == nil {
				t.Error("Opts should have failed checks and triggered panic")
			}
		}()
		metrics.NewCounterVecOpts(&prometheus.CounterOpts{}, "  b  ", " a   ")
	}()

	func() {
		defer func() {
			if p := recover(); p == nil {
				t.Error("Opts should have failed checks and triggered panic")
			}
		}()
		metrics.NewCounterVecOpts(&prometheus.CounterOpts{Name: "sfsdf"}, "  b  ", " a   ")
	}()

	func() {
		defer func() {
			if p := recover(); p == nil {
				t.Error("Opts should have failed checks and triggered panic")
			}
		}()
		metrics.NewCounterVecOpts(&prometheus.CounterOpts{Name: "   "}, "  b  ", " a   ")
	}()

	func() {
		defer func() {
			if p := recover(); p == nil {
				t.Error("Opts should have failed checks and triggered panic")
			}
		}()
		metrics.NewCounterVecOpts(&prometheus.CounterOpts{Help: "sdfsdfsdf"}, "  b  ", " a   ")
	}()

	func() {
		defer func() {
			if p := recover(); p == nil {
				t.Error("Opts should have failed checks and triggered panic")
			}
		}()
		metrics.NewCounterVecOpts(&prometheus.CounterOpts{Help: "   "}, "  b  ", " a   ")
	}()

	func() {
		defer func() {
			if p := recover(); p == nil {
				t.Error("Opts should have failed checks and triggered panic")
			}
		}()
		metrics.NewCounterVecOpts(&prometheus.CounterOpts{Name: "a", Help: "a"}, "  ", " a   ")
	}()
}

func TestNewGaugeVecOpts_TrimmingAndSortingLabel(t *testing.T) {
	opts := metrics.NewGaugeVecOpts(&prometheus.GaugeOpts{Name: " ABC ", Help: " XYZ"}, "  b  ", " a   ")
	if opts.Name != "ABC" {
		t.Errorf("Name have been trimmed, but was %q", opts.Name)
	}
	if opts.Help != "XYZ" {
		t.Errorf("Help should have been trimmed, but was %q", opts.Name)
	}
	if len(opts.Labels) != 2 || opts.Labels[0] != "a" || opts.Labels[1] != "b" {
		t.Errorf("Labels should have been trimmed and sorted :%v", opts.Labels)
	}
}

func TestNewGaugeVecOpts_ChecksFail(t *testing.T) {
	func() {
		defer func() {
			if p := recover(); p == nil {
				t.Error("Opts should have failed checks and triggered panic")
			}
		}()
		metrics.NewGaugeVecOpts(&prometheus.GaugeOpts{}, "  b  ", " a   ")
	}()

	func() {
		defer func() {
			if p := recover(); p == nil {
				t.Error("Opts should have failed checks and triggered panic")
			}
		}()
		metrics.NewGaugeVecOpts(&prometheus.GaugeOpts{Name: "sfsdf"}, "  b  ", " a   ")
	}()

	func() {
		defer func() {
			if p := recover(); p == nil {
				t.Error("Opts should have failed checks and triggered panic")
			}
		}()
		metrics.NewGaugeVecOpts(&prometheus.GaugeOpts{Name: "   "}, "  b  ", " a   ")
	}()

	func() {
		defer func() {
			if p := recover(); p == nil {
				t.Error("Opts should have failed checks and triggered panic")
			}
		}()
		metrics.NewGaugeVecOpts(&prometheus.GaugeOpts{Help: "sdfsdfsdf"}, "  b  ", " a   ")
	}()

	func() {
		defer func() {
			if p := recover(); p == nil {
				t.Error("Opts should have failed checks and triggered panic")
			}
		}()
		metrics.NewGaugeVecOpts(&prometheus.GaugeOpts{Help: "   "}, "  b  ", " a   ")
	}()

	func() {
		defer func() {
			if p := recover(); p == nil {
				t.Error("Opts should have failed checks and triggered panic")
			}
		}()
		metrics.NewGaugeVecOpts(&prometheus.GaugeOpts{Name: "a", Help: "a"}, "  ", " a   ")
	}()
}

func TestNewHistogramVecOpts_TrimmingAndSortingLabelAndBucketDedupeSorting(t *testing.T) {
	opts := metrics.NewHistogramVecOpts(&prometheus.HistogramOpts{Name: " ABC ", Help: " XYZ", Buckets: []float64{3, 2, 1, 2}}, "  b  ", " a   ")
	if opts.Name != "ABC" {
		t.Errorf("Name have been trimmed, but was %q", opts.Name)
	}
	if opts.Help != "XYZ" {
		t.Errorf("Help should have been trimmed, but was %q", opts.Name)
	}
	if len(opts.Labels) != 2 || opts.Labels[0] != "a" || opts.Labels[1] != "b" {
		t.Errorf("Labels should have been trimmed and sorted :%v", opts.Labels)
	}

	if len(opts.Buckets) != 3 {
		t.Errorf("buckets should have been deduped : %v", opts.Buckets)
	}
	if !sort.Float64sAreSorted(opts.Buckets) {
		t.Errorf("buckets should have been sorted : %v", opts.Buckets)
	}
}

func TestNewHistogramVecOpts_ChecksFail(t *testing.T) {
	func() {
		defer func() {
			if p := recover(); p == nil {
				t.Error("Opts should have failed checks and triggered panic")
			}
		}()
		metrics.NewHistogramVecOpts(&prometheus.HistogramOpts{}, "  b  ", " a   ")
	}()

	func() {
		defer func() {
			if p := recover(); p == nil {
				t.Error("Opts should have failed checks and triggered panic")
			}
		}()
		metrics.NewHistogramVecOpts(&prometheus.HistogramOpts{Name: "sfsdf"}, "  b  ", " a   ")
	}()

	func() {
		defer func() {
			if p := recover(); p == nil {
				t.Error("Opts should have failed checks and triggered panic")
			}
		}()
		metrics.NewHistogramVecOpts(&prometheus.HistogramOpts{Name: "   "}, "  b  ", " a   ")
	}()

	func() {
		defer func() {
			if p := recover(); p == nil {
				t.Error("Opts should have failed checks and triggered panic")
			}
		}()
		metrics.NewHistogramVecOpts(&prometheus.HistogramOpts{Help: "sdfsdfsdf"}, "  b  ", " a   ")
	}()

	func() {
		defer func() {
			if p := recover(); p == nil {
				t.Error("Opts should have failed checks and triggered panic")
			}
		}()
		metrics.NewHistogramVecOpts(&prometheus.HistogramOpts{Help: "   "}, "  b  ", " a   ")
	}()

	func() {
		defer func() {
			if p := recover(); p == nil {
				t.Error("Opts should have failed checks and triggered panic")
			}
		}()
		metrics.NewHistogramVecOpts(&prometheus.HistogramOpts{Name: "a", Help: "a"}, "  ", " a   ")
	}()
	func() {
		defer func() {
			if p := recover(); p == nil {
				t.Error("Opts should have failed checks and triggered panic")
			}
		}()
		metrics.NewHistogramVecOpts(&prometheus.HistogramOpts{Name: "a", Help: "a", Buckets: []float64{}}, "  ", " a   ")
	}()
}

func TestNewSummaryVecOpts_TrimmingAndSortingLabelAndDefaultObjectives(t *testing.T) {
	opts := metrics.NewSummaryVecOpts(prometheus.SummaryOpts{Name: " ABC ", Help: " XYZ"}, "  b  ", " a   ")
	if opts.Name != "ABC" {
		t.Errorf("Name have been trimmed, but was %q", opts.Name)
	}
	if opts.Help != "XYZ" {
		t.Errorf("Help should have been trimmed, but was %q", opts.Name)
	}
	if len(opts.Labels) != 2 || opts.Labels[0] != "a" || opts.Labels[1] != "b" {
		t.Errorf("Labels should have been trimmed and sorted :%v", opts.Labels)
	}

	if len(opts.Objectives) != 4 {
		t.Errorf("default objectives should have been set")
	}

	metrics.NewSummaryVecOpts(prometheus.SummaryOpts{Name: " ABC ", Help: " XYZ"}, "  b")
}

func TestNewSummaryVecOpts_ChecksFail(t *testing.T) {
	func() {
		defer func() {
			if p := recover(); p == nil {
				t.Error("Opts should have failed checks and triggered panic")
			}
		}()
		metrics.NewSummaryVecOpts(prometheus.SummaryOpts{}, "  b  ")
	}()

	func() {
		defer func() {
			if p := recover(); p == nil {
				t.Error("Opts should have failed checks and triggered panic")
			}
		}()
		metrics.NewSummaryVecOpts(prometheus.SummaryOpts{Name: "sfsdf"}, "  b  ", " a   ")
	}()

	func() {
		defer func() {
			if p := recover(); p == nil {
				t.Error("Opts should have failed checks and triggered panic")
			}
		}()
		metrics.NewSummaryVecOpts(prometheus.SummaryOpts{Name: "   "}, "  b  ", " a   ")
	}()

	func() {
		defer func() {
			if p := recover(); p == nil {
				t.Error("Opts should have failed checks and triggered panic")
			}
		}()
		metrics.NewSummaryVecOpts(prometheus.SummaryOpts{Help: "sdfsdfsdf"}, "  b  ", " a   ")
	}()

	func() {
		defer func() {
			if p := recover(); p == nil {
				t.Error("Opts should have failed checks and triggered panic")
			}
		}()
		metrics.NewSummaryVecOpts(prometheus.SummaryOpts{Help: "   "}, "  b  ", " a   ")
	}()

	func() {
		defer func() {
			if p := recover(); p == nil {
				t.Error("Opts should have failed checks and triggered panic")
			}
		}()
		metrics.NewSummaryVecOpts(prometheus.SummaryOpts{Name: "a", Help: "a"}, "  ", " a   ")
	}()

	func() {
		defer func() {
			if p := recover(); p == nil {
				t.Error("Opts should have failed checks and triggered panic")
			}
		}()
		metrics.NewSummaryVecOpts(prometheus.SummaryOpts{Name: "a", Help: "a"}, " g ", "    ")
	}()
}
