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
	"github.com/oysterpack/oysterpack.go/oysterpack/commons"
	"github.com/oysterpack/oysterpack.go/oysterpack/service"
	"reflect"
	"sync"
	"testing"
)

func TestApplicationContext_RegisterService(t *testing.T) {
	app := service.NewApplicationContext()
	app.Service().StartAsync()
	app.Service().AwaitUntilRunning()
	defer app.Service().Stop()
	serviceClient := app.RegisterService(EchoServiceClientConstructor)
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

	serviceClients := []service.ServiceClient{
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
	app := service.NewApplicationContext()
	app.Service().StartAsync()
	app.Service().AwaitUntilRunning()
	defer app.Service().Stop()
	serviceClient := app.RegisterService(EchoServiceClientConstructor)
	message := "CIAO MUNDO !!!"

	serviceClients := []service.ServiceClient{
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
	app := service.NewApplicationContext()
	app.Service().StartAsync()
	app.Service().AwaitUntilRunning()
	defer app.Service().Stop()
	serviceClient := app.RegisterService(EchoServiceClientConstructor)
	if app.RegisterService(EchoServiceClientConstructor) != nil {
		t.Errorf("nil should have been returned because the service is already registered")
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
	serviceClient2 := app.RegisterService(EchoServiceClientConstructor)
	if serviceClient2 == nil {
		t.Errorf("service should have been registered")
	}
	if serviceClient == serviceClient2 {
		t.Errorf("A new ServiceClient instance should have been returned")
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
	app := service.NewApplicationContext()
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
				t.Errorf("Application.RegisterService should have panicked")
			}
		}()

		app.RegisterService(func() service.ServiceClient {
			return &InvalidEchoServiceClient{
				RestartableService: service.NewRestartableService(func() *service.Service {
					var svc EchoService = &SimpleEchoService{}
					serviceInterface, _ := commons.ObjectInterface(&svc)
					return service.NewService(serviceInterface, nil, nil, nil)
				}),
			}
		})
	}()

}

func TestApplicationContext_ServiceByType_NotRegistered(t *testing.T) {
	app := service.NewApplicationContext()
	app.Service().StartAsync()
	app.Service().AwaitUntilRunning()
	defer app.Service().Stop()

	if app.ServiceByType(EchoServiceInterface) != nil {
		t.Error("ERROR: no services are registered")
	}
	if app.ServiceByKey(service.InterfaceTypeToServiceKey(EchoServiceInterface)) != nil {
		t.Error("ERROR: no services are registered")
	}

	app.RegisterService(EchoServiceClientConstructor)

	serviceClients := []service.ServiceClient{
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
	fooType, err := commons.ObjectInterface(&foo)
	if err != nil {
		t.Fatalf("Failed to get Foo type : %v", err)
	}
	if app.ServiceByType(fooType) != nil {
		t.Error("ERROR: no Foo service is registered")
	}
}

func TestApplicationContext_ServiceInterfaces(t *testing.T) {
	app := service.NewApplicationContext()
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

	echoService := app.RegisterService(EchoServiceClientConstructor).(EchoService)
	serviceInterfaces = app.ServiceInterfaces()
	if app.ServiceCount() != 1 {
		t.Errorf("ERROR: there should be 1 services registered : %d : %v", len(serviceInterfaces), serviceInterfaces)
	}
	if len(serviceInterfaces) != 1 {
		t.Errorf("ERROR: there should be 1 services registered : %d : %d : %v", app.ServiceCount(), len(serviceInterfaces), serviceInterfaces)
	}
	t.Logf("echo : %v", echoService.Echo("TestApplicationContext_ServiceInterfaces"))
	heartbeatService := app.RegisterService(HeartbeatServiceClientConstructor).(HeartbeatService)
	t.Logf("heartbeat ping : %v", heartbeatService.Ping())

	serviceInterfaces = app.ServiceInterfaces()
	t.Logf("serviceInterfaces : %v", serviceInterfaces)
	if app.ServiceCount() != 2 {
		t.Errorf("ERROR: there should be 2 services registered : %d : %v", len(serviceInterfaces), serviceInterfaces)
	}
	if len(serviceInterfaces) != 2 {
		t.Errorf("ERROR: there should be 2 services registered : %d : %d : %v", app.ServiceCount(), len(serviceInterfaces), serviceInterfaces)
	}

	containsInterface := func(svcInterface commons.InterfaceType) bool {
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
	app := service.NewApplicationContext()
	app.Service().StartAsync()
	app.Service().AwaitUntilRunning()
	defer app.Service().Stop()

	app.RegisterService(EchoServiceClientConstructor)
	services := app.Services()
	if len(services) != 1 {
		t.Errorf("there should be 1 service registered : %v", services)
	}

	app.RegisterService(HeartbeatServiceClientConstructor)
	services = app.Services()

	if len(services) != 2 {
		t.Errorf("ERROR: there should be 2 services registered : %v", services)
	}

	containsService := func(svcInterface commons.InterfaceType) bool {
		for _, service := range services {
			if reflect.TypeOf(service).AssignableTo(svcInterface) {
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
	app := service.NewApplicationContext()
	app.Service().StartAsync()
	app.Service().AwaitUntilRunning()
	defer app.Service().Stop()

	serviceKeys := app.ServiceKeys()
	if len(serviceKeys) != 0 {
		t.Errorf("ERROR: there should be 0 services registered : %d : %d : %v", app.ServiceCount(), len(serviceKeys), serviceKeys)
	}

	app.RegisterService(EchoServiceClientConstructor)
	serviceKeys = app.ServiceKeys()
	if len(serviceKeys) != 1 {
		t.Errorf("ERROR: there should be 1 services registered : %v", serviceKeys)
	}
	app.RegisterService(HeartbeatServiceClientConstructor)
	serviceKeys = app.ServiceKeys()
	t.Logf("serviceKeys : %v", serviceKeys)
	if len(serviceKeys) != 2 {
		t.Errorf("ERROR: there should be 2 services registered : %d : %d : %v", app.ServiceCount(), len(serviceKeys), serviceKeys)
	}

	containsInterface := func(svcInterface commons.InterfaceType) bool {
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
	app := service.NewApplicationContext()
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
		t.Log(echoService.Echo("service ticket has been fullfilled"))
	}()
	if app.ServiceTicketCounts()[EchoServiceInterface] != 1 {
		t.Errorf("There should be ticket in the queue for the EchoService : %d", app.ServiceTicketCounts()[EchoServiceInterface])
	}
	app.RegisterService(EchoServiceClientConstructor)
	wait.Wait()

	serviceTicket = app.ServiceByTypeAsync(EchoServiceInterface)
	serviceClient := <-serviceTicket.Channel()
	echoService := serviceClient.(EchoService)
	t.Log(echoService.Echo("service ticket has been fullfilled"))

	const count = 100
	wait.Add(count)
	for i := 0; i < count; i++ {
		go func(index int, serviceTicket *service.ServiceTicket) {
			defer wait.Done()
			serviceClient := <-serviceTicket.Channel()
			echoService := serviceClient.(EchoService)
			t.Logf("#%d : %s", index, echoService.Echo("service ticket has been fullfilled"))
		}(i, app.ServiceByTypeAsync(EchoServiceInterface))
	}
	wait.Wait()
}
