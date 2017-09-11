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
	"sync"
	"testing"
)

func TestApplicationContext_RegisterService(t *testing.T) {
	app := service.NewApplicationContext()
	app.Service().StartAsync()
	app.Service().AwaitUntilRunning()
	defer app.Service().Stop()
	serviceClient := app.RegisterService(EchoServiceConstructor)
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

	app.Service().Stop()
	if !serviceClient.Service().State().Stopped() {
		t.Error("service should be stopped")
	}
	serviceClient.RestartService()
	serviceClient.Service().AwaitUntilRunning()
	if response := echoService.Echo(message); response != message {
		t.Errorf("echo : return unexpected response : %v", response)
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
		t.Logf("echo after service restarted : %v", echoService.Echo(message))
		echoWaitGroup.Done()
	}()
	runningWaitGroup.Wait()
	serviceClient.RestartService()
	t.Log("service is being restarted ...")
	echoWaitGroup.Wait()
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

// EchoService defines the Service interface
type EchoService interface {
	Echo(msg interface{}) interface{}
}

// EchoServiceClient is the EchoService implementation
type EchoServiceClient struct {
	// embedded service
	*service.RestartableService

	// service channels
	echo chan *EchoRequest
}

////////// service messages //////////////////

type EchoRequest struct {
	Message interface{}
	ReplyTo chan interface{}
}

// message constructor
func NewEchoRequest(msg interface{}) *EchoRequest {
	return &EchoRequest{msg, make(chan interface{}, 1)}
}

////////// service methods //////////////////

// Echo
func (a *EchoServiceClient) Echo(msg interface{}) interface{} {
	req := NewEchoRequest(msg)
	a.echo <- req
	return <-req.ReplyTo
}

////////// service constructor //////////////////

// ServiceConstructor
func (a *EchoServiceClient) newService() *service.Service {
	var svc EchoService = a
	serviceInterface, _ := commons.ObjectInterface(&svc)
	return service.NewService(serviceInterface, nil, a.run, nil)
}

// Service Run func
func (a *EchoServiceClient) run(ctx *service.RunContext) error {
	for {
		select {
		case req := <-a.echo:
			go func() { req.ReplyTo <- req.Message }()
		case <-ctx.StopTrigger():
			return nil
		}
	}
}

// EchoServiceConstructor is the ServiceClientConstructor
func EchoServiceConstructor() service.ServiceClient {
	serviceClient := &EchoServiceClient{
		echo: make(chan *EchoRequest),
	}
	serviceClient.RestartableService = service.NewRestartableService(serviceClient.newService)
	return serviceClient
}
