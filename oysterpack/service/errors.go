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
	"fmt"
	"github.com/oysterpack/oysterpack.go/oysterpack/commons"
)

// InvalidStateTransition indicates an invalid transition was attempted
type InvalidStateTransition struct {
	From State
	To   State
}

func (e *InvalidStateTransition) Error() string {
	return fmt.Sprintf("InvalidStateTransition: %v -> %v", e.From, e.To)
}

// IllegalStateError indicates we are in an illegal state
type IllegalStateError struct {
	State
	Message string
}

func (e *IllegalStateError) Error() string {
	if e.Message == "" {
		return e.State.String()
	}
	return fmt.Sprintf("%v : %v", e.State, e.Message)
}

// UnknownFailureCause indicates that the service is in a Failed state, but the failure cause is unknown.
type UnknownFailureCause struct{}

func (e UnknownFailureCause) Error() string {
	return "UnknownFailureCause"
}

// PastStateError indicates that we currently in a state that is past the desired state
type PastStateError struct {
	Past    State
	Current State
}

func (e *PastStateError) Error() string {
	return fmt.Sprintf("Current state (%v) is past state (%v)", e.Current, e.Past)
}

// ServiceError contains the error and the state the service was in when the error occurred
type ServiceError struct {
	// State in which the error occurred
	State
	Err error
}

func (e *ServiceError) Error() string {
	return fmt.Sprintf("%v : %v", e.State, e.Err)
}

// PanicError is used to wrap any trapped panics along with a supplemental info about the context of the panic
type PanicError struct {
	Panic interface{}
	// additional info
	Message string
}

func (e *PanicError) Error() string {
	if e.Message != "" {
		return fmt.Sprintf("panic: %v : %v", e.Panic, e.Message)
	}
	return fmt.Sprintf("panic: %v", e.Panic)
}

// ServiceNotFoundError occurs when the service is unknown.
type ServiceNotFoundError struct {
	ServiceKey
}

func (e *ServiceNotFoundError) Error() string {
	return fmt.Sprintf("Service not found : %v.%v", e.ServiceKey.PackagePath, e.ServiceKey.TypeName)
}

// ServiceDependenciesMissing indicates that a service's dependencies are missing at runtime
type ServiceDependenciesMissing struct {
	ServiceInterface    commons.InterfaceType
	ServiceDependencies []commons.InterfaceType
}

func (e *ServiceDependenciesMissing) Error() string {
	return fmt.Sprintf("Service dependencies are missing : %v -> %v", e.ServiceInterface, e.ServiceDependencies)
}

// AddMissingDependency will add the missing dependency, if it has not yet already been added
func (e *ServiceDependenciesMissing) AddMissingDependency(dependency commons.InterfaceType) {
	if !e.Missing(dependency) {
		e.ServiceDependencies = append(e.ServiceDependencies, dependency)
	}
}

func (e *ServiceDependenciesMissing) Missing(dependency commons.InterfaceType) bool {
	for _, v := range e.ServiceDependencies {
		if v == dependency {
			return true
		}
	}
	return false
}

func (e *ServiceDependenciesMissing) HasMissing() bool {
	return len(e.ServiceDependencies) > 0
}
