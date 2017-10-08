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

package messaging

const (
	// MetricsSubSystem is used as the metric subsystem for nats related metrics
	MetricsSubSystem = "messaging"
	// CONST_LABEL_VENDOR is used to specify the messaging vendor, e.g., nats
	CONST_LABEL_VENDOR = "vendor"
)

// MetricLabels is used to set variable labels on metrics.
var MetricLabels = []string{"cluster"}
