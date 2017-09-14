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

// ServiceClient represents the client interface. If follows the client-server design pattern.
// The ServiceClient instance must implement the reference service interface - this is verified at runtime when registered with the Application.
//
// The client instance is stable, meaning when retrieved from the Application, the same instance is returned.
// The backend service instance is not stable - meaning the instance may change over time, i.e. when restarted.
//
// The client and server instances are decoupled using channels to communicate. Each service interface method should use
// a typed channel to communicate with the "backend" service reference. The backend service's Run function becomes a message processor.
//
// ServiceClient Implementation Design Pattern:
// 1. define the service interface
// 2. create ServiceClient struct that implements the service interface
//    - embed RestartableService, which provides the ServiceReference and Restartable functionality
//    - for each service interface method, define a channel and corresponding request and response messages
//    - implement the service interface method using the channel to communicate with the backend service
// 3. implement service life cycle functions (Init, Run, Destroy) as needed
//    - Init
//      - lookup service dependencies
//      - register additional services
//      - initialize resources
//    - Run
//      - process service messages
//      - exit when shutdown is triggered
//    - Destroy
//      - clean up any resources created by the service
// 4. create a ServiceClientConstructor function
// 5. create a service commons.InterfaceType for the service interface - this will be used to lookup the service within the Application
type ServiceClient interface {
	ServiceReference

	Restartable
}

// ServiceClientConstructor is a service factory function that will construct a new Service instance and return a pointer to it.
// The returned service will be in a New state.
// The constructor is given the application that is creating the service client, i.e., the service client will be managed by the given application.
//
// Use cases for making the application available to the service client :
// 1. The service client uses the application to lookup service dependencies. Service dependencies should be obtained
//    during the service's init phase. Once the service client dependency reference is obtained, the service should await
//    until the service dependency is running during the service's run phase.
// 2. The service uses the application to register additional services.
type ServiceClientConstructor func(application Application) ServiceClient
