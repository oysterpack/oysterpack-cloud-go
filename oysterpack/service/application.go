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
	"os"
	"os/signal"
	"reflect"
	"sync"
	"syscall"
	"time"

	"io"

	"github.com/oysterpack/oysterpack.go/oysterpack/commons"
)

// Application is a service, which serves as a container for other services.
type Application interface {
	// refers to the application service - used to manage the application lifecycle
	ServiceReference

	Registry

	ServiceDependencies

	// ApplicationServiceState(s) will be returned sorted by the service interface
	ServiceStates() map[ServiceInterface]State

	//StartStopRestartServices
}

// application.services map entry type
type registeredService struct {
	NewService ServiceClientConstructor
	Client
}

// application is used to manage a set of services
// application itself is a service, by composition.
//
// Features
// 1. Triggering the application to shutdown, triggers each of its registered services to shutdown.
// 2. Triggers shutdown when SIGTERM and SIGINT signals are received
//
// Use NewApplication() to create a new application instance
type application struct {
	mutex sync.RWMutex
	// once a service is stopped, it will be removed from this map
	services map[ServiceInterface]*registeredService

	// application can be managed itself as a service
	service Service

	// used to keep track of users who are waiting on services
	serviceTicketsMutex sync.RWMutex
	serviceTickets      []*ServiceTicket
}

// ServiceTicket represents a ticket issued to a user waiting for a service
type ServiceTicket struct {
	// the type of service the user is waiting for
	ServiceInterface
	// used to deliver the Client to the user
	channel chan Client

	// when the ticket was created
	time.Time
}

// Channel used to wait for the Client
// if nil is returned, then it means the channel was closed
func (a *ServiceTicket) Channel() <-chan Client {
	return a.channel
}

// ServiceTicketCounts returns the number of tickets that have been issued per service
func (a *application) ServiceTicketCounts() map[commons.InterfaceType]int {
	a.serviceTicketsMutex.RLock()
	defer a.serviceTicketsMutex.RUnlock()
	counts := make(map[commons.InterfaceType]int)
	for _, ticket := range a.serviceTickets {
		counts[ticket.ServiceInterface]++
	}
	return counts
}

// Service returns the Application Service
func (a *application) Service() Service {
	return a.service
}

// ApplicationSettings provides application settings used to create a new application
type ApplicationSettings struct {
	LogOutput io.Writer
}

// NewApplication returns a new application
func NewApplication(settings ApplicationSettings) Application {
	app := &application{
		services: make(map[ServiceInterface]*registeredService),
	}
	var service Application = app
	serviceInterface, _ := commons.ObjectInterface(&service)
	app.service = NewService(ServiceSettings{
		ServiceInterface: serviceInterface,
		Run:              app.run,
		Destroy:          app.destroy,
		LogOutput:        settings.LogOutput,
	})
	return service
}

// ServiceByType looks up a service via its service interface.
// If exists is true, then the service was found.
func (a *application) ServiceByType(serviceInterface ServiceInterface) Client {
	a.mutex.RLock()
	defer a.mutex.RUnlock()
	return a.serviceByType(serviceInterface)
}

func (a *application) serviceByType(serviceInterface ServiceInterface) Client {
	if s, exists := a.services[serviceInterface]; exists {
		return s.Client
	}
	return nil
}

// ServiceByTypeAsync returns channel that will be used to send the Client, when one is available
func (a *application) ServiceByTypeAsync(serviceInterface ServiceInterface) *ServiceTicket {
	ticket := &ServiceTicket{serviceInterface, make(chan Client, 1), time.Now()}
	serviceClient := a.ServiceByType(serviceInterface)
	if serviceClient != nil {
		ticket.channel <- serviceClient
		close(ticket.channel)
		return ticket
	}

	a.serviceTicketsMutex.Lock()
	a.serviceTickets = append(a.serviceTickets, ticket)
	a.serviceTicketsMutex.Unlock()

	go a.checkServiceTickets()
	return ticket
}

func (a *application) checkServiceTickets() {
	a.serviceTicketsMutex.RLock()
	defer a.serviceTicketsMutex.RUnlock()
	for _, ticket := range a.serviceTickets {
		serviceClient := a.ServiceByType(ticket.ServiceInterface)
		if serviceClient != nil {
			go func(ticket *ServiceTicket) {
				defer commons.IgnorePanic()
				ticket.channel <- serviceClient
				close(ticket.channel)
			}(ticket)
			go a.deleteServiceTicket(ticket)
		}
	}
}

