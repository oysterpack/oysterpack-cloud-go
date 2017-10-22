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

package comp

import (
	"fmt"
	"strings"

	"github.com/oysterpack/oysterpack.go/pkg/commons/reflect"
)

// DependencyErrors aggregates service dependency related errors. The types of errors are :
// 1. DependencyMissingError
// 2. DependencyNotRunningError
type DependencyErrors struct {
	Errors []error
}

func (a *DependencyErrors) Error() string {
	errorMessages := make([]string, len(a.Errors))
	for i, v := range a.Errors {
		errorMessages[i] = v.Error()
	}
	return fmt.Sprintf("Error count = %d : %v", len(errorMessages), strings.Join(errorMessages, " | "))
}

// HasErrors returns true if there are any dependency related errors
func (a *DependencyErrors) HasErrors() bool {
	return len(a.Errors) > 0
}

// DependenciesMissingErrors returns any DependencyMissingError errors
func (a *DependencyErrors) DependencyMissingErrors() []*DependencyMissingError {
	errors := []*DependencyMissingError{}
	for _, err := range a.Errors {
		switch e := err.(type) {
		case *DependencyMissingError:
			errors = append(errors, e)
		}
	}
	return errors
}

// DependenciesNotRunningErrors returns any DependencyNotRunningError errors
func (a *DependencyErrors) DependencyNotRunningErrors() []*DependencyNotRunningError {
	errors := []*DependencyNotRunningError{}
	for _, err := range a.Errors {
		switch e := err.(type) {
		case *DependencyNotRunningError:
			errors = append(errors, e)
		}
	}
	return errors
}

// DependencyMissingError indicates that a service's Dependencies are missing at runtime
type DependencyMissingError struct {
	*DependencyMappings
}

func (a *DependencyMissingError) Error() string {
	return fmt.Sprintf("Service Dependencies are missing : %v", a.DependencyMappings)
}

// AddMissingDependency will add the missing dependency, if it has not yet already been added
func (a *DependencyMissingError) AddMissingDependency(dependency reflect.InterfaceType) {
	a.addDependency(dependency)
}

// Missing returns true if the service is missing the specified dependency
func (a *DependencyMissingError) Missing(dependency reflect.InterfaceType) bool {
	return a.contains(dependency)
}

// HasMissing returns true if the service has any missing Dependencies
func (a *DependencyMissingError) HasMissing() bool {
	return len(a.Dependencies) > 0
}

// DependencyNotRunningError indicates that a service's Dependencies are registered, but not running
type DependencyNotRunningError struct {
	*DependencyMappings
}

func (a *DependencyNotRunningError) Error() string {
	return fmt.Sprintf("Service Dependencies are not running : %v", a.DependencyMappings)
}

// AddDependencyNotRunning will add the missing dependency, if it has not yet already been added
func (a *DependencyNotRunningError) AddDependencyNotRunning(dependency reflect.InterfaceType) {
	a.addDependency(dependency)
}

// NotRunning returns true if the service is missing the specified dependency
func (a *DependencyNotRunningError) NotRunning(dependency reflect.InterfaceType) bool {
	return a.contains(dependency)
}

// HasNotRunning returns true if the service has any missing Dependencies
func (a *DependencyNotRunningError) HasNotRunning() bool {
	return len(a.Dependencies) > 0
}
