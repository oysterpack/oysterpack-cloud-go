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
	"github.com/oysterpack/oysterpack.go/pkg/messaging"
	"github.com/oysterpack/oysterpack.go/pkg/metrics"
	"github.com/oysterpack/oysterpack.go/pkg/service"
	"github.com/prometheus/client_golang/prometheus"
)

var (
	// MetricLabels are the variable labels used for all NATS related metrics
	MetricLabels = []string{"cluster"}
	// ConstLabels sets vendor=nats
	ConstLabels = prometheus.Labels{messaging.CONST_LABEL_VENDOR: "nats"}

	// ConnManagerMetrics define the metrics collected per ConnManager
	ConnManagerMetrics = &metrics.MetricOpts{
		CounterVecOpts: []*metrics.CounterVecOpts{
			CreatedCounterOpts,
			ClosedCounterOpts,
			ReconnectedCounterOpts,
			SubscriberErrorCounterOpts,

			TopicMessagesReceivedCounter,
			QueueMessagesReceivedCounter,

			TopicMessagesPublishedCounter,
		},
		GaugeVecOpts: []*metrics.GaugeVecOpts{
			ConnCountOpts,
			NotConnectedCountOpts,

			MsgsInGauge,
			MsgsOutGauge,
			BytesInGauge,
			BytesOutGauge,

			PublisherCount,

			TopicSubscriberCount,
			TopicPendingMessages,
			TopicPendingBytes,
			TopicMaxPendingMessages,
			TopicMaxPendingBytes,
			TopicMessagesDelivered,
			TopicMessagesDropped,

			QueueSubscriberCount,
			QueuePendingMessages,
			QueuePendingBytes,
			QueueMaxPendingMessages,
			QueueMaxPendingBytes,
			QueueMessagesDelivered,
			QueueMessagesDropped,
		},
	}

	// CreatedCounterOpts tracks the number of connections that have been created
	CreatedCounterOpts = &metrics.CounterVecOpts{
		&prometheus.CounterOpts{
			Namespace:   messaging.MetricsNamespace,
			Subsystem:   messaging.MetricsSubSystem,
			Name:        "created",
			Help:        "The number of connections that have been created",
			ConstLabels: service.AddServiceMetricLabels(ConstLabels, ConnManagerRegistryDescriptor),
		},
		MetricLabels,
	}
	createdCounter = metrics.GetOrMustRegisterCounterVec(CreatedCounterOpts)

	// ClosedCounterOpts tracks the number of connection that have been closed
	ClosedCounterOpts = &metrics.CounterVecOpts{
		&prometheus.CounterOpts{
			Namespace:   messaging.MetricsNamespace,
			Subsystem:   messaging.MetricsSubSystem,
			Name:        "closed",
			Help:        "The number of connections that have been closed",
			ConstLabels: service.AddServiceMetricLabels(ConstLabels, ConnManagerRegistryDescriptor),
		},
		MetricLabels,
	}
	closedCounter = metrics.GetOrMustRegisterCounterVec(ClosedCounterOpts)

	// ConnCountOpts tracks the number of current connections
	ConnCountOpts = &metrics.GaugeVecOpts{
		&prometheus.GaugeOpts{
			Namespace:   messaging.MetricsNamespace,
			Subsystem:   messaging.MetricsSubSystem,
			Name:        "count",
			Help:        "The number of active connections",
			ConstLabels: service.AddServiceMetricLabels(ConstLabels, ConnManagerRegistryDescriptor),
		},
		MetricLabels,
	}

	// NotConnectedCountOpts tracks the number of connections that are not currently connected. They may be disconnected, reconnecting, or connecting.
	NotConnectedCountOpts = &metrics.GaugeVecOpts{
		&prometheus.GaugeOpts{
			Namespace:   messaging.MetricsNamespace,
			Subsystem:   messaging.MetricsSubSystem,
			Name:        "not_connected_count",
			Help:        "The number of connections that are not connected",
			ConstLabels: service.AddServiceMetricLabels(ConstLabels, ConnManagerRegistryDescriptor),
		},
		MetricLabels,
	}

	// ReconnectedCounterOpts tracks when connections have reconnected
	ReconnectedCounterOpts = &metrics.CounterVecOpts{
		&prometheus.CounterOpts{
			Namespace:   messaging.MetricsNamespace,
			Subsystem:   messaging.MetricsSubSystem,
			Name:        "reconnects",
			Help:        "The total number of reconnects",
			ConstLabels: service.AddServiceMetricLabels(ConstLabels, ConnManagerRegistryDescriptor),
		},
		MetricLabels,
	}
	reconnectedCounter = metrics.GetOrMustRegisterCounterVec(ReconnectedCounterOpts)

	// SubscriberErrorCounterOpts tracks when subscriber related errors occur
	SubscriberErrorCounterOpts = &metrics.CounterVecOpts{
		&prometheus.CounterOpts{
			Namespace:   messaging.MetricsNamespace,
			Subsystem:   messaging.MetricsSubSystem,
			Name:        "subscriber_errors",
			Help:        "The number of errors encountered while processing inbound messages",
			ConstLabels: service.AddServiceMetricLabels(ConstLabels, ConnManagerRegistryDescriptor),
		},
		MetricLabels,
	}
	errorCounter = metrics.GetOrMustRegisterCounterVec(SubscriberErrorCounterOpts)
)

