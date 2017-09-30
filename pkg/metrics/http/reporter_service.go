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

package http

import (
	"github.com/oysterpack/oysterpack.go/pkg/commons/reflect"

	"github.com/oysterpack/oysterpack.go/pkg/service"
	"github.com/prometheus/client_golang/prometheus"
)

const (
	// Namespace name
	Namespace = "oysterpack"
	// System name
	System = "metrics"
	// Component name
	Component = "prometheus_http"
	// Version is the service version
	Version = "1.0.0"
)

// Reporter interface
type Reporter interface {
	// Registry is the registry that is used to report metrics
	Registry() *prometheus.Registry
}

// ReporterInterface is the service interface instance that can be used to lookup the registered service
var ReporterInterface service.Interface = func() service.Interface {
	var c Reporter = &reporter{}
	serviceInterface, err := reflect.ObjectInterface(&c)
	if err != nil {
		panic(err)
	}
	return serviceInterface
}()

// NewReporterClient service.ClientConstructor
func NewReporterClient(app service.Application) service.Client {
	c := &reporter{}
	c.RestartableService = service.NewRestartableService(c.newService)
	return c
}

func (a *reporter) newService() service.Service {
	settings := service.Settings{
		Descriptor: service.NewDescriptor(Namespace, System, Component, Version, ReporterInterface),
		Init:       a.init,
		Destroy:    a.destroy,
	}
	return service.NewService(settings)
}

func init() {
	// auto-register the service with the app
	service.App().MustRegisterService(NewReporterClient)
}
