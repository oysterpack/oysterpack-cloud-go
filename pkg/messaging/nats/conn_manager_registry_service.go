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
	"github.com/oysterpack/oysterpack.go/pkg/commons/reflect"
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
	registry = &connManagerRegistry{
		registry: make(map[ClusterName]ConnManager),
	}

	// ConnManagerRegistryService service singleton
	ConnManagerRegistryService ConnManagerRegistry = registry

	// ConnManagerRegistryDescriptor service descriptor
	ConnManagerRegistryDescriptor = service.NewDescriptor(Namespace, System, Component, Version, ConnManagerRegistryInterface)

	// ConnManagerRegistryInterface service interface
	ConnManagerRegistryInterface service.Interface = func() service.Interface {
		serviceInterface, err := reflect.ObjectInterface(&ConnManagerRegistryService)
		if err != nil {
			panic(err)
		}
		return serviceInterface
	}()
)

// NewConnManagerRegistryClient service.ClientConstructor
func NewConnManagerRegistryClient(app service.Application) service.Client {
	registry.RestartableService = service.NewRestartableService(registry.newService)
	return registry
}

func (a *connManagerRegistry) newService() service.Service {
	settings := service.Settings{
		Descriptor: ConnManagerRegistryDescriptor,
		Destroy: func(ctx *service.Context) error {
			a.CloseAll()
			return nil
		},
		Metrics: ConnManagerMetrics,
	}
	return service.NewService(settings)
}

func init() {
	service.App().MustRegisterService(NewConnManagerRegistryClient)
}
