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
	"strings"

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

type serviceDependencies struct {
	serviceInterface commons.InterfaceType
	dependencies     []commons.InterfaceType
}

func (a *serviceDependencies) Dependencies() []commons.InterfaceType {
	return a.dependencies
}

func (a *serviceDependencies) ServiceInterface() commons.InterfaceType {
	return a.serviceInterface
}

// AddDependency will add the missing dependency, if it has not yet already been added
func (a *serviceDependencies) addDependency(dependency commons.InterfaceType) {
	if !a.contains(dependency) {
		a.dependencies = append(a.dependencies, dependency)
	}
}

// Contains returns true if the service dependency exists
func (a *serviceDependencies) contains(dependency commons.InterfaceType) bool {
	for _, v := range a.dependencies {
		if v == dependency {
			return true
		}
	}
	return false
}

func (a *serviceDependencies) String() string {
	return fmt.Sprintf("%v -> %v", a.ServiceInterface(), a.dependencies)
}

// ServiceDependenciesMissing indicates that a service's dependencies are missing at runtime
type ServiceDependenciesMissing struct {
	*serviceDependencies
}

func (a *ServiceDependenciesMissing) Error() string {
	return fmt.Sprintf("Service dependencies are missing : %v", a.serviceDependencies)
}

// AddMissingDependency will add the missing dependency, if it has not yet already been added
func (a *ServiceDependenciesMissing) AddMissingDependency(dependency commons.InterfaceType) {
	a.addDependency(dependency)
}

// Missing returns true if the service is missing the specified dependency
func (a *ServiceDependenciesMissing) Missing(dependency commons.InterfaceType) bool {
	return a.contains(dependency)
}

// HasMissing returns true if the service has any missing dependencies
func (a *ServiceDependenciesMissing) HasMissing() bool {
	return len(a.dependencies) > 0
}

// ServiceDependenciesNotRunning indicates that a service's dependencies are registered, but not running
type ServiceDependenciesNotRunning struct {
	*serviceDependencies
}

func (a *ServiceDependenciesNotRunning) Error() string {
	return fmt.Sprintf("Service dependencies are not running : %v", a.serviceDependencies)
}

// AddDependencyNotRunning will add the missing dependency, if it has not yet already been added
func (a *ServiceDependenciesNotRunning) AddDependencyNotRunning(dependency commons.InterfaceType) {
	a.addDependency(dependency)
}

// NotRunning returns true if the service is missing the specified dependency
func (a *ServiceDependenciesNotRunning) NotRunning(dependency commons.InterfaceType) bool {
	return a.contains(dependency)
}

// HasNotRunning returns true if the service has any missing dependencies
func (a *ServiceDependenciesNotRunning) HasNotRunning() bool {
	return len(a.dependencies) > 0
}

// ServiceDependencyErrors aggregates service dependency related errors. The types of errors are :
// 1. ServiceDependenciesMissing
// 2. ServiceDependenciesNotRunning
type ServiceDependencyErrors struct {
	Errors []error
}

func (a *ServiceDependencyErrors) Error() string {
	errorMessages := make([]string, len(a.Errors))
	for i, v := range a.Errors {
		errorMessages[i] = v.Error()
	}
	return fmt.Sprintf("Error count = %d : %v", len(errorMessages), strings.Join(errorMessages, " | "))
}

// HasErrors returns true if there are any dependency related errors
func (a *ServiceDependencyErrors) HasErrors() bool {
	return len(a.Errors) > 0
}

// ServiceDependenciesMissingErrors returns any ServiceDependenciesMissing errors
func (a *ServiceDependencyErrors) ServiceDependenciesMissingErrors() []*ServiceDependenciesMissing {
	errors := []*ServiceDependenciesMissing{}
	for _, err := range a.Errors {
		switch e := err.(type) {
		case *ServiceDependenciesMissing:
			errors = append(errors, e)
		}
	}
	return errors
}

// ServiceDependenciesNotRunningErrors returns any ServiceDependenciesNotRunning errors
func (a *ServiceDependencyErrors) ServiceDependenciesNotRunningErrors() []*ServiceDependenciesNotRunning {
	errors := []*ServiceDependenciesNotRunning{}
	for _, err := range a.Errors {
		switch e := err.(type) {
		case *ServiceDependenciesNotRunning:
			errors = append(errors, e)
		}
	}
	return errors
}

// DependencyErrors returns any dependency errors for the specified service interface
func (a *ServiceDependencyErrors) DependencyErrors(serviceInterface commons.InterfaceType) []error {
	errors := []error{}

	type GetServiceInterface interface {
		ServiceInterface() commons.InterfaceType
	}

	for _, err := range a.Errors {
		switch err.(type) {
		case GetServiceInterface:
			errors = append(errors, err)
		default:
			logger.Warn().Msgf("error does not implement GetServiceInterface : %v", err)
		}
	}
	return errors
}
