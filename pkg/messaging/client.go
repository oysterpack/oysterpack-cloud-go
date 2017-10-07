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

import (
	"github.com/oysterpack/oysterpack.go/pkg/metrics"
)

// Client is the messaging client interface.
// The client is configured to connect to a cluster.
type Client interface {
	// Cluster is the name of the messaging cluster the client connects to
	Cluster() ClusterName

	// Connect creates a new connection to the cluster.
	// Tags are used to help identify how the connection is being used by the app. Tags are currently used to augment logging.
	Connect(tags ...string) (Conn, error)

	// CloseAllConns closes all connections.
	CloseAllConns()

	// ConnCount returns the number of open connections
	ConnCount(tags ...string) int

	// HealthCheck that sends a round trip message
	HealthChecks() []metrics.HealthCheck

	// Metrics returns the metrics that are collected for this client
	Metrics() *metrics.MetricOpts
}
