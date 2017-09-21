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

type Labels sets.Strings

type CounterVecOpts struct {
	prometheus.CounterOpts
	// Labels are the label names for the variable set of labels
	Labels []string
}

// NewCounterVecOpts returns a CounterVecOpts.
// If validation fails, then the func panics. The following checks are applied :
// - labels cannot be empty
// - label names cannot be blank
// - opts.Name vannot be blank
// - opts.Help cannot be blank
//
// All string fields will be trimmed, i.e., opts and labels may be modified.
func NewCounterVecOpts(opts prometheus.CounterOpts, labels Labels) *CounterVecOpts {
	FUNC := "NewCounterVecOpts"

	opts.Name = strings.TrimSpace(opts.Name)
	if len(opts.Name) == 0 {
		logger.Panic().Str(logging.FUNC, FUNC).Msg("name must not be blank")
	}

	opts.Help = strings.TrimSpace(opts.Help)
	if len(opts.Help) == 0 {
		logger.Panic().Str(logging.FUNC, FUNC).Msg("help must not be blank")
	}

	opts.Namespace = strings.TrimSpace(opts.Namespace)
	opts.Subsystem = strings.TrimSpace(opts.Subsystem)

	if labels.Empty() {
		logger.Panic().Str(logging.FUNC, FUNC).Msg("labels must not be empty")
	}
	labelNames := labels.Values()
	for i, label := range labelNames {
		labelNames[i] = strings.TrimSpace(label)
		if len(labelNames[i]) == 0 {
			logger.Panic().Str(logging.FUNC, FUNC).Msg("label names must not be blank")
		}
	}
	sort.Strings(labelNames)
	return &CounterVecOpts{CounterOpts: opts, Labels: labelNames}
}

type GaugeVecOpts struct {
	prometheus.GaugeOpts
	Labels sets.Strings
}
