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
// Implementation : RestartableService
type ServiceClient interface {
	ServiceReference

	Restartable
}

// ServiceClientConstructor is a service factory function that will construct a new Service instance and return a pointer to it.
// The returned service will be in a New state.
type ServiceClientConstructor func() ServiceClient
