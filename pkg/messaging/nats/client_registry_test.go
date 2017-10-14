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

package nats_test

import (
	"testing"

	"github.com/oysterpack/oysterpack.go/pkg/messaging"
	"github.com/oysterpack/oysterpack.go/pkg/messaging/nats"
	"github.com/oysterpack/oysterpack.go/pkg/messaging/natstest"
	"github.com/oysterpack/oysterpack.go/pkg/metrics"
	"github.com/oysterpack/oysterpack.go/pkg/service"
)

func TestConnManagerRegistry(t *testing.T) {
	metrics.ResetRegistry()
	defer metrics.ResetRegistry()
	nats.RegisterMetrics()

	serverConfigs := natstest.CreateNATSServerConfigsNoTLS(1)
	servers := natstest.CreateNATSServers(t, serverConfigs)
	natstest.StartServers(servers)
	defer natstest.ShutdownServers(servers)

	app := service.NewApplication(service.ApplicationSettings{})
	app.Start()
	defer app.Stop()

	serviceClient := app.MustRegisterService(nats.NewConnManagerRegistry)
	clientRegistry := serviceClient.(messaging.ClientRegistry)

	connManager := nats.NewConnManager(natstest.ConnManagerSettings(serverConfigs[0]))
	client := nats.NewClient(connManager)
	clientRegistry.MustRegister(client)

	client2 := clientRegistry.Unregister(client.Cluster())
	if client2 == nil {
		t.Error("Client was not returned")
	}

	clientRegistry.MustRegister(client)
	client2 = clientRegistry.Client(client.Cluster())
	if client2 == nil {
		t.Error("Client was not returned")
	}
	clusters := clientRegistry.Clusters()
	if len(clusters) != 1 && clusters[0] != client.Cluster() {
		t.Errorf("cluster names does not match : %v", clusters)
	}

	client.Connect()

	func() {
		defer func() {
			if p := recover(); p == nil {
				t.Error("should have panicked")
			} else {
				t.Log(p)
			}
		}()
		clientRegistry.MustRegister(client)
	}()

}
