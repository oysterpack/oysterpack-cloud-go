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

package service_test

import (
	stdreflect "reflect"
	"sync"
	"testing"
	"time"

	"github.com/oysterpack/oysterpack.go/oysterpack/commons/reflect"
	"github.com/oysterpack/oysterpack.go/oysterpack/service"
)

func TestApplicationContext_RegisterService(t *testing.T) {
	app := service.NewApplication(service.ApplicationSettings{})
	app.Service().StartAsync()
	app.Service().AwaitUntilRunning()
	defer app.Service().Stop()
	serviceClient := app.MustRegisterService(EchoServiceClientConstructor)

	func() {
		defer func() {
			if p := recover(); p == nil {
				t.Error("registration should have panicked because the service is already registered")
			} else {
				t.Logf("registration failed as expected because : %v", p)
			}
		}()
		app.MustRegisterService(EchoServiceClientConstructor)
	}()

	if err := serviceClient.Service().AwaitUntilRunning(); err != nil {
		t.Fatal(err)
	}

	if !serviceClient.Service().State().Running() {
		t.Fatal("service should be running")
	}
	echoService := serviceClient.(EchoService)
	message := "CIAO MUNDO !!!"
	if response := echoService.Echo(message); response != message {
		t.Errorf("echo : return unexpected response : %v", response)
	}

	serviceClients := []service.Client{
		serviceClient,
		app.ServiceByType(EchoServiceInterface),
		app.ServiceByKey(service.InterfaceTypeToServiceKey(EchoServiceInterface)),
	}

	for i, client := range serviceClients {
		if client == nil {
			t.Errorf("%d : the service client should not be nil - it was registered", i)
		}
		t.Logf("%d : %v", i, client.(EchoService).Echo(message))

		// check that all service clients point to the same service
		if i < len(serviceClients)-1 {
			if serviceClients[i].Service() != serviceClients[i+1].Service() {
				t.Errorf("each service client should point to the same backend service instance: %d != %d", i, i+1)
			}
		}
	}
}

func TestApplicationContext_ServiceClientIsStableReferenceAfterRestarting(t *testing.T) {
	app := service.NewApplication(service.ApplicationSettings{})
	app.Service().StartAsync()
	app.Service().AwaitUntilRunning()
	defer app.Service().Stop()
	serviceClient := app.MustRegisterService(EchoServiceClientConstructor)
	message := "CIAO MUNDO !!!"

	serviceClients := []service.Client{
		serviceClient,
		app.ServiceByType(EchoServiceInterface),
		app.ServiceByKey(service.InterfaceTypeToServiceKey(EchoServiceInterface)),
	}

	app.Service().Stop()
	for i, client := range serviceClients {
		if !client.Service().State().Stopped() {
			t.Error("service should be stopped : %d", i)
		}
	}
	serviceClient.Restart()
	serviceClient.Service().AwaitUntilRunning()
	for i, client := range serviceClients {
		if !client.Service().State().Running() {
			t.Error("service should be running : %d", i)
		} else {
			t.Logf("%d : %v", i, client.(EchoService).Echo(message))
		}

		// check that all service clients point to the same service
		if i < len(serviceClients)-1 {
			if serviceClients[i].Service() != serviceClients[i+1].Service() {
				t.Errorf("each service client should point to the same backend service instance: %d != %d", i, i+1)
			}
		}
	}

	serviceClient.Service().Stop()
	t.Log("service is stopped")
	runningWaitGroup := sync.WaitGroup{}
	runningWaitGroup.Add(1)
	echoWaitGroup := sync.WaitGroup{}
	echoWaitGroup.Add(1)
	go func() {
		t.Log("invoking echo while service is stopped")
		runningWaitGroup.Done()
		for i, client := range serviceClients {
			t.Logf("echo after service restarted : %d : %v", i, client.(EchoService).Echo(message))
		}
		echoWaitGroup.Done()
	}()
	runningWaitGroup.Wait()
	serviceClient.Restart()
	t.Log("service is being restarted ...")
	echoWaitGroup.Wait()
}