// These metrics are collected by each ConnManager.
var (
	// MsgsInGauge tracks the total number of messages that have been received on current connections
	MsgsInGauge = &metrics.GaugeVecOpts{
		&prometheus.GaugeOpts{
			Namespace:   messaging.MetricsNamespace,
			Subsystem:   messaging.MetricsSubSystem,
			Name:        "msgs_in",
			Help:        "The number of messages that have been received on all active connections.",
			ConstLabels: service.AddServiceMetricLabels(ConstLabels, ConnManagerRegistryDescriptor),
		},
		MetricLabels,
	}
	// MsgsOutGauge tracks the total number of messages that have been published on current connections
	MsgsOutGauge = &metrics.GaugeVecOpts{
		&prometheus.GaugeOpts{
			Namespace:   messaging.MetricsNamespace,
			Subsystem:   messaging.MetricsSubSystem,
			Name:        "msgs_out",
			Help:        "The number of messages that have been sent on all active connections.",
			ConstLabels: service.AddServiceMetricLabels(ConstLabels, ConnManagerRegistryDescriptor),
		},
		MetricLabels,
	}
	// BytesInGauge tracks the total number of bytes that have been received on current connections
	BytesInGauge = &metrics.GaugeVecOpts{
		&prometheus.GaugeOpts{
			Namespace:   messaging.MetricsNamespace,
			Subsystem:   messaging.MetricsSubSystem,
			Name:        "bytes_in",
			Help:        "The number of bytes that have been received on all active connections.",
			ConstLabels: service.AddServiceMetricLabels(ConstLabels, ConnManagerRegistryDescriptor),
		},
		MetricLabels,
	}
	// BytesOutGauge tracks the total number of bytes that have been published on current connections
	BytesOutGauge = &metrics.GaugeVecOpts{
		&prometheus.GaugeOpts{
			Namespace:   messaging.MetricsNamespace,
			Subsystem:   messaging.MetricsSubSystem,
			Name:        "bytes_out",
			Help:        "The number of bytes that have been sent on all active connections.",
			ConstLabels: service.AddServiceMetricLabels(ConstLabels, ConnManagerRegistryDescriptor),
		},
		MetricLabels,
	}
)

