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
	"time"
)

// EchoService defines the Service interface
type EchoService interface {
	Echo(msg interface{}) interface{}
}

var EchoServiceInterface commons.InterfaceType = echoServiceInterfaceType()

func echoServiceInterfaceType() commons.InterfaceType {
	var echoServicePrototype EchoService = &EchoServiceClient{}
	t, err := commons.ObjectInterface(&echoServicePrototype)
	if err != nil {
		panic(err)
	}
	return t
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
	return service.NewService(service.NewServiceParams{ServiceInterface: EchoServiceInterface, Run: a.run})
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

// EchoServiceClientConstructor is a ServiceClientConstructor
func EchoServiceClientConstructor(application service.Application) service.ServiceClient {
	serviceClient := &EchoServiceClient{
		echo: make(chan *EchoRequest),
	}
	serviceClient.RestartableService = service.NewRestartableService(serviceClient.newService)
	return serviceClient
}

///////////////////////////

type HeartbeatService interface {
	Ping() time.Duration
}

var HeartbeatServiceInterface commons.InterfaceType = heartbeatServiceInterfaceType()

func heartbeatServiceInterfaceType() commons.InterfaceType {
	var prototype HeartbeatService = &HeartbeatServiceClient{}
	t, err := commons.ObjectInterface(&prototype)
	if err != nil {
		panic(err)
	}
	return t
}

type HeartbeatServiceClient struct {
	*service.RestartableService

	pingChan chan *PingRequest
}

type PingRequest struct {
	time.Time
	replyTo chan time.Duration
}

func (a *HeartbeatServiceClient) Ping() time.Duration {
	req := &PingRequest{time.Now(), make(chan time.Duration, 1)}
	a.pingChan <- req
	return <-req.replyTo
}

func (a *HeartbeatServiceClient) run(ctx *service.RunContext) error {
	for {
		select {
		case <-ctx.StopTrigger():
			return nil
		case req := <-a.pingChan:
			req.replyTo <- time.Since(req.Time)
		}
	}
}

func (a *HeartbeatServiceClient) newService() *service.Service {
	return service.NewService(service.NewServiceParams{ServiceInterface: HeartbeatServiceInterface, Run: a.run})
}

func HeartbeatServiceClientConstructor(application service.Application) service.ServiceClient {
	serviceClient := &HeartbeatServiceClient{
		pingChan: make(chan *PingRequest),
	}
	serviceClient.RestartableService = service.NewRestartableService(serviceClient.newService)
	return serviceClient
}

////////////////////////////

type AService interface{}

type AServiceClient struct {
	*service.RestartableService
}

var AServiceInterface commons.InterfaceType = aServiceInterfaceType()

func aServiceInterfaceType() commons.InterfaceType {
	var prototype AService = &AServiceClient{}
	t, err := commons.ObjectInterface(&prototype)
	if err != nil {
		panic(err)
	}
	return t
}

func AServiceClientConstructor(application service.Application) service.ServiceClient {
	serviceClient := &AServiceClient{}
	serviceClient.RestartableService = service.NewRestartableService(func() *service.Service {
		return service.NewService(service.NewServiceParams{ServiceInterface: AServiceInterface})
	})
	return serviceClient
}

////////////////////////////////

type BService interface{}

type BServiceClient struct {
	*service.RestartableService
}

var BServiceInterface commons.InterfaceType = bServiceInterfaceType()

func bServiceInterfaceType() commons.InterfaceType {
	var prototype BService = &BServiceClient{}
	t, err := commons.ObjectInterface(&prototype)
	if err != nil {
		panic(err)
	}
	return t
}

func BServiceClientConstructor(app service.Application) service.ServiceClient {
	serviceClient := &AServiceClient{}
	serviceClient.RestartableService = service.NewRestartableService(func() *service.Service {

		var init service.Init = func(ctx *service.Context) error {
			<-app.ServiceByTypeAsync(AServiceInterface).Channel()
			return nil
		}

		return service.NewService(service.NewServiceParams{
			ServiceInterface:    BServiceInterface,
			ServiceDependencies: []commons.InterfaceType{AServiceInterface},
			Init:                init,
		})
	})
	return serviceClient
}