func TestApplicationContext_UnRegisterService(t *testing.T) {
	app := service.NewApplication(service.ApplicationSettings{})
	app.Service().StartAsync()
	app.Service().AwaitUntilRunning()
	defer app.Service().Stop()
	serviceClient := app.MustRegisterService(EchoServiceClientConstructor)
	if _, err := app.RegisterService(EchoServiceClientConstructor); err == nil {
		t.Errorf("registering a service that is already registered should fail")
	}
	if app.ServiceCount() != 1 {
		t.Errorf("service count should be 1 , but was %d", app.ServiceCount())
	}
	if !app.UnRegisterService(serviceClient) {
		t.Errorf("the service should have been unregistered")
	}
	if app.ServiceCount() != 0 {
		t.Errorf("service count should be 0 , but was %d", app.ServiceCount())
	}
	if app.UnRegisterService(serviceClient) {
		t.Errorf("the service should not be registered")
	}
	serviceClient2 := app.MustRegisterService(EchoServiceClientConstructor)
	if serviceClient2 == nil {
		t.Errorf("service should have been registered")
	}
	if serviceClient == serviceClient2 {
		t.Errorf("A new Client instance should have been returned")
	}
	if serviceClient2 != app.ServiceByType(EchoServiceInterface) {
		t.Errorf("the registered service client instance should be the same instance")
	}
	if serviceClient == app.ServiceByType(EchoServiceInterface) {
		t.Errorf("the registered service client instance should be different")
	}
}

type SimpleEchoService struct{}

func (a *SimpleEchoService) Echo(msg interface{}) interface{} {
	return msg
}

func TestApplicationContext_RegisterService_ServiceClientNotAssignableToServiceInterface(t *testing.T) {
	app := service.NewApplication(service.ApplicationSettings{})
	app.Service().StartAsync()
	app.Service().AwaitUntilRunning()
	defer app.Service().Stop()

	type InvalidEchoServiceClient struct {
		// embedded service
		*service.RestartableService
	}

	func() {
		defer func() {
			if p := recover(); p != nil {
				t.Logf("panic was expected : [%v]", p)
			} else {
				t.Errorf("Application.MustRegisterService should have panicked")
			}
		}()

		app.MustRegisterService(func(app service.Application) service.Client {
			return &InvalidEchoServiceClient{
				RestartableService: service.NewRestartableService(func() service.Service {
					var svc EchoService = &SimpleEchoService{}
					serviceInterface, _ := reflect.ObjectInterface(&svc)
					return service.NewService(service.ServiceSettings{ServiceInterface: serviceInterface})
				}),
			}
		})
	}()

}

func TestApplicationContext_ServiceByType_NotRegistered(t *testing.T) {
	app := service.NewApplication(service.ApplicationSettings{})
	app.Service().StartAsync()
	app.Service().AwaitUntilRunning()
	defer app.Service().Stop()

	if app.ServiceByType(EchoServiceInterface) != nil {
		t.Error("ERROR: no services are registered")
	}
	if app.ServiceByKey(service.InterfaceTypeToServiceKey(EchoServiceInterface)) != nil {
		t.Error("ERROR: no services are registered")
	}

	app.MustRegisterService(EchoServiceClientConstructor)

	serviceClients := []service.Client{
		app.ServiceByType(EchoServiceInterface),
		app.ServiceByKey(service.InterfaceTypeToServiceKey(EchoServiceInterface)),
	}

	for i, client := range serviceClients {
		if client == nil {
			t.Errorf("ERROR: services is registered : %d", i)
		}
		client.(EchoService).Echo("OysterPack is your world.")
	}

	type Foo interface{}
	type Bar struct{}
	var foo Foo = &Bar{}
	fooType, err := reflect.ObjectInterface(&foo)
	if err != nil {
		t.Fatalf("Failed to get Foo type : %v", err)
	}
	if app.ServiceByType(fooType) != nil {
		t.Error("ERROR: no Foo service is registered")
	}
}

