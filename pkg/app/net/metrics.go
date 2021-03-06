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

package net

import "github.com/oysterpack/oysterpack.go/pkg/app"

// server metrics
const (
	// gauges

	// the current number of connections
	SERVER_CONN_COUNT_METRIC_ID = app.MetricID(0xaa8b66726c359b98)

	// counters

	// the total number of connections created since the server started
	SERVER_CONN_TOTAL_CREATED_METRIC_ID   = app.MetricID(0xc33ede6c39ca8a07)
	SERVER_REQUEST_COUNT_METRIC_ID        = app.MetricID(0xc33ede6c39ca8a07)
	SERVER_REQUEST_FAILED_COUNT_METRIC_ID = app.MetricID(0xc33ede6c39ca8a07)
)
