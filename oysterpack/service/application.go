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

package service

import (
	"fmt"
	"github.com/oysterpack/oysterpack.go/oysterpack/commons"
	"os"
	"os/signal"
	"reflect"
	"sync"
	"syscall"
	"time"
)

// App is a global variable
// It is used to register services application wide, i.e., process wide
var App Application = NewApplicationContext()

// Application is a service interface. It's functionality is defined by the interfaces it composes.
type Application interface {
	ServiceRegistry
}

// ServiceClient uses a client-server model.
// The client instance is stable and is a singleton. The backend service instance is not stable - meaning the instance may change.
// For example, the backend service may be restarted, which means a new service instance will be created.
// The client and server instances are decoupled using channels to communicate. The client owns the channels.
//
// For restart functionality, RestartableService can be leveraged.
// The ServiceClient must implement the service interface.
type ServiceClient interface {
	Service() *Service

	// restarts the service backend
	RestartService()
}

// ServiceClientConstructor is a service factory function that will construct a new Service instance and return a pointer to it.
// The returned service will be in a New state.
type ServiceClientConstructor func() ServiceClient

type ServiceKey struct {
	commons.PackagePath
	commons.TypeName
}

func (s *ServiceKey) String() string {
	return fmt.Sprintf("%v.%v", s.PackagePath, s.TypeName)
}

type ApplicationServiceState struct {
	ServiceKey
	State
}

func InterfaceTypeToServiceKey(serviceInterface commons.InterfaceType) *ServiceKey {
	return &ServiceKey{
		commons.PackagePath(serviceInterface.PkgPath()),
		commons.TypeName(serviceInterface.Name()),
	}
}

type RegisteredService struct {
	NewService ServiceClientConstructor
	ServiceClient
}

// ApplicationContext is used to manage a set of services
// ApplicationContext itself is a service, by composition.
//
// Features
// 1. Triggering the application to shutdown, triggers each of its registered services to shutdown.
// 2. Triggers shutdown when SIGTERM and SIGINT signals are received
//
// Use NewApplicationContext() to create a new ApplicationContext instance
type ApplicationContext struct {
	mutex sync.RWMutex
	// once a service is stopped, it will be removed from this map
	services map[commons.InterfaceType]*RegisteredService

	// ApplicationContext can be managed itself as a service
	service *Service
}

func (a *ApplicationContext) Service() *Service {
	return a.service
}

// NewApplicationContext returns a new ApplicationContext
func NewApplicationContext() *ApplicationContext {
	app := &ApplicationContext{
		services: make(map[commons.InterfaceType]*RegisteredService),
	}
	var service Application = app
	serviceInterface, _ := commons.ObjectInterface(&service)
	app.service = NewService(serviceInterface, nil, nil, app.destroy)
	return app
}

// ServiceByType looks up a service via its service interface.
// If exists is true, then the service was found.
func (a *ApplicationContext) ServiceByType(serviceInterface commons.InterfaceType) ServiceClient {
	a.mutex.RLock()
	defer a.mutex.RUnlock()
	if s, exists := a.services[serviceInterface]; exists {
		return s
	}
	return nil
}

func (a *ApplicationContext) ServiceByKey(serviceKey *ServiceKey) ServiceClient {
	a.mutex.RLock()
	defer a.mutex.RUnlock()
	for _, s := range a.services {
		if *InterfaceTypeToServiceKey(s.Service().serviceInterface) == *serviceKey {
			return s
		}
	}
	return nil
}

func (a *ApplicationContext) Services() []ServiceClient {
	a.mutex.RLock()
	defer a.mutex.RUnlock()
	var services []ServiceClient
	for _, s := range a.services {
		services = append(services, s)
	}
	return services
}

// ServiceInterfaces returns all service interfaces for all registered services
func (a *ApplicationContext) ServiceInterfaces() []commons.InterfaceType {
	a.mutex.RLock()
	defer a.mutex.RUnlock()
	interfaces := make([]commons.InterfaceType, len(a.services))
	for k := range a.services {
		interfaces = append(interfaces, k)
	}
	return interfaces
}

func (a *ApplicationContext) ServiceKeys() []*ServiceKey {
	a.mutex.RLock()
	defer a.mutex.RUnlock()
	interfaces := make([]*ServiceKey, len(a.services))
	for _, v := range a.ServiceInterfaces() {
		interfaces = append(interfaces, InterfaceTypeToServiceKey(v))
	}
	return interfaces
}

// RegisterService will register the service and start it, if it is not already registered.
// Returns the new registered service or nil if a service with the same interface was already registered.
// A panic occurs if the ServiceClient type is not assignable to the Service.
func (a *ApplicationContext) RegisterService(newService ServiceClientConstructor) ServiceClient {
	a.mutex.Lock()
	defer a.mutex.Unlock()
	service := newService()
	if !reflect.TypeOf(service).AssignableTo(service.Service().serviceInterface) {
		panic(fmt.Sprintf("%T is not assignable to %v", reflect.TypeOf(service), service.Service().serviceInterface))
	}
	if _, exists := a.services[service.Service().serviceInterface]; !exists {
		a.services[service.Service().serviceInterface] = &RegisteredService{NewService: newService, ServiceClient: service}
		service.Service().StartAsync()
		return service
	}
	return nil
}

// UnRegisterService unregisters the specified service.
// The service is simply unregistered, i.e., it is not stopped.
func (a *ApplicationContext) UnRegisterService(service *Service) bool {
	a.mutex.Lock()
	defer a.mutex.Unlock()
	if _, exists := a.services[service.serviceInterface]; exists {
		delete(a.services, service.serviceInterface)
		return true
	}
	return false
}

func (a *ApplicationContext) run(ctx *RunContext) error {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGTERM, syscall.SIGINT)
	for {
		select {
		case <-sigs:
			return nil
		case <-ctx.StopTrigger():
			return nil
		}
	}
}

func (a *ApplicationContext) destroy(ctx *Context) error {
	for _, v := range a.services {
		v.Service().StopAsyc()
	}
	for _, v := range a.services {
		for {
			v.Service().AwaitStopped(5 * time.Second)
			if v.Service().State().Stopped() {
				break
			}
			v.Service().Logger.Warn().Msg("Waiting for service to stop")
		}
	}
	return nil
}
