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

	"github.com/oysterpack/oysterpack.go/pkg/commons/reflect"
)

// CheckAllDependencies checks that all Dependencies for each  are available
// nil is returned if there are no errors
func CheckAllDependencies(container *Container) *DependencyErrors {
	errors := []error{}
	for _, err := range CheckAllDependenciesRegistered(container) {
		errors = append(errors, err)
	}

	for _, err := range CheckAllDependenciesRunning(container) {
		errors = append(errors, err)
	}

	if len(errors) > 0 {
		return &DependencyErrors{errors}
	}
	return nil
}

// CheckAllDependenciesRegistered checks that all Dependencies are currently satisfied.
func CheckAllDependenciesRegistered(container *Container) []*DependencyMissingError {
	errors := []*DependencyMissingError{}
	for _, comp := range container.comps {
		if missingDependencies := CheckDependenciesRegistered(container, comp); missingDependencies != nil {
			errors = append(errors, missingDependencies)
		}
	}
	return errors
}

// CheckAllDependenciesRunning checks that all Dependencies are currently satisfied.
func CheckAllDependenciesRunning(container *Container) []*DependencyNotRunningError {
	errors := []*DependencyNotRunningError{}
	for _, comp := range container.comps {
		if notRunning := CheckDependenciesRunning(container, comp); notRunning != nil {
			errors = append(errors, notRunning)
		}
	}
	return errors
}

// CheckDependencies checks that the all Dependencies are available
func CheckDependencies(container *Container, comp Component) *DependencyErrors {
	errors := []error{}
	if err := CheckDependenciesRegistered(container, comp); err != nil {
		errors = append(errors, err)
	}

	if err := CheckDependenciesRunning(container, comp); err != nil {
		errors = append(errors, err)
	}

	if len(errors) > 0 {
		return &DependencyErrors{errors}
	}
	return nil
}

// CheckDependenciesRegistered checks that the component's Dependencies are currently registered
// nil is returned if there is no error
func CheckDependenciesRegistered(container *Container, comp Component) *DependencyMissingError {
	missingDependencies := &DependencyMissingError{&DependencyMappings{Interface: comp.Interface()}}
	for dependency, constraints := range comp.Dependencies() {
		b, exists := container.comps[dependency]
		if !exists {
			missingDependencies.AddMissingDependency(dependency)
		} else {
			if constraints != nil {
				if !constraints.Check(b.Desc().Version()) {
					missingDependencies.AddMissingDependency(dependency)
				}
			}
		}
	}
	if missingDependencies.HasMissing() {
		return missingDependencies
	}
	return nil
}

// CheckDependenciesRunning checks that the component's Dependencies are currently running
// nil is returned if there is no error
func CheckDependenciesRunning(container *Container, comp Component) *DependencyNotRunningError {
	notRunning := &DependencyNotRunningError{&DependencyMappings{Interface: comp.Interface()}}
	for dependency, constraints := range comp.Dependencies() {
		if compDependency, exists := container.comps[dependency]; !exists ||
			(constraints != nil && !constraints.Check(compDependency.Desc().Version())) ||
			!compDependency.State().Running() {
			notRunning.AddDependencyNotRunning(dependency)
		}
	}
	if notRunning.HasNotRunning() {
		return notRunning
	}
	return nil
}

// DependencyMappings maps an Interface to its Interface dependencies
type DependencyMappings struct {
	Interface
	Dependencies []Interface
}

// AddDependency will add the missing dependency, if it has not yet already been added
func (a *DependencyMappings) addDependency(dependency reflect.InterfaceType) {
	if !a.contains(dependency) {
		a.Dependencies = append(a.Dependencies, dependency)
	}
}

// Contains returns true if the service dependency exists
func (a *DependencyMappings) contains(dependency reflect.InterfaceType) bool {
	for _, v := range a.Dependencies {
		if v == dependency {
			return true
		}
	}
	return false
}

func (a *DependencyMappings) String() string {
	return fmt.Sprintf("%v -> %v", a.Interface, a.Dependencies)
}