func TestApplicationContext_ServiceInterfaces(t *testing.T) {
	app := service.NewApplication(service.ApplicationSettings{})
	app.Service().StartAsync()
	app.Service().AwaitUntilRunning()
	defer app.Service().Stop()

	serviceInterfaces := app.ServiceInterfaces()
	if app.ServiceCount() != 0 {
		t.Errorf("ERROR: there should be 0 services registered : %d : %v", len(serviceInterfaces), serviceInterfaces)
	}
	if len(serviceInterfaces) != 0 {
		t.Errorf("ERROR: there should be 0 services registered : %d : %d : %v", app.ServiceCount(), len(serviceInterfaces), serviceInterfaces)
	}

	echoService := app.MustRegisterService(EchoServiceClientConstructor).(EchoService)
	serviceInterfaces = app.ServiceInterfaces()
	if app.ServiceCount() != 1 {
		t.Errorf("ERROR: there should be 1 services registered : %d : %v", len(serviceInterfaces), serviceInterfaces)
	}
	if len(serviceInterfaces) != 1 {
		t.Errorf("ERROR: there should be 1 services registered : %d : %d : %v", app.ServiceCount(), len(serviceInterfaces), serviceInterfaces)
	}
	t.Logf("echo : %v", echoService.Echo("TestApplicationContext_ServiceInterfaces"))
	heartbeatService := app.MustRegisterService(HeartbeatServiceClientConstructor).(HeartbeatService)
	t.Logf("heartbeat ping : %v", heartbeatService.Ping())

	serviceInterfaces = app.ServiceInterfaces()
	t.Logf("serviceInterfaces : %v", serviceInterfaces)
	if app.ServiceCount() != 2 {
		t.Errorf("ERROR: there should be 2 services registered : %d : %v", len(serviceInterfaces), serviceInterfaces)
	}
	if len(serviceInterfaces) != 2 {
		t.Errorf("ERROR: there should be 2 services registered : %d : %d : %v", app.ServiceCount(), len(serviceInterfaces), serviceInterfaces)
	}

	containsInterface := func(svcInterface reflect.InterfaceType) bool {
		for _, serviceInterface := range serviceInterfaces {
			if serviceInterface == svcInterface {
				return true
			}
		}
		return false
	}

	if !containsInterface(HeartbeatServiceInterface) {
		t.Error("ERROR: HeartbeatService was not found")
	}

	if !containsInterface(EchoServiceInterface) {
		t.Error("ERROR: EchoService was not found")
	}
}

func TestApplicationContext_Services(t *testing.T) {
	app := service.NewApplication(service.ApplicationSettings{})
	app.Service().StartAsync()
	app.Service().AwaitUntilRunning()
	defer app.Service().Stop()

	app.MustRegisterService(EchoServiceClientConstructor)
	services := app.Services()
	if len(services) != 1 {
		t.Errorf("there should be 1 service registered : %v", services)
	}

	app.MustRegisterService(HeartbeatServiceClientConstructor)
	services = app.Services()

	if len(services) != 2 {
		t.Errorf("ERROR: there should be 2 services registered : %v", services)
	}

	containsService := func(svcInterface reflect.InterfaceType) bool {
		for _, service := range services {
			if stdreflect.TypeOf(service).AssignableTo(svcInterface) {
				return true
			}
		}
		return false
	}

	if !containsService(HeartbeatServiceInterface) {
		t.Error("ERROR: HeartbeatService was not found")
	}

	if !containsService(EchoServiceInterface) {
		t.Error("ERROR: EchoService was not found")
	}
}

func TestApplicationContext_ServiceKeys(t *testing.T) {
	app := service.NewApplication(service.ApplicationSettings{})
	app.Service().StartAsync()
	app.Service().AwaitUntilRunning()
	defer app.Service().Stop()

	serviceKeys := app.ServiceKeys()
	if len(serviceKeys) != 0 {
		t.Errorf("ERROR: there should be 0 services registered : %d : %d : %v", app.ServiceCount(), len(serviceKeys), serviceKeys)
	}

	app.MustRegisterService(EchoServiceClientConstructor)
	serviceKeys = app.ServiceKeys()
	if len(serviceKeys) != 1 {
		t.Errorf("ERROR: there should be 1 services registered : %v", serviceKeys)
	}
	app.MustRegisterService(HeartbeatServiceClientConstructor)
	serviceKeys = app.ServiceKeys()
	t.Logf("serviceKeys : %v", serviceKeys)
	if len(serviceKeys) != 2 {
		t.Errorf("ERROR: there should be 2 services registered : %d : %d : %v", app.ServiceCount(), len(serviceKeys), serviceKeys)
	}

	containsInterface := func(svcInterface reflect.InterfaceType) bool {
		for _, key := range serviceKeys {
			if service.InterfaceTypeToServiceKey(svcInterface) == key {
				return true
			}
		}
		return false
	}

	if !containsInterface(HeartbeatServiceInterface) {
		t.Error("ERROR: HeartbeatService was not found")
	}

	if !containsInterface(EchoServiceInterface) {
		t.Error("ERROR: EchoService was not found")
	}
}