var (
	// TopicMetricLabels are the variable labels for topic subscriptions
	TopicMetricLabels = append(MetricLabels, "topic")
	// QueueMetricLabels are the variable labels for queue subscriptions
	QueueMetricLabels = append(MetricLabels, "topic", "queue")

	// PublisherCount tracks the number of publishers per topic across all active connections.
	PublisherCount = &metrics.GaugeVecOpts{
		&prometheus.GaugeOpts{
			Namespace:   messaging.MetricsNamespace,
			Subsystem:   messaging.MetricsSubSystem,
			Name:        "publisher_count",
			Help:        "The number of publishers per topic across all active connections.",
			ConstLabels: service.AddServiceMetricLabels(ConstLabels, ConnManagerRegistryDescriptor),
		},
		TopicMetricLabels,
	}

	// TopicSubscriberCount tracks the number of subscribers per topic across all active connections
	TopicSubscriberCount = &metrics.GaugeVecOpts{
		&prometheus.GaugeOpts{
			Namespace:   messaging.MetricsNamespace,
			Subsystem:   messaging.MetricsSubSystem,
			Name:        "topic_subscriber_count",
			Help:        "The number of subscribers per topic across all active connections.",
			ConstLabels: service.AddServiceMetricLabels(ConstLabels, ConnManagerRegistryDescriptor),
		},
		TopicMetricLabels,
	}

	// TopicPendingMessages tracks the number of queued messages per topic across all active connections
	TopicPendingMessages = &metrics.GaugeVecOpts{
		&prometheus.GaugeOpts{
			Namespace:   messaging.MetricsNamespace,
			Subsystem:   messaging.MetricsSubSystem,
			Name:        "topic_pending_msgs",
			Help:        "The number of queued messages per topic across all active connections.",
			ConstLabels: service.AddServiceMetricLabels(ConstLabels, ConnManagerRegistryDescriptor),
		},
		TopicMetricLabels,
	}

	// TopicPendingBytes tracks the number of queued message bytes per topic across all active connections
	TopicPendingBytes = &metrics.GaugeVecOpts{
		&prometheus.GaugeOpts{
			Namespace:   messaging.MetricsNamespace,
			Subsystem:   messaging.MetricsSubSystem,
			Name:        "topic_pending_bytes",
			Help:        "The number of queued message bytes per topic across all active connections.",
			ConstLabels: service.AddServiceMetricLabels(ConstLabels, ConnManagerRegistryDescriptor),
		},
		TopicMetricLabels,
	}

	// TopicMaxPendingMessages tracks the number of queued messages per topic across all active connections
	TopicMaxPendingMessages = &metrics.GaugeVecOpts{
		&prometheus.GaugeOpts{
			Namespace:   messaging.MetricsNamespace,
			Subsystem:   messaging.MetricsSubSystem,
			Name:        "topic_pending_msgs_max",
			Help:        "The max number of queued messages per topic across all active connections.",
			ConstLabels: service.AddServiceMetricLabels(ConstLabels, ConnManagerRegistryDescriptor),
		},
		TopicMetricLabels,
	}

	// TopicMaxPendingBytes tracks the number of queued message bytes per topic across all active connections
	TopicMaxPendingBytes = &metrics.GaugeVecOpts{
		&prometheus.GaugeOpts{
			Namespace:   messaging.MetricsNamespace,
			Subsystem:   messaging.MetricsSubSystem,
			Name:        "topic_pending_bytes_max",
			Help:        "The max number of queued message bytes per topic across all active connections.",
			ConstLabels: service.AddServiceMetricLabels(ConstLabels, ConnManagerRegistryDescriptor),
		},
		TopicMetricLabels,
	}

	// TopicMessagesDropped tracks the number of known dropped messages per topic across all active connections
	TopicMessagesDropped = &metrics.GaugeVecOpts{
		&prometheus.GaugeOpts{
			Namespace: messaging.MetricsNamespace,
			Subsystem: messaging.MetricsSubSystem,
			Name:      "topic_msgs_dropped",
			Help: "The number of known dropped messages per topic across all active connections. " +
				"This will correspond to messages dropped by violations of PendingLimits. " +
				"If the server declares the connection a SlowConsumer, this number may not be valid.",
			ConstLabels: service.AddServiceMetricLabels(ConstLabels, ConnManagerRegistryDescriptor),
		},
		TopicMetricLabels,
	}

	// TopicMessagesDelivered tracks the number of messages delivered to subscriptions per topic across all active connections
	TopicMessagesDelivered = &metrics.GaugeVecOpts{
		&prometheus.GaugeOpts{
			Namespace:   messaging.MetricsNamespace,
			Subsystem:   messaging.MetricsSubSystem,
			Name:        "topic_msgs_delivered",
			Help:        "The number of messages delivered to subscriptions per topic across all active connections.",
			ConstLabels: service.AddServiceMetricLabels(ConstLabels, ConnManagerRegistryDescriptor),
		},
		TopicMetricLabels,
	}

	// QueueSubscriberCount tracks the number of subscribers per topic queue across all active connections.
	QueueSubscriberCount = &metrics.GaugeVecOpts{
		&prometheus.GaugeOpts{
			Namespace:   messaging.MetricsNamespace,
			Subsystem:   messaging.MetricsSubSystem,
			Name:        "queue_subscriber_count",
			Help:        "The number of subscribers per topic queue across all active connections.",
			ConstLabels: service.AddServiceMetricLabels(ConstLabels, ConnManagerRegistryDescriptor),
		},
		QueueMetricLabels,
	}

	// QueuePendingMessages tracks the number of queued messages per queue across all active connections
	QueuePendingMessages = &metrics.GaugeVecOpts{
		&prometheus.GaugeOpts{
			Namespace:   messaging.MetricsNamespace,
			Subsystem:   messaging.MetricsSubSystem,
			Name:        "queue_pending_msgs",
			Help:        "The number of queued messages per queue across all active connections.",
			ConstLabels: service.AddServiceMetricLabels(ConstLabels, ConnManagerRegistryDescriptor),
		},
		QueueMetricLabels,
	}

	// QueuePendingBytes tracks the number of queued message bytes per queue across all active connections
	QueuePendingBytes = &metrics.GaugeVecOpts{
		&prometheus.GaugeOpts{
			Namespace:   messaging.MetricsNamespace,
			Subsystem:   messaging.MetricsSubSystem,
			Name:        "queue_pending_bytes",
			Help:        "The number of queued message bytes per queue across all active connections.",
			ConstLabels: service.AddServiceMetricLabels(ConstLabels, ConnManagerRegistryDescriptor),
		},
		QueueMetricLabels,
	}

	// QueueMaxPendingMessages tracks the number of queued messages per queue across all active connections
	QueueMaxPendingMessages = &metrics.GaugeVecOpts{
		&prometheus.GaugeOpts{
			Namespace:   messaging.MetricsNamespace,
			Subsystem:   messaging.MetricsSubSystem,
			Name:        "queue_pending_msgs_max",
			Help:        "The max number of queued messages per queue across all active connections.",
			ConstLabels: service.AddServiceMetricLabels(ConstLabels, ConnManagerRegistryDescriptor),
		},
		QueueMetricLabels,
	}

	// QueueMaxPendingBytes tracks the number of queued message bytes per queue across all active connections
	QueueMaxPendingBytes = &metrics.GaugeVecOpts{
		&prometheus.GaugeOpts{
			Namespace:   messaging.MetricsNamespace,
			Subsystem:   messaging.MetricsSubSystem,
			Name:        "queue_pending_bytes_max",
			Help:        "The max number of queued message bytes per queue across all active connections.",
			ConstLabels: service.AddServiceMetricLabels(ConstLabels, ConnManagerRegistryDescriptor),
		},
		QueueMetricLabels,
	}

	// QueueMessagesDropped tracks the number of known dropped messages per queue across all active connections
	QueueMessagesDropped = &metrics.GaugeVecOpts{
		&prometheus.GaugeOpts{
			Namespace: messaging.MetricsNamespace,
			Subsystem: messaging.MetricsSubSystem,
			Name:      "queue_msgs_dropped",
			Help: "The number of known dropped messages per queue across all active connections. " +
				"This will correspond to messages dropped by violations of PendingLimits. " +
				"If the server declares the connection a SlowConsumer, this number may not be valid.",
			ConstLabels: service.AddServiceMetricLabels(ConstLabels, ConnManagerRegistryDescriptor),
		},
		QueueMetricLabels,
	}

	// QueueMessagesDelivered tracks the number of messages delivered to subscriptions per queue across all active connections
	QueueMessagesDelivered = &metrics.GaugeVecOpts{
		&prometheus.GaugeOpts{
			Namespace:   messaging.MetricsNamespace,
			Subsystem:   messaging.MetricsSubSystem,
			Name:        "queue_msgs_delivered",
			Help:        "The number of messages delivered to subscriptions per queue across all active connections.",
			ConstLabels: service.AddServiceMetricLabels(ConstLabels, ConnManagerRegistryDescriptor),
		},
		QueueMetricLabels,
	}
)

