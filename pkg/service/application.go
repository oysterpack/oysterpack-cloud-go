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
	"os"
	"os/signal"
	stdreflect "reflect"
	"syscall"
	"time"

	"io"

	"fmt"

	"sync"

	"runtime"

	"github.com/oysterpack/oysterpack.go/pkg/commons/reflect"
	"github.com/oysterpack/oysterpack.go/pkg/logging"
	"github.com/oysterpack/oysterpack.go/pkg/metrics"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/rs/zerolog"
)

// Application is a service, which serves as a container for other services.
type Application interface {
	// Descriptor exposes the application Descriptor
	Descriptor() *Descriptor
	// UpdateDescriptor is exposed to set the application name and version before the application starts running.
	// This should be set in the application main() method before the application is started.
	UpdateDescriptor(namespace string, system string, component string, version string)

	// refers to the application service - used to manage the application lifecycle
	ServiceReference

	Registry

	ServiceManager

	ServiceDependencyChecks

	// Start the application and waits until the app is running
	Start()

	// Stop the application and waits until the app is stopped
	Stop()

	// RegisterShutdownHook registers a function that will be invoked when the application shutsdown.
	// The function will be invoked after all application managed services are stopped.
	RegisterShutdownHook(f func())

	// DumpAllStacks dumps all goroutine stacks to the specified writer
	DumpAllStacks(out io.Writer)
}

// ApplicationSettings provides application settings used to create a new application
type ApplicationSettings struct {
	*Descriptor

	LogOutput io.Writer
	LogLevel  *zerolog.Level
}

// NewApplicationDesc creates a new ApplicationDesc
func NewApplicationDesc(
	namespace string,
	system string,
	component string,
	version string,
) *Descriptor {
	var service Application = &application{}
	serviceInterface, _ := reflect.ObjectInterface(&service)
	return NewDescriptor(namespace, system, component, version, serviceInterface)
}

// NewApplication returns a new application
func NewApplication(settings ApplicationSettings) Application {
	app := &application{serviceTickets: &serviceTickets{}, services: make(map[Interface]*registeredService)}
	app.serviceDependencyChecks = &serviceDependencyChecks{app}
	var service Application = app

	if settings.Descriptor == nil {
		settings.Descriptor = NewApplicationDesc("oysterpack", "service", "application", "1.0.0")
	} else {
		serviceInterface, _ := reflect.ObjectInterface(&service)
		settings.Descriptor.serviceInterface = serviceInterface
	}

	app.service = NewService(Settings{
		Descriptor:  settings.Descriptor,
		LogSettings: LogSettings{LogOutput: settings.LogOutput, LogLevel: settings.LogLevel},
		Run:         app.run,
		Destroy:     app.destroy,
	})
	return service
}

// application.services map entry type
type registeredService struct {
	NewService ClientConstructor
	Client
}

// application is used to manage a set of services
// application itself is a service, by composition.
//
// Features
// 1. Triggering the application to shutdown, triggers each of its registered services to shutdown.
// 2. SIGTERM, SIGINT, SIGQUIT signals are trapped and trigger application shutdown.
//
// Use NewApplication() to create a new application instance
type application struct {
	servicesMutex sync.RWMutex
	// once a service is stopped, it will be removed from this map
	services map[Interface]*registeredService

	service Service

	*serviceTickets
	*serviceDependencyChecks

	shutdownHooks []func()

	serviceDependencyHealthCheck metrics.HealthCheck
}

func (a *application) DumpAllStacks(out io.Writer) {
	size := 1024 * 8
	for {
		buf := make([]byte, size)
		if len := runtime.Stack(buf, true); len <= size {
			out.Write(buf[:len])
			return
		}
		size = size + (1024 * 8)
	}
}

func (a *application) RegisterShutdownHook(f func()) {
	a.servicesMutex.Lock()
	defer a.servicesMutex.Unlock()
	a.shutdownHooks = append(a.shutdownHooks, func() {
		defer func() {
			if p := recover(); p != nil {
				a.service.Logger().Error().Err(fmt.Errorf("%v", p)).Msg("shutdown hook panic")
			}
		}()
		f()
	})
}

// Channel used to wait for the Client
// if nil is returned, then it means the channel was closed
func (a *ServiceTicket) Channel() <-chan Client {
	return a.channel
}

// Service returns the Application Service
func (a *application) Service() Service {
	return a.service
}

