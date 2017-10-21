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

// DependencyChecks groups methods that are used to check component dependencies
type DependencyChecks interface {

	// CheckAllDependencies checks that all Dependencies for each component are available
	CheckAllDependencies() *DependencyErrors

	// CheckAllDependenciesRegistered checks that all Dependencies for each component are registered
	CheckAllDependenciesRegistered() []*DependencyMissingError

	//CheckAllDependenciesRunning checks that all Dependencies for each component are running
	CheckAllDependenciesRunning() []*DependenciesNotRunningError

	// CheckDependencies checks that the service Dependencies are available for the specified component
	CheckDependencies(comp Component) *DependencyErrors

	// CheckDependenciesRegistered checks that the service Dependencies are registered for the specified component
	CheckDependenciesRegistered(comp Component) *DependencyMissingError

	// CheckDependenciesRunning checks that the service Dependencies are running for the specified component
	CheckDependenciesRunning(comp Component) *DependenciesNotRunningError
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
