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

const (
	// MetricsNamespace is used as the metric namespace for nats related metrics
	MetricsNamespace = "nats"
	// MetricsSubSystem is used as the metric subsystem for nats related metrics
	MetricsSubSystem = "conn"
)

var (
	// MetricLabels are the variable metric labels
	MetricLabels = []string{"nats_cluster"}

	// ConnManagerMetrics are the metrics registered for the ConnManager service
	ConnManagerMetrics = &metrics.MetricOpts{
		CounterVecOpts: []*metrics.CounterVecOpts{
			CreatedCounterOpts,
			ClosedCounterOpts,
			DisconnectedCounterOpts,
			ReconnectedCounterOpts,
			SubscriberErrorCounterOpts,
		},
		GaugeVecOpts: []*metrics.GaugeVecOpts{
			ConnCountOpts,
		},
	}

	// CreatedCounterOpts tracks the number of connections that have been created
	CreatedCounterOpts = &metrics.CounterVecOpts{
		&prometheus.CounterOpts{
			Namespace:   MetricsNamespace,
			Subsystem:   MetricsSubSystem,
			Name:        "created",
			Help:        "The number of connections that have been created",
			ConstLabels: service.AddServiceMetricLabels(prometheus.Labels{}, ConnManagerRegistryDescriptor),
		},
		MetricLabels,
	}
	createdCounter = metrics.GetOrMustRegisterCounterVec(CreatedCounterOpts)

	// ClosedCounterOpts tracks the number of connection that have been closed
	ClosedCounterOpts = &metrics.CounterVecOpts{
		&prometheus.CounterOpts{
			Namespace:   MetricsNamespace,
			Subsystem:   MetricsSubSystem,
			Name:        "closed",
			Help:        "The number of connections that have been closed",
			ConstLabels: service.AddServiceMetricLabels(prometheus.Labels{}, ConnManagerRegistryDescriptor),
		},
		MetricLabels,
	}
	closedCounter = metrics.GetOrMustRegisterCounterVec(ClosedCounterOpts)

	// ConnCountOpts tracks the number of current connections
	ConnCountOpts = &metrics.GaugeVecOpts{
		&prometheus.GaugeOpts{
			Namespace:   MetricsNamespace,
			Subsystem:   MetricsSubSystem,
			Name:        "count",
			Help:        "The number of current connections",
			ConstLabels: service.AddServiceMetricLabels(prometheus.Labels{}, ConnManagerRegistryDescriptor),
		},
		MetricLabels,
	}
	connCount = metrics.GetOrMustRegisterGaugeVec(ConnCountOpts)

	// DisconnectedCounterOpts tracks when connections have disconnected
	DisconnectedCounterOpts = &metrics.CounterVecOpts{
		&prometheus.CounterOpts{
			Namespace:   MetricsNamespace,
			Subsystem:   MetricsSubSystem,
			Name:        "disconnects",
			Help:        "The number of times connections disconnected",
			ConstLabels: service.AddServiceMetricLabels(prometheus.Labels{}, ConnManagerRegistryDescriptor),
		},
		MetricLabels,
	}
	disconnectedCounter = metrics.GetOrMustRegisterCounterVec(DisconnectedCounterOpts)

	// ReconnectedCounterOpts tracks when connections have reconnected
	ReconnectedCounterOpts = &metrics.CounterVecOpts{
		&prometheus.CounterOpts{
			Namespace:   MetricsNamespace,
			Subsystem:   MetricsSubSystem,
			Name:        "reconnects",
			Help:        "The number of times connections reconnected",
			ConstLabels: service.AddServiceMetricLabels(prometheus.Labels{}, ConnManagerRegistryDescriptor),
		},
		MetricLabels,
	}
	reconnectedCounter = metrics.GetOrMustRegisterCounterVec(ReconnectedCounterOpts)

	// SubscriberErrorCounterOpts tracks when subscriber related errors occur
	SubscriberErrorCounterOpts = &metrics.CounterVecOpts{
		&prometheus.CounterOpts{
			Namespace:   MetricsNamespace,
			Subsystem:   MetricsSubSystem,
			Name:        "subscriber_errors",
			Help:        "The number of errors encountered while processing inbound messages.",
			ConstLabels: service.AddServiceMetricLabels(prometheus.Labels{}, ConnManagerRegistryDescriptor),
		},
		MetricLabels,
	}
	errorCounter = metrics.GetOrMustRegisterCounterVec(SubscriberErrorCounterOpts)
)

var (

	// MsgsInGauge tracks the total number of messages that have been received on current connections
	MsgsInGauge = &metrics.GaugeVecOpts{
		&prometheus.GaugeOpts{
			Namespace:   MetricsNamespace,
			Subsystem:   MetricsSubSystem,
			Name:        "msgs_in",
			Help:        "The number of messages that have been received on all current connections.",
			ConstLabels: service.AddServiceMetricLabels(prometheus.Labels{}, ConnManagerRegistryDescriptor),
		},
		MetricLabels,
	}
	// MsgsOutGauge tracks the total number of messages that have been published on current connections
	MsgsOutGauge = &metrics.GaugeVecOpts{
		&prometheus.GaugeOpts{
			Namespace:   MetricsNamespace,
			Subsystem:   MetricsSubSystem,
			Name:        "msgs_out",
			Help:        "The number of messages that have been sent on all current connections.",
			ConstLabels: service.AddServiceMetricLabels(prometheus.Labels{}, ConnManagerRegistryDescriptor),
		},
		MetricLabels,
	}
	// BytesInGauge tracks the total number of bytes that have been received on current connections
	BytesInGauge = &metrics.GaugeVecOpts{
		&prometheus.GaugeOpts{
			Namespace:   MetricsNamespace,
			Subsystem:   MetricsSubSystem,
			Name:        "bytes_in",
			Help:        "The number of bytes that have been received on all current connections.",
			ConstLabels: service.AddServiceMetricLabels(prometheus.Labels{}, ConnManagerRegistryDescriptor),
		},
		MetricLabels,
	}
	// BytesOutGauge tracks the total number of bytes that have been published on current connections
	BytesOutGauge = &metrics.GaugeVecOpts{
		&prometheus.GaugeOpts{
			Namespace:   MetricsNamespace,
			Subsystem:   MetricsSubSystem,
			Name:        "bytes_out",
			Help:        "The number of bytes that have been sent on all current connections.",
			ConstLabels: service.AddServiceMetricLabels(prometheus.Labels{}, ConnManagerRegistryDescriptor),
		},
		MetricLabels,
	}
)

// RegisterMetrics is meant for testing purposes.
// For many service related tests, the globa metrics registry is reset. This is used to re-register the metrics for
// for tests that need to check metrics.
func RegisterMetrics() {
	createdCounter = metrics.GetOrMustRegisterCounterVec(CreatedCounterOpts)
	closedCounter = metrics.GetOrMustRegisterCounterVec(ClosedCounterOpts)
	connCount = metrics.GetOrMustRegisterGaugeVec(ConnCountOpts)
	disconnectedCounter = metrics.GetOrMustRegisterCounterVec(DisconnectedCounterOpts)
	reconnectedCounter = metrics.GetOrMustRegisterCounterVec(ReconnectedCounterOpts)
	errorCounter = metrics.GetOrMustRegisterCounterVec(SubscriberErrorCounterOpts)
}
