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
)

func NewClient(connManager ConnManager) messaging.Client {
	return &client{connManager}
}

type client struct {
	connManager ConnManager
}

// Cluster is the name of the messaging cluster the client connects to
func (a *client) Cluster() messaging.ClusterName {
	return a.connManager.Cluster()
}

// Connect creates a new connection to the cluster.
// Tags are used to help identify how the connection is being used by the app. Tags are currently used to augment logging.
func (a *client) Connect(tags ...string) (messaging.Conn, error) {
	conn, err := a.connManager.Connect(tags...)
	if err != nil {
		return nil, err
	}
	return NewConn(conn)
}

// CloseAllConns closes all connections.
func (a *client) CloseAllConns() {
	a.connManager.CloseAll()
}

// ConnCount returns the number of open connections
func (a *client) ConnCount(tags ...string) int {
	if len(tags) == 0 {
		return a.connManager.ConnCount()
	}

	return len(a.connManager.ManagedConns(tags...))
}

// HealthCheck that sends a round trip message
func (a *client) HealthChecks() []metrics.HealthCheck {
	return a.connManager.HealthChecks()
}

// Metrics returns the metrics that are collected for this client
func (a *client) Metrics() *metrics.MetricOpts {
	return ConnManagerMetrics
}
