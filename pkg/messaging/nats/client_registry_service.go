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
	"github.com/oysterpack/oysterpack.go/pkg/service"
)

const (
	// Namespace name
	Namespace = "oysterpack"
	// System name
	System = "nats"
	// Component name
	Component = "conn_manager_registry"
	// Version is the service version
	Version = "1.0.0"
)

var (
	registry = &clientRegistry{
		registry: make(map[messaging.ClusterName]messaging.Client),
	}

	// ClientRegistryService service singleton
	ClientRegistryService messaging.ClientRegistry = registry

	// ClientRegistryDescriptor service descriptor
	ClientRegistryDescriptor = service.NewDescriptor(Namespace, System, Component, Version, ClientRegistryInterface)

	// ClientRegistryInterface service interface
	ClientRegistryInterface service.Interface = service.ServiceInterface(ClientRegistryService)
)

// NewConnManagerRegistry service.ClientConstructor
func NewConnManagerRegistry(app service.Application) service.Client {
	registry.RestartableService = service.NewRestartableService(registry.newService)
	return registry
}

func (a *clientRegistry) newService() service.Service {
	settings := service.Settings{
		Descriptor: ClientRegistryDescriptor,
		Destroy:    a.Destroy,
		Metrics:    ConnManagerMetrics,
	}
	return service.NewService(settings)
}

func init() {
	service.App().MustRegisterService(NewConnManagerRegistry)
}
