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
	"strings"

	"github.com/oysterpack/oysterpack.go/oysterpack/commons/collections/sets"
	"github.com/oysterpack/oysterpack.go/oysterpack/logging"
	"github.com/prometheus/client_golang/prometheus"
)

// CounterVecOpts represents the settings for a prometheus counter vector metric
type CounterVecOpts struct {
	prometheus.CounterOpts

	Labels []string
}

// CheckCounterOpts panics if checks fail :
// - name must not be blank
// - help must not be blank
// The returned opts has all string fields trimmed.
func CheckCounterOpts(opts prometheus.CounterOpts) prometheus.CounterOpts {
	FUNC := "CheckCounterOpts"
	opts.Name = strings.TrimSpace(opts.Name)
	mustNotBeBlank(opts.Name, FUNC, "Name")

	opts.Help = strings.TrimSpace(opts.Help)
	mustNotBeBlank(opts.Help, FUNC, "Help")

	opts.Namespace = strings.TrimSpace(opts.Namespace)
	opts.Subsystem = strings.TrimSpace(opts.Subsystem)

	return opts
}

// NewCounterVecOpts returns a new CounterVecOpts.
// If validation fails, then the func panics. The following checks are applied :
// - label names cannot be blank
// - opts.Name vannot be blank
// - opts.Help cannot be blank
//
// All string fields will be trimmed, i.e., opts and labels may be modified.
func NewCounterVecOpts(opts prometheus.CounterOpts, label string, labels ...string) *CounterVecOpts {
	return &CounterVecOpts{CounterOpts: CheckCounterOpts(opts), Labels: labelNames(label, labels...)}
}

// GaugeVecOpts represents the settings for a prometheus gauge vector metric
type GaugeVecOpts struct {
	prometheus.GaugeOpts
	Labels []string
}

// CheckGaugeOpts panics if checks fail :
// - name must not be blank
// - help must not be blank
// The returned opts has all string fields trimmed.
func CheckGaugeOpts(opts prometheus.GaugeOpts) prometheus.GaugeOpts {
	FUNC := "CheckGaugeOpts"

	opts.Name = strings.TrimSpace(opts.Name)
	mustNotBeBlank(opts.Name, FUNC, "Name")

	opts.Help = strings.TrimSpace(opts.Help)
	mustNotBeBlank(opts.Help, FUNC, "Help")

	opts.Namespace = strings.TrimSpace(opts.Namespace)
	opts.Subsystem = strings.TrimSpace(opts.Subsystem)

	return opts
}

// NewGaugeVecOpts returns a new GaugeVecOpts.
// If validation fails, then the func panics. The following checks are applied :
// - labels cannot be empty
// - label names cannot be blank
// - opts.Name vannot be blank
// - opts.Help cannot be blank
//
// All string fields will be trimmed, i.e., opts and labels may be modified.
func NewGaugeVecOpts(opts prometheus.GaugeOpts, label string, labels ...string) *GaugeVecOpts {
	return &GaugeVecOpts{GaugeOpts: CheckGaugeOpts(opts), Labels: labelNames(label, labels...)}
}

// HistogramVecOpts represents the settings for a prometheus histogram vector metric
type HistogramVecOpts struct {
	prometheus.HistogramOpts
	Labels []string
}

// CheckGaugeOpts panics if checks fail :
// - name must not be blank
// - help must not be blank
// - buckets are required
// The returned opts has all string fields trimmed and the buckets are deduped and sorted.
func CheckHistogramOpts(opts prometheus.HistogramOpts) prometheus.HistogramOpts {
	FUNC := "CheckHistogramOpts"

	if len(opts.Buckets) == 0 {
		logger.Panic().Str(logging.FUNC, FUNC).Msg("buckets are required")
	}

	opts.Name = strings.TrimSpace(opts.Name)
	mustNotBeBlank(opts.Name, FUNC, "Name")

	opts.Help = strings.TrimSpace(opts.Help)
	mustNotBeBlank(opts.Help, FUNC, "Help")

	opts.Namespace = strings.TrimSpace(opts.Namespace)
	opts.Subsystem = strings.TrimSpace(opts.Subsystem)

	// dedupe and sort
	bucketSet := map[float64]struct{}{}
	empty := struct{}{}
	for _, v := range opts.Buckets {
		bucketSet[v] = empty
	}
	opts.Buckets = make([]float64, len(bucketSet))
	i := 0
	for k, _ := range bucketSet {
		opts.Buckets[i] = k
		i++
	}
	sort.Float64s(opts.Buckets)

	return opts
}