// Descriptor exposes the application Descriptor
func (a *application) Descriptor() *Descriptor {
	return a.service.Desc()
}

// UpdateDescriptor is exposed to set the application name and version before the application starts running.
// This should be set in the application main() method before the application is started.
func (a *application) UpdateDescriptor(namespace string, system string, component string, version string) {
	desc := NewApplicationDesc(namespace, system, component, version)
	a.Descriptor().namespace = desc.namespace
	a.Descriptor().system = desc.system
	a.Descriptor().component = desc.component
	a.Descriptor().version = desc.version
}

// ServiceByType looks up a service via its service interface.
// If exists is true, then the service was found.
func (a *application) ServiceByType(serviceInterface Interface) Client {
	a.servicesMutex.RLock()
	defer a.servicesMutex.RUnlock()
	return a.serviceByType(serviceInterface)
}

func (a *application) serviceByType(serviceInterface Interface) Client {
	if s, exists := a.services[serviceInterface]; exists {
		return s.Client
	}
	return nil
}

// ServiceByTypeAsync returns channel that will be used to send the Client, when one is available
func (a *application) ServiceByTypeAsync(serviceInterface Interface) *ServiceTicket {
	ticket := NewServiceTicket(serviceInterface)
	serviceClient := a.ServiceByType(serviceInterface)
	if serviceClient != nil {
		ticket.channel <- serviceClient
		close(ticket.channel)
		return ticket
	}

	a.serviceTickets.add(ticket)
	go a.checkServiceTickets(a)
	return ticket
}

// ServiceByKey looks up a service by ServiceKey and returns the registered Client.
// If the service is not found, then nil is returned.
func (a *application) ServiceByKey(serviceKey ServiceKey) Client {
	a.servicesMutex.RLock()
	defer a.servicesMutex.RUnlock()
	for _, s := range a.services {
		if InterfaceToServiceKey(s.Service().Desc().Interface()) == serviceKey {
			return s.Client
		}
	}
	return nil
}

// Services returns all registered services
func (a *application) Services() []Client {
	a.servicesMutex.RLock()
	defer a.servicesMutex.RUnlock()
	var services []Client
	for _, s := range a.services {
		services = append(services, s.Client)
	}
	return services
}

// ServiceInterfaces returns all service interfaces for all registered services
func (a *application) ServiceInterfaces() []Interface {
	a.servicesMutex.RLock()
	defer a.servicesMutex.RUnlock()
	interfaces := []Interface{}
	for k := range a.services {
		interfaces = append(interfaces, k)
	}
	return interfaces
}

// ServiceCount returns the number of registered services
func (a *application) ServiceCount() int {
	a.servicesMutex.RLock()
	defer a.servicesMutex.RUnlock()
	return len(a.services)
}

// ServiceKeys returns ServiceKey(s) for all registered services
func (a *application) ServiceKeys() []ServiceKey {
	a.servicesMutex.RLock()
	defer a.servicesMutex.RUnlock()
	interfaces := []ServiceKey{}
	for _, v := range a.ServiceInterfaces() {
		interfaces = append(interfaces, InterfaceToServiceKey(v))
	}
	return interfaces
}

// MustRegisterService will register the service and start it, if it is not already registered.
// A panic occurs if registration fails for any of the following reasons:
// 1. Interface is nil
// 2. Version is nil
// 3. the Client instance type is not assignable to the Service.
// 4. A service with the same Interface is already registered.
func (a *application) MustRegisterService(create ClientConstructor) Client {
	validate := func(service Client) {
		if service.Service().Desc().Interface() == nil {
			a.Service().Logger().Panic().Msg("Service failed to register because it has no Interface")
		}

		if service.Service().Desc().Version() == nil {
			a.Service().Logger().Panic().
				Str(logging.SERVICE, service.Service().Desc().Interface().String()).
				Msgf("Service failed to register because it has no version")
		}

		if !stdreflect.TypeOf(service).AssignableTo(service.Service().Desc().Interface()) {
			a.Service().Logger().Panic().
				Str(logging.SERVICE, service.Service().Desc().Interface().String()).
				Msgf("Service failed to register because %T is not assignable to %v", stdreflect.TypeOf(service), service.Service().Desc().Interface())
		}
	}

	a.servicesMutex.Lock()
	defer a.servicesMutex.Unlock()
	service := create(Application(a))
	validate(service)
	if _, exists := a.services[service.Service().Desc().Interface()]; exists {
		a.Service().Logger().Panic().
			Str(logging.SERVICE, service.Service().Desc().Interface().String()).
			Msgf("Service is already registered : %v", service.Service().Desc().Interface())
	}
	a.services[service.Service().Desc().Interface()] = &registeredService{NewService: create, Client: service}
	service.Service().StartAsync()
	go a.checkServiceTickets(a)
	return service
}

