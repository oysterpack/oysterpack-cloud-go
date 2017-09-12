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
	"github.com/oysterpack/oysterpack.go/oysterpack/commons"
	"time"
)

// ServiceManager is used to start/stop services
// TODO: Add to Application interface
type ServiceManager interface {
	ServiceStates() []ApplicationServiceState

	// Stop and start the service
	// returns ServiceNotFoundError if the service is not registered
	// returns ServiceError if the service restart failed
	RestartServiceByType(serviceInterface commons.InterfaceType) error
	RestartServiceByKey(key ServiceKey) error

	RestartAllServices() []ApplicationServiceState

	// Stop the service
	// returns ServiceNotFoundError if the service is not registered
	// returns ServiceError if the service stopped in a failed state
	StopServiceByType(serviceInterface commons.InterfaceType) error
	StopServiceByKey(key ServiceKey) error

	// Start the service
	// returns ServiceNotFoundError if the service is not registered
	// returns ServiceError if the service failed to start
	// returns IllegalStateError if the service is not in a NEW state
	StartServiceByType(serviceInterface commons.InterfaceType) error
	StartServiceByKey(key ServiceKey) error
}

// ApplicationServiceState contains the state for a registered application service at a point in time
type ApplicationServiceState struct {
	ServiceKey
	State
	time.Time
}
