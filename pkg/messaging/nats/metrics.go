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
	"github.com/oysterpack/oysterpack.go/pkg/metrics"
	"github.com/oysterpack/oysterpack.go/pkg/service"
	"github.com/prometheus/client_golang/prometheus"
)

var (
	// MetricOpts are the metrics registered for the ConnManager service
	MetricOpts = &metrics.MetricOpts{
		CounterOpts: []*prometheus.CounterOpts{
			CreatedCounterOpts,
			DisconnectedCounterOpts,
			ReconnectedCounterOpts,
			SubscriberErrorCounterOpts,
			ClosedCounterOpts,
		},
		GaugeOpts: []*prometheus.GaugeOpts{
			ConnCountOpts,
			MsgsInGauge,
			MsgsOutGauge,
			BytesInGauge,
			BytesOutGauge,
		},
	}

	CreatedCounterOpts = &prometheus.CounterOpts{
		Namespace:   MetricsNamespace,
		Subsystem:   MetricsSubSystem,
		Name:        "created",
		Help:        "The number of connections that have been created",
		ConstLabels: service.AddServiceMetricLabels(prometheus.Labels{}, ConnManagerDescriptor),
	}
	createdCounter = metrics.GetOrMustRegisterCounter(CreatedCounterOpts)

	ClosedCounterOpts = &prometheus.CounterOpts{
		Namespace:   MetricsNamespace,
		Subsystem:   MetricsSubSystem,
		Name:        "closed",
		Help:        "The number of connections that have been closed",
		ConstLabels: service.AddServiceMetricLabels(prometheus.Labels{}, ConnManagerDescriptor),
	}
	closedCounter = metrics.GetOrMustRegisterCounter(ClosedCounterOpts)

	ConnCountOpts = &prometheus.GaugeOpts{
		Namespace:   MetricsNamespace,
		Subsystem:   MetricsSubSystem,
		Name:        "count",
		Help:        "The number of current connections",
		ConstLabels: service.AddServiceMetricLabels(prometheus.Labels{}, ConnManagerDescriptor),
	}
	connCount = metrics.GetOrMustRegisterGauge(ConnCountOpts)

	DisconnectedCounterOpts = &prometheus.CounterOpts{
		Namespace:   MetricsNamespace,
		Subsystem:   MetricsSubSystem,
		Name:        "disconnects",
		Help:        "The number of times connections disconnected",
		ConstLabels: service.AddServiceMetricLabels(prometheus.Labels{}, ConnManagerDescriptor),
	}
	disconnectedCounter = metrics.GetOrMustRegisterCounter(DisconnectedCounterOpts)

	ReconnectedCounterOpts = &prometheus.CounterOpts{
		Namespace:   MetricsNamespace,
		Subsystem:   MetricsSubSystem,
		Name:        "reconnects",
		Help:        "The number of times connections reconnected",
		ConstLabels: service.AddServiceMetricLabels(prometheus.Labels{}, ConnManagerDescriptor),
	}
	reconnectedCounter = metrics.GetOrMustRegisterCounter(ReconnectedCounterOpts)

	SubscriberErrorCounterOpts = &prometheus.CounterOpts{
		Namespace:   MetricsNamespace,
		Subsystem:   MetricsSubSystem,
		Name:        "subscriber_errors",
		Help:        "The number of errors encountered while processing inbound messages.",
		ConstLabels: service.AddServiceMetricLabels(prometheus.Labels{}, ConnManagerDescriptor),
	}
	errorCounter = metrics.GetOrMustRegisterCounter(SubscriberErrorCounterOpts)

	MsgsInGauge = &prometheus.GaugeOpts{
		Namespace:   MetricsNamespace,
		Subsystem:   MetricsSubSystem,
		Name:        "msgs_in",
		Help:        "The number of messages that have been received on all current connections.",
		ConstLabels: service.AddServiceMetricLabels(prometheus.Labels{}, ConnManagerDescriptor),
	}
	MsgsOutGauge = &prometheus.GaugeOpts{
		Namespace:   MetricsNamespace,
		Subsystem:   MetricsSubSystem,
		Name:        "msgs_out",
		Help:        "The number of messages that have been sent on all current connections.",
		ConstLabels: service.AddServiceMetricLabels(prometheus.Labels{}, ConnManagerDescriptor),
	}
	BytesInGauge = &prometheus.GaugeOpts{
		Namespace:   MetricsNamespace,
		Subsystem:   MetricsSubSystem,
		Name:        "bytes_in",
		Help:        "The number of bytes that have been received on all current connections.",
		ConstLabels: service.AddServiceMetricLabels(prometheus.Labels{}, ConnManagerDescriptor),
	}
	BytesOutGauge = &prometheus.GaugeOpts{
		Namespace:   MetricsNamespace,
		Subsystem:   MetricsSubSystem,
		Name:        "bytes_out",
		Help:        "The number of bytes that have been sent on all current connections.",
		ConstLabels: service.AddServiceMetricLabels(prometheus.Labels{}, ConnManagerDescriptor),
	}
)

// RegisterMetrics is meant for testing purposes.
// For many service related tests, the globa metrics registry is reset. This is used to re-register the metrics for
// for tests that need to check metrics.
func RegisterMetrics() {
	createdCounter = metrics.GetOrMustRegisterCounter(CreatedCounterOpts)
	closedCounter = metrics.GetOrMustRegisterCounter(ClosedCounterOpts)
	connCount = metrics.GetOrMustRegisterGauge(ConnCountOpts)
	disconnectedCounter = metrics.GetOrMustRegisterCounter(DisconnectedCounterOpts)
	reconnectedCounter = metrics.GetOrMustRegisterCounter(ReconnectedCounterOpts)
	errorCounter = metrics.GetOrMustRegisterCounter(SubscriberErrorCounterOpts)
}
