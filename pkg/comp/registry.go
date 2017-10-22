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

// Registry is used to register components and shutdown hooks
type Registry interface {

	// Register will create a new instance of the component using the supplied constructor.
	// The Component must implement the component Interface - otherwise the method panics.
	// It will then register the Component and start it async.
	//
	// NOTE: only a single version of the service may be registered
	//
	// An error occurs if registration fails for any of the following reasons:
	// 1. If a service with the same component Interface is already registered
	// 2. Interface is nil
	// 3. Version is nil
	// 4. the Client instance type is not assignable to the Service.
	Register(newComp ComponentConstructor) (Component, error)

	// MustRegister does the same as register, except that any error returned from Register triggers a panic
	MustRegister(newComp ComponentConstructor) Component

	// Component looks up the service via its Interface.
	// The component will be returned once it is up and running.
	Component(comp Interface) <-chan Component

	// Components returns all registered components
	Components() <-chan Component

	DependencyChecks

	// RegisterShutdownHook registers a function that will be invoked when the container shuts down.
	// The function will be invoked after all container managed components are stopped.
	// The functions are invoked in reverse order, i.e., the last one registered will be the first one that is run.
	// The name is used for tracking purposes, e.g., for logging.
	RegisterShutdownHook(name string, f func() error)
}
