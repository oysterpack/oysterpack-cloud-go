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
		t.Error(err)
	} else {
		if !serviceClient.Service().State().Running() {
			t.Error("service should be running")
		}
		echoService := serviceClient.(EchoService)
		t.Logf("echo : %v", echoService.Echo("CIAO MUNDO !!!"))
		app.Service().Stop()
		if !serviceClient.Service().State().Stopped() {
			t.Error("service should be stopped")
		}
	}
}

type EchoService interface {
	Echo(msg interface{}) interface{}
}

type EchoServiceClient struct {
	serviceMutex sync.RWMutex
	service      *service.Service

	echo chan *EchoRequest
}

type EchoRequest struct {
	Message interface{}
	ReplyTo chan interface{}
}

func NewEchoRequest(msg interface{}) *EchoRequest {
	return &EchoRequest{msg, make(chan interface{}, 1)}
}

func (a *EchoServiceClient) Echo(msg interface{}) interface{} {
	req := NewEchoRequest(msg)
	a.echo <- req
	return <-req.ReplyTo
}

func (a *EchoServiceClient) Service() *service.Service {
	a.serviceMutex.RLock()
	defer a.serviceMutex.RUnlock()
	return a.service
}

func (a *EchoServiceClient) RestartService() {
	a.serviceMutex.Lock()
	defer a.serviceMutex.Unlock()
	if a.service != nil {
		a.service.StopAsyc()
		a.service.AwaitUntilStopped()
	}
	a.service = a.newService()
	a.service.StartAsync()
}

func (a *EchoServiceClient) run(ctx *service.RunContext) error {
	for {
		select {
		case req := <-a.echo:
			req.ReplyTo <- req.Message
		case <-ctx.StopTrigger():
			return nil
		}
	}
}

func (a *EchoServiceClient) newService() *service.Service {
	var svc EchoService = a
	serviceInterface, _ := commons.ObjectInterface(&svc)
	return service.NewService(serviceInterface, nil, a.run, nil)
}

func EchoServiceConstructor() service.ServiceClient {
	serviceClient := &EchoServiceClient{
		echo: make(chan *EchoRequest),
	}
	serviceClient.service = serviceClient.newService()
	return serviceClient
}