// NewHistogramVecOpts returns a new HistogramVecOpts.
// If validation fails, then the func panics. The following checks are applied :
// - labels cannot be empty
// - label names cannot be blank
// - opts.Name vannot be blank
// - opts.Help cannot be blank
//
// All string fields will be trimmed, i.e., opts and labels may be modified.
// Buckets will be deduped and sorted.
func NewHistogramVecOpts(opts prometheus.HistogramOpts, label string, labels ...string) *HistogramVecOpts {
	return &HistogramVecOpts{HistogramOpts: CheckHistogramOpts(opts), Labels: labelNames(label, labels...)}
}

// SummaryVecOpts represents the settings for a prometheus summary vector metric
type SummaryVecOpts struct {
	prometheus.SummaryOpts
	Labels []string
}

// DefObjectives are the default summary objectives used by CheckSummaryOpts() if none are specified.
var DefObjectives = map[float64]float64{0.5: 0.05, 0.9: 0.01, 0.95: 0.001, 0.99: 0.001}

// NewHistogramVecOpts returns a new HistogramVecOpts.
// If validation fails, then the func panics. The following checks are applied :
// - labels cannot be empty
// - label names cannot be blank
// - opts.Name vannot be blank
// - opts.Help cannot be blank
//
// All string fields will be trimmed, i.e., opts and labels may be modified.
// If there were no objectives set, then default objectives for the 50th, 90th, 95th, and 99th percentiles are set
func CheckSummaryOpts(opts prometheus.SummaryOpts) prometheus.SummaryOpts {
	FUNC := "CheckSummaryOpts"

	opts.Name = strings.TrimSpace(opts.Name)
	mustNotBeBlank(opts.Name, FUNC, "Name")

	opts.Help = strings.TrimSpace(opts.Help)
	mustNotBeBlank(opts.Help, FUNC, "Help")

	opts.Namespace = strings.TrimSpace(opts.Namespace)
	opts.Subsystem = strings.TrimSpace(opts.Subsystem)

	if len(opts.Objectives) == 0 {
		opts.Objectives = DefObjectives
	}

	return opts
}

// NewSummaryVecOpts returns a v.
// If validation fails, then the func panics. The following checks are applied :
// - labels cannot be empty
// - label names cannot be blank
// - opts.Name vannot be blank
// - opts.Help cannot be blank
//
// All string fields will be trimmed, i.e., opts and labels may be modified.
// If there were no objectives set, then default objectives for the 50th, 90th, 95th, and 99th percentiles are set.
func NewSummaryVecOpts(opts prometheus.SummaryOpts, label string, labels ...string) *SummaryVecOpts {
	return &SummaryVecOpts{SummaryOpts: CheckSummaryOpts(opts), Labels: labelNames(label, labels...)}
}

// LabelSet is an alias for a set of strings that represent a set of labels
type LabelSet sets.Strings

// labels converts the LabelSet into a []string.
// The label names are sorted lexographically.
// The labels are validated and panic if :
// - labels are empty
// - any og the label names is blank
func labelNames(label string, labels ...string) []string {
	FUNC := "labelNames"
	label = strings.TrimSpace(label)
	if len(label) == 0 {
		logger.Panic().Str(logging.FUNC, FUNC).Msg("label name must not be blank")
	}

	if len(labels) == 0 {
		return []string{label}
	}

	labelSet := sets.NewStrings()
	labelSet.Add(label)
	for _, label := range labels {
		label = strings.TrimSpace(label)
		if len(label) == 0 {
			logger.Panic().Str(logging.FUNC, FUNC).Msg("label name must not be blank")
		}
		labelSet.Add(label)
	}

	names := labelSet.Values()
	sort.Strings(names)
	return names
}

// panics if s is blank
// funcName and field name are used to construct the panic log message where
// funcName is the name of the calling function and fieldName is the name of the field that is being checked
func mustNotBeBlank(s string, funcName string, fieldName string) {
	if len(s) == 0 {
		logger.Panic().Str(logging.FUNC, funcName).Msgf("%s must not be blank", fieldName)
	}
}
