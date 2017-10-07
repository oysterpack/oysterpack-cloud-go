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

package nats

import (
	"time"

	"github.com/oysterpack/oysterpack.go/pkg/messaging"
	"github.com/oysterpack/oysterpack.go/pkg/metrics"
	"github.com/oysterpack/oysterpack.go/pkg/service"
	"github.com/prometheus/client_golang/prometheus"
)

var connectivityHealthCheck = &metrics.GaugeVecOpts{
	&prometheus.GaugeOpts{
		Namespace:   messaging.MetricsNamespace,
		Subsystem:   messaging.MetricsSubSystem,
		Name:        "connectivity",
		Help:        "The healthcheck fails if any connections are disconnected.",
		ConstLabels: service.AddServiceMetricLabels(ConstLabels, ConnManagerRegistryDescriptor),
	},
	MetricLabels,
}

const runinterval = 15 * time.Second