func TestApplicationContext_ServiceByTypeAsync(t *testing.T) {
	app := service.NewApplication(service.ApplicationSettings{})
	app.Service().StartAsync()
	app.Service().AwaitUntilRunning()
	defer app.Service().Stop()

	wait := sync.WaitGroup{}
	serviceTicket := app.ServiceByTypeAsync(EchoServiceInterface)
	wait.Add(1)
	go func() {
		defer wait.Done()
		serviceClient := <-serviceTicket.Channel()
		echoService := serviceClient.(EchoService)
		t.Log(echoService.Echo("service ticket has been fulfilled"))
	}()
	if app.ServiceTicketCounts()[EchoServiceInterface] != 1 {
		t.Errorf("There should be ticket in the queue for the EchoService : %d", app.ServiceTicketCounts()[EchoServiceInterface])
	}
	app.MustRegisterService(EchoServiceClientConstructor)
	wait.Wait()

	serviceTicket = app.ServiceByTypeAsync(EchoServiceInterface)
	serviceClient := <-serviceTicket.Channel()
	echoService := serviceClient.(EchoService)
	t.Log(echoService.Echo("service ticket has been fulfilled"))

	const count = 100
	wait.Add(count)
	for i := 0; i < count; i++ {
		go func(index int, serviceTicket *service.ServiceTicket) {
			defer wait.Done()
			serviceClient := <-serviceTicket.Channel()
			echoService := serviceClient.(EchoService)
			t.Logf("#%d : %s", index, echoService.Echo("service ticket has been fulfilled"))
		}(i, app.ServiceByTypeAsync(EchoServiceInterface))
	}
	wait.Wait()
}

func TestApplicationContext_CheckAllServiceDependenciesRegistered(t *testing.T) {
	app := service.NewApplication(service.ApplicationSettings{})
	app.Service().StartAsync()
	app.Service().AwaitUntilRunning()
	defer app.Service().Stop()

	app.MustRegisterService(EchoServiceClientConstructor)
	app.MustRegisterService(HeartbeatServiceClientConstructor)

	servicesMissingDependencies := app.CheckAllServiceDependenciesRegistered()
	if len(servicesMissingDependencies) > 0 {
		t.Errorf("There should be no services missing dependencies : %v", servicesMissingDependencies)
	}

	bService := app.MustRegisterService(BServiceClientConstructor)

	servicesMissingDependencies = app.CheckAllServiceDependenciesRegistered()
	t.Log(servicesMissingDependencies)
	if len(servicesMissingDependencies) != 1 {
		t.Errorf("BService should be missing AService dependency: %v", servicesMissingDependencies)
	}

	checkMissingDepencencyServiceA := func(err *service.ServiceDependenciesMissing) {
		if err.ServiceInterface != bService.Service().Interface() {
			t.Errorf("ServiceInterface should be: %v , but was %v", bService.Service().Interface(), err.ServiceInterface)
		}
		if len(err.Dependencies) != 1 {
			t.Errorf("There shold be i missing Dependencies : %v", err.Dependencies)
		}
		if err.Dependencies[0] != AServiceInterface {
			t.Errorf("Missing service dependency should be AService, but was : %v", err.Dependencies[0])
		}
	}

	for _, err := range servicesMissingDependencies {
		checkMissingDepencencyServiceA(err)
	}

	err := app.CheckServiceDependenciesRegistered(bService)
	if err == nil {
		t.Error("BService should be missing AService dependency")
	}
	checkMissingDepencencyServiceA(err)

	bService.Service().AwaitRunning(time.Duration(5 * time.Millisecond))
	if !bService.Service().State().Starting() {
		t.Errorf("BService should be blocked while starting waiting on reference to A : %v", bService.Service().State())
	}

	app.MustRegisterService(AServiceClientConstructorFactory("1.0.0"))
	bService.Service().AwaitUntilRunning()

	servicesMissingDependencies = app.CheckAllServiceDependenciesRegistered()
	if len(servicesMissingDependencies) > 0 {
		t.Errorf("There should be no services missing dependencies : %v", servicesMissingDependencies)
	}
}

