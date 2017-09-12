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

///////////////////////////

type HeartbeatService interface {
	Ping() time.Duration
}

var HeartbeatServiceInterface commons.InterfaceType = echoServiceInterfaceType()

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
	var svc HeartbeatService = a
	serviceInterface, _ := commons.ObjectInterface(&svc)
	return service.NewService(serviceInterface, nil, a.run, nil)
}

func HeartbeatServiceConstructor() service.ServiceClient {
	serviceClient := &HeartbeatServiceClient{
		pingChan: make(chan *PingRequest),
	}
	serviceClient.RestartableService = service.NewRestartableService(serviceClient.newService)
	return serviceClient
}
