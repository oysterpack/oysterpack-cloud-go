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

type ComponentRegistry interface {
	// MustRegister will create a new instance of the component using the supplied constructor.
	// The Component must implement the component Interface - otherwise the method panics.
	// It will then register the Component and start it async.
	//
	// NOTE: only a single version of the service may be registered
	//
	// A panic occurs if registration fails for any of the following reasons:
	// 1. If a service with the same component Interface is already registered
	// 2. Interface is nil
	// 3. Version is nil
	// 4. the Client instance type is not assignable to the Service.
	MustRegister(comp ComponentConstructor) Component

	// Register does the same as MustRegister, but instead returns an error instead of panicking
	Register(comp ComponentConstructor) (Component, error)

	// Unregister will unregister the service and return it. If no such service exists then nil is returned.
	Unregister(comp Interface) Component

	// Component looks up the service via its Interface.
	// The component will be returned once it is up and running.
	Component(comp Interface) <-chan Component

	// Components returns all registered components
	Components() <-chan Component

	DependencyChecks
}