func (a *application) deleteServiceTicket(ticket *ServiceTicket) {
	a.serviceTicketsMutex.Lock()
	defer a.serviceTicketsMutex.Unlock()
	for i, elem := range a.serviceTickets {
		if elem == ticket {
			a.serviceTickets[i] = a.serviceTickets[len(a.serviceTickets)-1] // Replace it with the last one.
			a.serviceTickets = a.serviceTickets[:len(a.serviceTickets)-1]
			return
		}
	}
}

func (a *application) closeAllServiceTickets() {
	a.serviceTicketsMutex.RLock()
	defer a.serviceTicketsMutex.RUnlock()
	for _, ticket := range a.serviceTickets {
		func(ticket *ServiceTicket) {
			defer commons.IgnorePanic()
			close(ticket.channel)
		}(ticket)
	}
}

// ServiceByKey looks up a service by ServiceKey and returns the registered Client.
// If the service is not found, then nil is returned.
func (a *application) ServiceByKey(serviceKey ServiceKey) Client {
	a.mutex.RLock()
	defer a.mutex.RUnlock()
	for _, s := range a.services {
		if InterfaceTypeToServiceKey(s.Service().Interface()) == serviceKey {
			return s.Client
		}
	}
	return nil
}

// Services returns all registered services
func (a *application) Services() []Client {
	a.mutex.RLock()
	defer a.mutex.RUnlock()
	var services []Client
	for _, s := range a.services {
		services = append(services, s.Client)
	}
	return services
}

// ServiceInterfaces returns all service interfaces for all registered services
func (a *application) ServiceInterfaces() []ServiceInterface {
	a.mutex.RLock()
	defer a.mutex.RUnlock()
	interfaces := []ServiceInterface{}
	for k := range a.services {
		interfaces = append(interfaces, k)
	}
	return interfaces
}

// ServiceCount returns the number of registered services
func (a *application) ServiceCount() int {
	return len(a.services)
}

// ServiceKeys returns ServiceKey(s) for all registered services
func (a *application) ServiceKeys() []ServiceKey {
	a.mutex.RLock()
	defer a.mutex.RUnlock()
	interfaces := []ServiceKey{}
	for _, v := range a.ServiceInterfaces() {
		interfaces = append(interfaces, InterfaceTypeToServiceKey(v))
	}
	return interfaces
}

// RegisterService will register the service and start it, if it is not already registered.
// Returns the new registered service or nil if a service with the same interface was already registered.
// A panic occurs if the Client type is not assignable to the Service.
func (a *application) RegisterService(newService ServiceClientConstructor) Client {
	a.mutex.Lock()
	defer a.mutex.Unlock()
	service := newService(Application(a))
	if !reflect.TypeOf(service).AssignableTo(service.Service().Interface()) {
		panic(fmt.Sprintf("%T is not assignable to %v", reflect.TypeOf(service), service.Service().Interface()))
	}
	if _, exists := a.services[service.Service().Interface()]; !exists {
		a.services[service.Service().Interface()] = &registeredService{NewService: newService, Client: service}
		service.Service().StartAsync()
		go a.checkServiceTickets()
		return service
	}
	return nil
}

// UnRegisterService unregisters the specified service.
// The service is simply unregistered, i.e., it is not stopped.
func (a *application) UnRegisterService(service Client) bool {
	a.mutex.Lock()
	defer a.mutex.Unlock()
	if _, exists := a.services[service.Service().Interface()]; exists {
		delete(a.services, service.Service().Interface())
		return true
	}
	return false
}

