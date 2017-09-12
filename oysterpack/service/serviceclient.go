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

// ServiceClient represents the client interface. If follows the client-server model.
// The client must implement the service interface - this is verified at runtime when registered with the Application.
//
// The client instance is stable, meaning when retrieved from the Application, the same instance is returned.
// The backend service instance is not stable - meaning the instance may change. The backend service may be restarted,
// which means a new service instance will be created. The client and server instances are decoupled using channels to
// communicate. The client owns the channels.
//
// For restart functionality, RestartableService can be leveraged to simplify the implementation.
type ServiceClient interface {
	Service() *Service

	// restarts the service backend
	RestartService()
}

// ServiceClientConstructor is a service factory function that will construct a new Service instance and return a pointer to it.
// The returned service will be in a New state.
type ServiceClientConstructor func() ServiceClient