func TestApplicationContext_StopAppWhileWaitingForServiceDependencies(t *testing.T) {
	app := service.NewApplication(service.ApplicationSettings{})
	app.Service().StartAsync()
	app.Service().AwaitUntilRunning()

	app.MustRegisterService(BServiceClientConstructor)
	dependencyErrors := app.CheckAllServiceDependencies()
	t.Log(dependencyErrors)
	if len(dependencyErrors.Errors) == 0 {
		t.Errorf("BService should be missing AService dependency: %v", dependencyErrors)
	}

	app.Service().Stop()
}

func TestApplicationContext_CheckAllServiceDependencies_DependencyVersionConstraintsFail(t *testing.T) {
	app := service.NewApplication(service.ApplicationSettings{})
	app.Service().StartAsync()
	app.Service().AwaitUntilRunning()
	defer app.Service().Stop()

	bService := app.MustRegisterService(BServiceClientConstructor)

	// serviceB depends on serviceA version >= 1.0, < 2
	serviceA := app.MustRegisterService(AServiceClientConstructorFactory("0.9.0"))
	serviceA.Service().AwaitUntilRunning()

	dependencyErrors := app.CheckAllServiceDependencies()
	if len(dependencyErrors.Errors) != 2 {
		t.Errorf("BService should be missing AService dependency - it should not be registered and thus not running: %v", dependencyErrors)
	}

	if !app.UnRegisterService(serviceA) {
		t.Error("Failed to unregister serviceA")
	}

	// register compatible serviceA version
	serviceA = app.MustRegisterService(AServiceClientConstructorFactory("1.9.0"))
	serviceA.Service().AwaitUntilRunning()
	bService.Service().AwaitUntilRunning()
	if dependencyErrors = app.CheckAllServiceDependencies(); dependencyErrors != nil {
		t.Errorf("There should be no services dependency errors : %v", dependencyErrors)
	}
	if dependencyErrors := app.CheckServiceDependencies(bService); dependencyErrors != nil {
		t.Errorf("There should be no services dependency errors : %v", dependencyErrors)
	}

	serviceA.Service().Stop()
	dependencyErrors = app.CheckAllServiceDependencies()
	if len(dependencyErrors.Errors) != 1 {
		t.Errorf("AService dependency is not running: %v", dependencyErrors)
	}

	serviceA.Restart()
	serviceA.Service().AwaitUntilRunning()
	dependencyErrors = app.CheckAllServiceDependencies()
	if dependencyErrors = app.CheckAllServiceDependencies(); dependencyErrors != nil {
		t.Errorf("There should be no services dependency errors : %v", dependencyErrors)
	}
}