func (a *application) run(ctx *Context) error {
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

func (a *application) destroy(ctx *Context) error {
	a.closeAllServiceTickets()
	for _, v := range a.services {
		v.Service().StopAsyc()
	}
	for _, v := range a.services {
		for {
			v.Service().AwaitStopped(5 * time.Second)
			if v.Service().State().Stopped() {
				break
			}
			v.Service().Logger().Warn().Msg("Waiting for service to stop")
		}
	}
	return nil
}

// CheckAllServiceDependenciesRegistered checks that are service Dependencies are currently satisfied.
func (a *application) CheckAllServiceDependenciesRegistered() []*ServiceDependenciesMissing {
	a.mutex.RLock()
	defer a.mutex.RUnlock()
	errors := []*ServiceDependenciesMissing{}
	for _, serviceClient := range a.services {
		if missingDependencies := a.checkServiceDependenciesRegistered(serviceClient); missingDependencies != nil {
			errors = append(errors, missingDependencies)
		}
	}
	return errors
}

// CheckServiceDependenciesRegistered checks that the service's Dependencies are currently satisfied
// nil is returned if there is no error
func (a *application) CheckServiceDependenciesRegistered(serviceClient Client) *ServiceDependenciesMissing {
	a.mutex.RLock()
	defer a.mutex.RUnlock()
	return a.checkServiceDependenciesRegistered(serviceClient)
}

func (a *application) checkServiceDependenciesRegistered(serviceClient Client) *ServiceDependenciesMissing {
	missingDependencies := &ServiceDependenciesMissing{&DependencyMappings{ServiceInterface: serviceClient.Service().Interface()}}
	for _, dependency := range serviceClient.Service().ServiceDependencies() {
		if a.serviceByType(dependency) == nil {
			missingDependencies.AddMissingDependency(dependency)
		}
	}
	if missingDependencies.HasMissing() {
		return missingDependencies
	}
	return nil
}

// CheckAllServiceDependenciesRunning checks that are service Dependencies are currently satisfied.
func (a *application) CheckAllServiceDependenciesRunning() []*ServiceDependenciesNotRunning {
	a.mutex.RLock()
	defer a.mutex.RUnlock()
	errors := []*ServiceDependenciesNotRunning{}
	for _, serviceClient := range a.services {
		if notRunning := a.checkServiceDependenciesRunning(serviceClient); notRunning != nil {
			errors = append(errors, notRunning)
		}
	}
	return errors
}

// CheckServiceDependenciesRunning checks that the service's Dependencies are currently satisfied
// nil is returned if there is no error
func (a *application) CheckServiceDependenciesRunning(serviceClient Client) *ServiceDependenciesNotRunning {
	a.mutex.RLock()
	defer a.mutex.RUnlock()
	return a.checkServiceDependenciesRunning(serviceClient)
}

func (a *application) checkServiceDependenciesRunning(serviceClient Client) *ServiceDependenciesNotRunning {
	notRunning := &ServiceDependenciesNotRunning{&DependencyMappings{ServiceInterface: serviceClient.Service().Interface()}}
	for _, dependency := range serviceClient.Service().ServiceDependencies() {
		if client := a.serviceByType(dependency); client == nil || !client.Service().State().Running() {
			notRunning.AddDependencyNotRunning(dependency)
		}
	}
	if notRunning.HasNotRunning() {
		return notRunning
	}
	return nil
}

// CheckAllServiceDependencies checks that all Dependencies for each service are available
// nil is returned if there are no errors
func (a *application) CheckAllServiceDependencies() *ServiceDependencyErrors {
	errors := []error{}
	for _, err := range a.CheckAllServiceDependenciesRegistered() {
		errors = append(errors, err)
	}

	for _, err := range a.CheckAllServiceDependenciesRunning() {
		errors = append(errors, err)
	}

	if len(errors) > 0 {
		return &ServiceDependencyErrors{errors}
	}
	return nil
}

// CheckServiceDependencies checks that the service Dependencies are available
func (a *application) CheckServiceDependencies(client Client) *ServiceDependencyErrors {
	errors := []error{}
	if err := a.CheckServiceDependenciesRegistered(client); err != nil {
		errors = append(errors, err)
	}

	if err := a.CheckServiceDependenciesRunning(client); err != nil {
		errors = append(errors, err)
	}

	if len(errors) > 0 {
		return &ServiceDependencyErrors{errors}
	}
	return nil
}

func (a *application) ServiceStates() map[ServiceInterface]State {
	a.mutex.RLock()
	defer a.mutex.RUnlock()
	serviceStates := make(map[ServiceInterface]State)
	for k, s := range a.services {
		serviceStates[k] = s.Service().State()
	}
	return serviceStates
}

func (a *application) StopServiceByType(serviceInterface ServiceInterface) error {
	client := a.ServiceByType(serviceInterface)
	if client == nil {
		return &ServiceNotFoundError{InterfaceTypeToServiceKey(serviceInterface)}
	}
	return client.Service().StartAsync()
}

func (a *application) StopServiceByKey(key ServiceKey) error {
	client := a.ServiceByKey(key)
	if client == nil {
		return &ServiceNotFoundError{key}
	}
	return client.Service().StartAsync()
}