func (a *application) RegisterService(create ClientConstructor) (client Client, err error) {
	defer func() {
		if p := recover(); p != nil {
			err = fmt.Errorf("%v", p)
		}
	}()
	return a.MustRegisterService(create), nil
}

// UnRegisterService unregisters the specified service.
// The service is simply unregistered, i.e., it is not stopped.
func (a *application) UnRegisterService(service Client) bool {
	a.servicesMutex.Lock()
	defer a.servicesMutex.Unlock()
	if _, exists := a.services[service.Service().Desc().Interface()]; exists {
		delete(a.services, service.Service().Desc().Interface())
		return true
	}
	return false
}

func (a *application) run(ctx *Context) error {
	a.service.Logger().Info().Msg(a.Descriptor().ID())

	if a.serviceDependencyHealthCheck == nil {
		gauge := prometheus.GaugeOpts{
			Namespace:   metrics.METRIC_NAMESPACE_OYSTERPACK,
			Subsystem:   "application",
			Name:        "check_service_dependencies",
			Help:        "The healthcheck fails if any dependencies are not satisfied.",
			ConstLabels: AddServiceMetricLabels(prometheus.Labels{}, a.Descriptor()),
		}
		runInterval := time.Minute
		healthcheck := func() error { return a.CheckAllServiceDependencies() }
		a.serviceDependencyHealthCheck = metrics.NewHealthCheck(gauge, runInterval, healthcheck)
	}

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)
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
	metrics.HealthChecks.StopAllHealthCheckTickers()
	a.serviceTickets.closeAllServiceTickets()

	a.servicesMutex.RLock()
	defer a.servicesMutex.RUnlock()

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

	for _, f := range a.shutdownHooks {
		f()
	}
	return nil
}

func (a *application) ServiceStates() map[Interface]State {
	a.servicesMutex.RLock()
	defer a.servicesMutex.RUnlock()
	serviceStates := make(map[Interface]State)
	for k, s := range a.services {
		serviceStates[k] = s.Service().State()
	}
	return serviceStates
}

func (a *application) StopServiceByType(serviceInterface Interface) error {
	client := a.ServiceByType(serviceInterface)
	if client == nil {
		return &ServiceNotFoundError{InterfaceToServiceKey(serviceInterface)}
	}
	client.Service().StopAsyc()
	return nil
}

func (a *application) StopServiceByKey(key ServiceKey) error {
	client := a.ServiceByKey(key)
	if client == nil {
		return &ServiceNotFoundError{key}
	}
	client.Service().StopAsyc()
	return nil
}

func (a *application) RestartServiceByType(serviceInterface Interface) error {
	client := a.ServiceByType(serviceInterface)
	if client == nil {
		return &ServiceNotFoundError{InterfaceToServiceKey(serviceInterface)}
	}
	client.Restart()
	return nil
}

func (a *application) RestartServiceByKey(key ServiceKey) error {
	client := a.ServiceByKey(key)
	if client == nil {
		return &ServiceNotFoundError{key}
	}
	client.Restart()
	return nil
}

func (a *application) RestartAllServices() {
	a.servicesMutex.RLock()
	defer a.servicesMutex.RUnlock()
	for _, client := range a.services {
		go client.Restart()
	}
}

func (a *application) RestartAllFailedServices() {
	a.servicesMutex.RLock()
	defer a.servicesMutex.RUnlock()
	for _, client := range a.services {
		if client.Service().FailureCause() != nil {
			go client.Restart()
		}
	}
}

func (a *application) StopAllServices() {
	a.servicesMutex.RLock()
	defer a.servicesMutex.RUnlock()
	for _, client := range a.services {
		client.Service().Stop()
	}
}

func (a *application) Start() {
	a.Service().StartAsync()
	a.Service().AwaitUntilRunning()
}

func (a *application) Stop() {
	a.Service().Stop()
}