func TestApplicationContext_CheckAllServiceDependencies(t *testing.T) {
	app := service.NewApplication(service.ApplicationSettings{})
	app.Service().StartAsync()
	app.Service().AwaitUntilRunning()
	defer app.Service().Stop()

	app.MustRegisterService(EchoServiceClientConstructor)

	dependencyErrors := app.CheckAllServiceDependencies()
	if dependencyErrors != nil {
		t.Errorf("There should be no services missing dependencies : %v", dependencyErrors)
	}

	bService := app.MustRegisterService(BServiceClientConstructor)

	dependencyErrors = app.CheckAllServiceDependencies()
	t.Log(dependencyErrors)
	if len(dependencyErrors.Errors) != 2 {
		t.Errorf("BService should be missing AService dependency - it should not be registered and thus not running: %v", dependencyErrors)
	}

	checkServiceADependencyErrors := func(err *service.ServiceDependencyErrors) {
		t.Helper()
		if len(err.Errors) != 2 {
			t.Errorf("There shold be 2 missing Dependencies : %v", err.Errors)
		}
		for _, dependencyError := range err.Errors {
			switch e := dependencyError.(type) {
			case *service.ServiceDependenciesMissing:
				if !e.Missing(AServiceInterface) {
					t.Error("AServiceShould not be registered")
				}
			case *service.ServiceDependenciesNotRunning:
				if !e.NotRunning(AServiceInterface) {
					t.Error("AServiceShould not be running")
				}
			default:
				t.Errorf("Unexpected error type : %[0]T : %[0]v", err)
			}
		}
	}

	checkServiceADependencyErrors(dependencyErrors)
	checkServiceADependencyErrors(app.CheckServiceDependencies(bService))

	serviceA := app.MustRegisterService(AServiceClientConstructorFactory("1.9.0"))
	serviceA.Service().AwaitUntilRunning()
	bService.Service().AwaitUntilRunning()
	if dependencyErrors = app.CheckAllServiceDependencies(); dependencyErrors != nil {
		t.Errorf("There should be no services dependency errors : %v", dependencyErrors)
	}
	if dependencyErrors := app.CheckServiceDependencies(bService); dependencyErrors != nil {
		t.Errorf("There should be no services dependency errors : %v", dependencyErrors)
	}

	serviceA.Service().Stop()
	dependencyErrors = app.CheckAllServiceDependencies()
	if len(dependencyErrors.Errors) != 1 {
		t.Errorf("AService dependency is not running: %v", dependencyErrors)
	}

	serviceA.Restart()
	serviceA.Service().AwaitUntilRunning()
	dependencyErrors = app.CheckAllServiceDependencies()
	if dependencyErrors = app.CheckAllServiceDependencies(); dependencyErrors != nil {
		t.Errorf("There should be no services dependency errors : %v", dependencyErrors)
	}
}

func TestApplicationContext_CheckAllServiceDependenciesRunning(t *testing.T) {
	app := service.NewApplication(service.ApplicationSettings{})
	app.Service().StartAsync()
	app.Service().AwaitUntilRunning()
	defer app.Service().Stop()

	app.MustRegisterService(EchoServiceClientConstructor)
	app.MustRegisterService(HeartbeatServiceClientConstructor)

	notRunning := app.CheckAllServiceDependenciesRunning()
	if len(notRunning) > 0 {
		t.Errorf("There should be no services missing dependencies : %v", notRunning)
	}

	bService := app.MustRegisterService(BServiceClientConstructor)

	notRunning = app.CheckAllServiceDependenciesRunning()
	t.Log(notRunning)
	if len(notRunning) != 1 {
		t.Errorf("BService should be missing AService dependency: %v", notRunning)
	}

	checkMissingDepencencyServiceA := func(err *service.ServiceDependenciesNotRunning) {
		if err.ServiceInterface != bService.Service().Interface() {
			t.Errorf("ServiceInterface should be: %v , but was %v", bService.Service().Interface(), err.ServiceInterface)
		}
		if len(err.Dependencies) != 1 {
			t.Errorf("There shold be i missing Dependencies : %v", err.Dependencies)
		}
		if err.Dependencies[0] != AServiceInterface {
			t.Errorf("Missing service dependency should be AService, but was : %v", err.Dependencies[0])
		}
	}

	for _, err := range notRunning {
		checkMissingDepencencyServiceA(err)
	}

	err := app.CheckServiceDependenciesRunning(bService)
	if err == nil {
		t.Error("BService should be missing AService dependency")
	}
	checkMissingDepencencyServiceA(err)

	bService.Service().AwaitRunning(time.Duration(5 * time.Millisecond))
	if !bService.Service().State().Starting() {
		t.Errorf("BService should be blocked while starting waiting on reference to A : %v", bService.Service().State())
	}

	aService := app.MustRegisterService(AServiceClientConstructorFactory("1.0.0"))
	aService.Service().AwaitUntilRunning()
	bService.Service().AwaitUntilRunning()

	notRunning = app.CheckAllServiceDependenciesRunning()
	if len(notRunning) > 0 {
		t.Errorf("There should be no services missing dependencies : %v", notRunning)
	}

	aService.Service().Stop()

	notRunning = app.CheckAllServiceDependenciesRunning()
	if len(notRunning) == 0 {
		t.Errorf("AService should not be running : %v", notRunning)
	}
}

