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

	"github.com/oysterpack/oysterpack.go/pkg/commons/reflect"
)

// Registry is a Client registry.
// It is used to register/unregister Client(s) and lookup Client(s).
type Registry interface {
	// ServiceByType looks up a service and returns nil if the service is not founc.
	ServiceByType(serviceInterface ServiceInterface) Client

	// ServiceByTypeAsync returns a ServiceTicket that can be used to receive the service client via a channel.
	ServiceByTypeAsync(serviceInterface ServiceInterface) *ServiceTicket

	// Returns a snapshot of the number of open ServiceTicket(s) per ServiceInterface
	ServiceTicketCounts() map[ServiceInterface]int

	// ServiceByKey looks up a service and returns nil if the service is not founc.
	ServiceByKey(key ServiceKey) Client

	// Services returns the list of registered services as Client(s)
	Services() []Client

	// ServiceCount returns the number of services registered
	ServiceCount() int

	// ServiceInterfaces returns the service interfaces for all registered services
	ServiceInterfaces() []ServiceInterface

	// ServiceKeys returns the ServiceKey(s) for all registered services
	ServiceKeys() []ServiceKey

	// MustRegisterService will create a new instance of the service using the supplied service constructor.
	// The Client must implement the service interface - otherwise the method panics.
	// It will then register the service and start it async.
	// If a service with the same service interface is already registered, then the service will not be started and nil will be returned.
	// NOTE: only a single version of the service may be registered
	//
	// A panic occurs if registration fails for any of the following reasons:
	// 1. ServiceInterface is nil
	// 2. Version is nil
	// 3. the Client instance type is not assignable to the Service.
	MustRegisterService(newService ClientConstructor) Client

	RegisterService(newService ClientConstructor) (Client, error)

	// UnRegisterService will unregister the service and returns false if no such service is registered.
	// The service is simply unregistered, i.e., it is not stopped.
	// This will not normally be used in an application. It's main purpose is to support testing.
	UnRegisterService(service Client) bool
}

// ServiceKey represents the service interface.
// It can be used to lookup a service.
type ServiceKey struct {
	reflect.PackagePath
	reflect.TypeName
}

func (s *ServiceKey) String() string {
	return fmt.Sprintf("%v.%v", s.PackagePath, s.TypeName)
}

// InterfaceToServiceKey converts an interface type to a ServiceKey
func InterfaceToServiceKey(serviceInterface ServiceInterface) ServiceKey {
	return ServiceKey{
		reflect.PackagePath(serviceInterface.PkgPath()),
		reflect.TypeName(serviceInterface.Name()),
	}
}
