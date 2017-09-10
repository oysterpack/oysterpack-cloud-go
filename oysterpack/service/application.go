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
	"sync"
	"syscall"
	"time"
)

// App is a global variable
// It is used to register services application wide, i.e., process wide
var App Application = NewApplicationContext()

type Application interface {
	// LookupServiceByType looks up a service via its service interface.
	// If exists is true, then the service was found.
	ServiceByType(serviceInterface commons.InterfaceType) ServiceClient

	ServiceByKey(key ServiceKey) ServiceClient

	Services() []ServiceClient

	// ServiceInterfaces returns the service interfaces for all registered services
	ServiceInterfaces() []commons.InterfaceType

	// ServiceKeys returns the ServiceKey(s) for all registered services
	ServiceKeys() []ServiceKey

	// RegisterService will create a new instance of the service using the supplied service constructor.
	// It will then register the service and start it async.
	// If a service with the same service interface is already registered, then the service will not be started and nill will be returned.
	// The ServiceConstructor is retained until the service is unregistered for the purpose of restarting the service using a new instance.
	RegisterService(newService ServiceConstructor) ServiceClient

	UnRegisterService(service *Service) bool
}

// ServiceClient uses a client-server model.
// The client instance is stable and is a singleton. The backend service instance is not stable - meaning the instance may change.
// For example, the backend service may be restarted, which means a new service instance will be created.
// The client and server instances are decoupled using channels to communicate.
// The client owns the channels.
type ServiceClient interface {
	Service() *Service

	// restarts the service backend
	RestartService()
}

// ServiceConstructor is a service factory function that will construct a new Service instance and return a pointer to it.
// The returned service will be in a New state.
type ServiceConstructor func() ServiceClient

type ServiceKey struct {
	packagePath commons.PackagePath
	typeName    commons.TypeName
}

func (s *ServiceKey) String() string {
	return fmt.Sprintf("%v.%v", s.packagePath, s.typeName)
}

func (k *ServiceKey) Package() commons.PackagePath {
	return k.packagePath
}

func (k *ServiceKey) Type() commons.TypeName {
	return k.typeName
}

type ApplicationServiceState struct {
	ServiceKey
	State
}

func InterfaceTypeToServiceKey(serviceInterface commons.InterfaceType) ServiceKey {
	return ServiceKey{
		packagePath: commons.PackagePath(serviceInterface.PkgPath()),
		typeName:    commons.TypeName(serviceInterface.Name()),
	}
}

type RegisteredService struct {
	NewService ServiceConstructor
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

func (a *ApplicationContext) ServiceByKey(serviceKey ServiceKey) ServiceClient {
	a.mutex.RLock()
	defer a.mutex.RUnlock()
	for _, s := range a.services {
		if InterfaceTypeToServiceKey(s.Service().serviceInterface) == serviceKey {
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

func (a *ApplicationContext) ServiceKeys() []ServiceKey {
	a.mutex.RLock()
	defer a.mutex.RUnlock()
	interfaces := make([]ServiceKey, len(a.services))
	for _, v := range a.ServiceInterfaces() {
		interfaces = append(interfaces, InterfaceTypeToServiceKey(v))
	}
	return interfaces
}

// RegisterService will register the service and start it, if it is not already registered.
// returns the new registered service or nil if a service with the same interface was already registered
func (a *ApplicationContext) RegisterService(newService ServiceConstructor) ServiceClient {
	a.mutex.Lock()
	defer a.mutex.Unlock()
	service := newService()
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