func TestApplication_ServiceStates(t *testing.T) {
	app := service.NewApplication(service.ApplicationSettings{})
	app.Service().StartAsync()
	app.Service().AwaitUntilRunning()
	defer app.Service().Stop()

	serviceStates := app.ServiceStates()
	if len(serviceStates) != 0 {
		t.Errorf("there should be no services registered : %v", serviceStates)
	}

	echoService := app.MustRegisterService(EchoServiceClientConstructor)
	serviceStates = app.ServiceStates()
	t.Logf("serviceStates: %v", serviceStates)
	if len(serviceStates) != 1 {
		t.Errorf("there should be 1 service registered : %v", serviceStates)
	}

	echoService.Service().AwaitUntilRunning()
	serviceStates = app.ServiceStates()
	if len(serviceStates) != 1 {
		t.Errorf("there should be 1 service registered : %v", serviceStates)
	}
	if serviceStates[EchoServiceInterface] != echoService.Service().State() {
		t.Errorf("EchoService should be running: %v", serviceStates)
	}

	heartbeatService := app.MustRegisterService(HeartbeatServiceClientConstructor)
	serviceStates = app.ServiceStates()
	if len(serviceStates) != 2 {
		t.Errorf("there should be 1 service registered : %v", serviceStates)
	}
	heartbeatService.Service().AwaitUntilRunning()
	serviceStates = app.ServiceStates()
	if len(serviceStates) != 2 {
		t.Errorf("there should be 2 services registered : %v", serviceStates)
	}
	if !serviceStates[EchoServiceInterface].Running() {
		t.Errorf("EchoService should be running: %v", serviceStates)
	}
	if !serviceStates[HeartbeatServiceInterface].Running() {
		t.Errorf("HeartbeatService should be running: %v", serviceStates)
	}
	app.Service().Stop()

	serviceStates = app.ServiceStates()
	if !serviceStates[EchoServiceInterface].Stopped() {
		t.Errorf("EchoService should be stopped: %v", serviceStates)
	}
	if !serviceStates[HeartbeatServiceInterface].Stopped() {
		t.Errorf("HeartbeatService should be stopped: %v", serviceStates)
	}
}

func TestApplication_StopRestartServices(t *testing.T) {
	app := service.NewApplication(service.ApplicationSettings{})
	app.Service().StartAsync()
	app.Service().AwaitUntilRunning()
	defer app.Service().Stop()

	echoService := app.MustRegisterService(EchoServiceClientConstructor)
	echoService.Service().AwaitUntilRunning()

	heartbeatService := app.MustRegisterService(HeartbeatServiceClientConstructor)
	heartbeatService.Service().AwaitUntilRunning()

	app.RestartServiceByType(EchoServiceInterface)
	for {
		if echoService.RestartCount() == 1 && heartbeatService.RestartCount() == 0 {
			break
		}

		t.Logf("Waiting for services to restart ...")
		time.Sleep(1 * time.Millisecond)
	}

	app.RestartServiceByKey(service.InterfaceTypeToServiceKey(HeartbeatServiceInterface))
	for {
		if echoService.RestartCount() == 1 && heartbeatService.RestartCount() == 1 {
			break
		}

		t.Logf("Waiting for services to restart ...")
		time.Sleep(1 * time.Millisecond)
	}

	if err := app.StopServiceByType(EchoServiceInterface); err != nil {
		t.Error(err)
	}
	// stopping a stopped service should be ok
	if err := app.StopServiceByType(EchoServiceInterface); err != nil {
		t.Error(err)
	}
	echoService.Service().AwaitUntilStopped()

	if err := app.StopServiceByKey(service.InterfaceTypeToServiceKey(HeartbeatServiceInterface)); err != nil {
		t.Error(err)
	}
	// stopping a stopped service should be ok
	if err := app.StopServiceByKey(service.InterfaceTypeToServiceKey(HeartbeatServiceInterface)); err != nil {
		t.Error(err)
	}
	heartbeatService.Service().AwaitUntilStopped()

	app.StopAllServices()
	app.RestartAllServices()

	for {
		if echoService.RestartCount() == 2 && heartbeatService.RestartCount() == 2 {
			break
		}

		t.Logf("Waiting for services to restart ...")
		time.Sleep(1 * time.Millisecond)
	}
}

