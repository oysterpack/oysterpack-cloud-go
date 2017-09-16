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

// StopRestartServices is used to stop/restart services
// all operations are async, i.e., they initiate the action, but do not block to wait for the operation to complete
type StopRestartServices interface {

	// returns ServiceNotFoundError if the service is not registered
	RestartServiceByType(serviceInterface ServiceInterface) error
	RestartServiceByKey(key ServiceKey) error

	RestartAllServices()
	RestartAllFailedServices()

	StopAllServices()

	// returns ServiceNotFoundError if the service is not registered
	// returns ServiceError if the service stopped in a failed state
	StopServiceByType(serviceInterface ServiceInterface) error
	StopServiceByKey(key ServiceKey) error
}
