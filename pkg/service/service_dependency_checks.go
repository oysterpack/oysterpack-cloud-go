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

// ServiceDependencyChecks groups methods that are used to check service dependencies
type ServiceDependencyChecks interface {
	// TODO: add healthcheck metric gauges and alerts for dependency checks

	// CheckAllServiceDependencies checks that all Dependencies for each service are available
	CheckAllServiceDependencies() *ServiceDependencyErrors

	// CheckAllServiceDependenciesRegistered checks that all Dependencies for each service are registered
	CheckAllServiceDependenciesRegistered() []*ServiceDependenciesMissing

	//CheckAllServiceDependenciesRunning checks that all Dependencies for each service are running
	CheckAllServiceDependenciesRunning() []*ServiceDependenciesNotRunning

	// CheckServiceDependencies checks that the service Dependencies are available for the specified service client
	CheckServiceDependencies(client Client) *ServiceDependencyErrors

	// CheckServiceDependenciesRegistered checks that the service Dependencies are registered for the specified service client
	CheckServiceDependenciesRegistered(client Client) *ServiceDependenciesMissing

	// CheckServiceDependenciesRunning checks that the service Dependencies are running for the specified service client
	CheckServiceDependenciesRunning(client Client) *ServiceDependenciesNotRunning
}

type serviceDependencyChecks struct {
	*application
}

// CheckAllServiceDependenciesRegistered checks that are service Dependencies are currently satisfied.
func (a serviceDependencyChecks) CheckAllServiceDependenciesRegistered() []*ServiceDependenciesMissing {
	a.registry.RLock()
	defer a.registry.RUnlock()
	errors := []*ServiceDependenciesMissing{}
	for _, serviceClient := range a.registry.services {
		if missingDependencies := a.checkServiceDependenciesRegistered(serviceClient); missingDependencies != nil {
			errors = append(errors, missingDependencies)
		}
	}
	return errors
}

// CheckServiceDependenciesRegistered checks that the service's Dependencies are currently satisfied
// nil is returned if there is no error
func (a serviceDependencyChecks) CheckServiceDependenciesRegistered(serviceClient Client) *ServiceDependenciesMissing {
	a.registry.RLock()
	defer a.registry.RUnlock()
	return a.checkServiceDependenciesRegistered(serviceClient)
}

func (a serviceDependencyChecks) checkServiceDependenciesRegistered(serviceClient Client) *ServiceDependenciesMissing {
	missingDependencies := &ServiceDependenciesMissing{&DependencyMappings{Interface: serviceClient.Service().Desc().Interface()}}
	for dependency, constraints := range serviceClient.Service().Dependencies() {
		b := a.serviceByType(dependency)
		if b == nil {
			missingDependencies.AddMissingDependency(dependency)
		} else {
			if constraints != nil {
				if !constraints.Check(b.Service().Desc().Version()) {
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

// CheckAllServiceDependenciesRunning checks that are service Dependencies are currently satisfied.
func (a serviceDependencyChecks) CheckAllServiceDependenciesRunning() []*ServiceDependenciesNotRunning {
	a.registry.RLock()
	defer a.registry.RUnlock()
	errors := []*ServiceDependenciesNotRunning{}
	for _, serviceClient := range a.registry.services {
		if notRunning := a.checkServiceDependenciesRunning(serviceClient); notRunning != nil {
			errors = append(errors, notRunning)
		}
	}
	return errors
}

// CheckServiceDependenciesRunning checks that the service's Dependencies are currently satisfied
// nil is returned if there is no error
func (a serviceDependencyChecks) CheckServiceDependenciesRunning(serviceClient Client) *ServiceDependenciesNotRunning {
	a.registry.RLock()
	defer a.registry.RUnlock()
	return a.checkServiceDependenciesRunning(serviceClient)
}

func (a serviceDependencyChecks) checkServiceDependenciesRunning(serviceClient Client) *ServiceDependenciesNotRunning {
	notRunning := &ServiceDependenciesNotRunning{&DependencyMappings{Interface: serviceClient.Service().Desc().Interface()}}
	for dependency, constraints := range serviceClient.Service().Dependencies() {
		if client := a.serviceByType(dependency); client == nil ||
			(constraints != nil && !constraints.Check(client.Service().Desc().Version())) ||
			!client.Service().State().Running() {
			notRunning.AddDependencyNotRunning(dependency)
		}
	}
	if notRunning.HasNotRunning() {
		return notRunning
	}
	return nil
}

// CheckAllServiceDependencies checks that all Dependencies for each service are available
// nil is returned if there are no errors
func (a serviceDependencyChecks) CheckAllServiceDependencies() *ServiceDependencyErrors {
	errors := []error{}
	for _, err := range a.CheckAllServiceDependenciesRegistered() {
		errors = append(errors, err)
	}

	for _, err := range a.CheckAllServiceDependenciesRunning() {
		errors = append(errors, err)
	}

	if len(errors) > 0 {
		return &ServiceDependencyErrors{errors}
	}
	return nil
}

// CheckServiceDependencies checks that the service Dependencies are available
func (a serviceDependencyChecks) CheckServiceDependencies(client Client) *ServiceDependencyErrors {
	errors := []error{}
	if err := a.CheckServiceDependenciesRegistered(client); err != nil {
		errors = append(errors, err)
	}

	if err := a.CheckServiceDependenciesRunning(client); err != nil {
		errors = append(errors, err)
	}

	if len(errors) > 0 {
		return &ServiceDependencyErrors{errors}
	}
	return nil
}
