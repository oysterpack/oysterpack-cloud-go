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

import "errors"

var (
	// ErrMetricAlreadyRegisteredWithDifferentOpts indicates a metric opts collision
	ErrMetricAlreadyRegisteredWithDifferentOpts = errors.New("ErrMetricAlreadyRegisteredWithDifferentOpts")
	// ErrMetricNameUsedByDifferentMetricType indicates the metric name collision between different metric types
	ErrMetricNameUsedByDifferentMetricType = errors.New("ErrMetricNameUsedByDifferentMetricType")
	// ErrMetricNameCannotBeBlank metric name is required and cannot be blank
	ErrMetricNameCannotBeBlank = errors.New("ErrMetricNameCannotBeBlank")
	// ErrMetricAlreadyRegistered indicates a metric name collision
	ErrMetricAlreadyRegistered = errors.New("ErrMetricAlreadyRegistered")
)
