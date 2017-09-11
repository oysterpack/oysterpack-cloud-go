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

import "github.com/oysterpack/oysterpack.go/oysterpack/commons"

// ServiceRegistry is a service registry.
// A Service register itself as a ServiceClient - via a ServiceClientConstructor.
type ServiceRegistry interface {
	// ServiceByType looks up a service and returns nil if the service is not founc.
	ServiceByType(serviceInterface commons.InterfaceType) ServiceClient

	// ServiceByKey looks up a service and returns nil if the service is not founc.
	ServiceByKey(key *ServiceKey) ServiceClient

	// Services returns the list of registered services as ServiceClient(s)
	Services() []ServiceClient

	// ServiceInterfaces returns the service interfaces for all registered services
	ServiceInterfaces() []commons.InterfaceType

	// ServiceKeys returns the ServiceKey(s) for all registered services
	ServiceKeys() []*ServiceKey

	// RegisterService will create a new instance of the service using the supplied service constructor.
	// The ServiceClient must implement the service interface - otherwise the method panics.
	// It will then register the service and start it async.
	// If a service with the same service interface is already registered, then the service will not be started and nill will be returned.
	// The ServiceClientConstructor is retained until the service is unregistered for the purpose of restarting the service using a new instance.
	RegisterService(newService ServiceClientConstructor) ServiceClient

	// UnRegisterService will unregister the service and returns false if no such service is registered.
	// The service is simply unregistered, i.e., it is not stopped.
	UnRegisterService(service *Service) bool
}