// message counters
var (
	// TopicMessagesReceivedCounter tracks the number of messsages received per topic since the app started
	TopicMessagesReceivedCounter = &metrics.CounterVecOpts{
		&prometheus.CounterOpts{
			Namespace:   messaging.MetricsNamespace,
			Subsystem:   messaging.MetricsSubSystem,
			Name:        "topic_msgs_received",
			Help:        "The number of messages received per topic since the app started.",
			ConstLabels: service.AddServiceMetricLabels(ConstLabels, ConnManagerRegistryDescriptor),
		},
		TopicMetricLabels,
	}
	topicMsgsReceivedCounter = metrics.GetOrMustRegisterCounterVec(TopicMessagesReceivedCounter)

	// QueueMessagesReceivedCounter tracks the number of messages received per topic queue since the app started
	QueueMessagesReceivedCounter = &metrics.CounterVecOpts{
		&prometheus.CounterOpts{
			Namespace:   messaging.MetricsNamespace,
			Subsystem:   messaging.MetricsSubSystem,
			Name:        "queue_msgs_received",
			Help:        "The number of messages received per topic queue since the app started.",
			ConstLabels: service.AddServiceMetricLabels(ConstLabels, ConnManagerRegistryDescriptor),
		},
		QueueMetricLabels,
	}
	queueMsgsReceivedCounter = metrics.GetOrMustRegisterCounterVec(QueueMessagesReceivedCounter)

	// TopicMessagesPublishedCounter tracks the number of messages published per topic since the app started
	TopicMessagesPublishedCounter = &metrics.CounterVecOpts{
		&prometheus.CounterOpts{
			Namespace:   messaging.MetricsNamespace,
			Subsystem:   messaging.MetricsSubSystem,
			Name:        "topic_msgs_published",
			Help:        "The number of messages published per topic since the app started.",
			ConstLabels: service.AddServiceMetricLabels(ConstLabels, ConnManagerRegistryDescriptor),
		},
		TopicMetricLabels,
	}
	topicMsgsPublishedCounter = metrics.GetOrMustRegisterCounterVec(TopicMessagesPublishedCounter)
)

// RegisterMetrics is meant for testing purposes.
// For many service related tests, the globa metrics registry is reset. This is used to re-register the metrics for
// for tests that need to check metrics.
func RegisterMetrics() {
	createdCounter = metrics.GetOrMustRegisterCounterVec(CreatedCounterOpts)
	closedCounter = metrics.GetOrMustRegisterCounterVec(ClosedCounterOpts)

	reconnectedCounter = metrics.GetOrMustRegisterCounterVec(ReconnectedCounterOpts)
	errorCounter = metrics.GetOrMustRegisterCounterVec(SubscriberErrorCounterOpts)

	topicMsgsReceivedCounter = metrics.GetOrMustRegisterCounterVec(TopicMessagesReceivedCounter)
	queueMsgsReceivedCounter = metrics.GetOrMustRegisterCounterVec(QueueMessagesReceivedCounter)
	topicMsgsPublishedCounter = metrics.GetOrMustRegisterCounterVec(TopicMessagesPublishedCounter)
}
