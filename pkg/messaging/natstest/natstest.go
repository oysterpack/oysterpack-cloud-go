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

package natstest

import (
	"fmt"

	"testing"

	"github.com/oysterpack/oysterpack.go/pkg/messaging"
	"github.com/oysterpack/oysterpack.go/pkg/messaging/nats"
	"github.com/oysterpack/oysterpack.go/pkg/messaging/nats/server"
)

var (
	NATS_SEED_SERVER_URL = fmt.Sprintf("nats://localhost:%d", server.DEFAULT_CLUSTER_PORT)
)

func CreateNATSServers(t *testing.T, configs []*server.NATSServerConfig) []server.NATSServer {
	var servers []server.NATSServer
	for _, config := range configs {
		server, err := server.NewNATSServer(config)
		if err != nil {
			t.Fatalf("server.NewNATSServer failed : %v", err)
		}
		servers = append(servers, server)
	}
	return servers
}

func CreateNATSServerConfigsNoTLS(count int) []*server.NATSServerConfig {
	configs := []*server.NATSServerConfig{}
	for i := 0; i < count; i++ {
		config := &server.NATSServerConfig{
			Cluster:             messaging.ClusterName("osyterpack-test"),
			ServerPort:          server.DEFAULT_SERVER_PORT + i,
			MonitorPort:         server.DEFAULT_MONITOR_PORT + i,
			ClusterPort:         server.DEFAULT_CLUSTER_PORT + i,
			MetricsExporterPort: server.DEFAULT_PROMETHEUS_EXPORTER_HTTP_PORT + i,

			Routes:   defaultRoutesWithSeed(),
			LogLevel: server.DEBUG,
		}
		configs = append(configs, config)
	}
	return configs
}

func defaultRoutesWithSeed(ports ...int) []string {
	routes := []string{NATS_SEED_SERVER_URL}
	for _, port := range ports {
		routes = append(routes, fmt.Sprintf("nats://localhost:%d", port))
	}
	return routes
}

func StartServers(servers []server.NATSServer) {
	for _, server := range servers {
		server.Start()
	}
}

func ShutdownServers(servers []server.NATSServer) {
	for _, server := range servers {
		server.Shutdown()
	}
}

func ConnManagerSettings(config *server.NATSServerConfig) *nats.ConnManagerSettings {
	settings := &nats.ConnManagerSettings{ClusterName: config.Cluster}

	if config.TLSConfig != nil {
		settings.Options = append(settings.Options, nats.ConnectUrl(fmt.Sprintf("tls://localhost:%d", config.ServerPort)))
	} else {
		settings.Options = append(settings.Options, nats.ConnectUrl(fmt.Sprintf("nats://localhost:%d", config.ServerPort)))
	}

	return settings
}