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
	System = "messaging"
	// Component name
	Component = "nats_conn_manager"
	// Version is the service version
	Version = "1.0.0"
)

var ConnectionManagerInterface service.Interface = func() service.Interface {
	var c ConnManager = &connManager{}
	serviceInterface, err := reflect.ObjectInterface(&c)
	if err != nil {
		panic(err)
	}
	return serviceInterface
}()

func (a *connManager) newService() service.Service {
	settings := service.Settings{
		Descriptor: service.NewDescriptor(Namespace, System, Component, Version, ConnectionManagerInterface),
		Init: func(ctx *service.Context) error {
			a.init()
			return nil
		},
		Destroy: func(ctx *service.Context) error {
			a.CloseAll()
			return nil
		},
	}
	return service.NewService(settings)
}

// NewConnManagerClient service.ClientConstructor
func NewConnManagerClient(app service.Application) service.Client {
	c := &connManager{}
	c.RestartableService = service.NewRestartableService(c.newService)
	return c
}