func TestApplication_RestartAllServices(t *testing.T) {
	app := service.NewApplication(service.ApplicationSettings{})
	app.Service().StartAsync()
	app.Service().AwaitUntilRunning()
	defer app.Service().Stop()

	echoService := app.MustRegisterService(EchoServiceClientConstructor)
	echoService.Service().AwaitUntilRunning()

	heartbeatService := app.MustRegisterService(HeartbeatServiceClientConstructor)
	heartbeatService.Service().AwaitUntilRunning()

	app.RestartAllServices()

	for {
		if echoService.RestartCount() == 1 && heartbeatService.RestartCount() == 1 {
			break
		}

		t.Logf("Waiting for services to restart ...")
		time.Sleep(1 * time.Millisecond)
	}

	app.RestartAllServices()

	for {
		if echoService.RestartCount() == 2 && heartbeatService.RestartCount() == 2 {
			break
		}

		t.Logf("Waiting for services to restart ...")
		time.Sleep(1 * time.Millisecond)
	}
}

func TestApplication_StopRestartServices_OnEmptyApp(t *testing.T) {
	app := service.NewApplication(service.ApplicationSettings{})
	app.Service().StartAsync()
	app.Service().AwaitUntilRunning()
	defer app.Service().Stop()

	checkErrorTypeIsServiceNotFoundError := func(err error) {
		t.Helper()
		switch err.(type) {
		case *service.ServiceNotFoundError:
		default:
			t.Errorf("ERROR: Expected ServiceNotFoundError, but was %v", err)
		}
	}

	// should be ok to call when no services are registered
	app.RestartAllServices()
	app.RestartAllFailedServices()
	app.StopAllServices()
	if err := app.RestartServiceByType(EchoServiceInterface); err != nil {
		checkErrorTypeIsServiceNotFoundError(err)
	} else {
		t.Error("ERROR: no services are registered")
	}
	app.RestartServiceByKey(service.InterfaceTypeToServiceKey(EchoServiceInterface))
	if err := app.StopServiceByType(EchoServiceInterface); err != nil {
		checkErrorTypeIsServiceNotFoundError(err)
	} else {
		t.Error("ERROR: no services are registered")
	}
	if err := app.StopServiceByKey(service.InterfaceTypeToServiceKey(EchoServiceInterface)); err != nil {
		checkErrorTypeIsServiceNotFoundError(err)
	} else {
		t.Error("ERROR: no services are registered")
	}
}

func TestApplication_RestartAllFailedServices(t *testing.T) {
	app := service.NewApplication(service.ApplicationSettings{})
	app.Service().StartAsync()
	app.Service().AwaitUntilRunning()
	defer app.Service().Stop()

	echoService := app.MustRegisterService(EchoServiceClientConstructor)
	echoService.Service().AwaitUntilRunning()

	client := app.MustRegisterService(func(app service.Application) service.Client {
		client := &testApplication_RestartAllFailedServices_client{fail: make(chan struct{})}
		client.RestartableService = service.NewRestartableService(client.newService)
		return client
	})
	client.Service().AwaitUntilRunning()

	client.(testApplication_RestartAllFailedServices).Fail()
	client.Service().AwaitUntilStopped()
	app.RestartAllFailedServices()
	client.Service().AwaitUntilRunning()

	for {
		if c := client.RestartCount(); c == 1 {
			break
		}
		t.Log("Waiting for restart count to update")
		time.Sleep(time.Millisecond)
	}

}

type testApplication_RestartAllFailedServices interface {
	Fail()
}

type testApplication_RestartAllFailedServices_client struct {
	*service.RestartableService

	fail chan struct{}
}

func (a *testApplication_RestartAllFailedServices_client) Fail() {
	a.fail <- struct{}{}
}

func (a *testApplication_RestartAllFailedServices_client) run(ctx *service.Context) error {
	for {
		select {
		case <-ctx.StopTrigger():
			return nil
		case <-a.fail:
			panic("BOOM !!!")
		}
	}
}

func (a *testApplication_RestartAllFailedServices_client) newService() service.Service {
	return service.NewService(service.ServiceSettings{ServiceInterface: testApplication_RestartAllFailedServices_interface, Version: service.NewVersion("1.0.0"), Run: a.run})
}

var testApplication_RestartAllFailedServices_interface service.ServiceInterface = func() service.ServiceInterface {
	var o testApplication_RestartAllFailedServices = &testApplication_RestartAllFailedServices_client{}
	i, _ := reflect.ObjectInterface(&o)
	return i
}()
