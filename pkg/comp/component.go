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
	stdreflect "reflect"

	"github.com/Masterminds/semver"
	"github.com/oysterpack/oysterpack.go/pkg/commons/reflect"
	"github.com/oysterpack/oysterpack.go/pkg/logging"
	"github.com/oysterpack/oysterpack.go/pkg/metrics"
	"zombiezen.com/go/capnproto2"
)

type ComponentConstructor func() Component

// Component represents a component that provides the functionality defined by the Interface
type Component interface {
	// Desc describes the component
	Desc() *Descriptor

	// Interface is the interface the component implements, i.e., the component instance can be safely casted to this interface
	Interface() Interface

	// State the current component lifecycle state
	State() State
	// FailureCause if the service is in FAILED state, then this is the error that caused it
	FailureCause() error

	// Start the service, if not already started
	// As the component transitions states, the new state is sent on the channel.
	// The channel will be closed once the component is running or reaches a FAILED state
	// If the component requires config to start, then it is passed in as a capnp message.
	// The registry is provided for the component to lookup dependencies.
	Start(config *capnp.Message, registry ComponentRegistry) <-chan State

	// Stop the service, if not already stopped.
	// As the component transitions states, the new state is sent on the channel.
	// The channel will be closed once the component reaches a terminal state.
	Stop() <-chan State

	// MetricOpts returns the metrics that are collected for this component
	MetricOpts() *metrics.MetricOpts

	// HealthChecks returns the component's health checks
	HealthChecks() HealthChecks

	// LogEvents returns what events are logged by this component
	LogEvents() []logging.Event

	// ConfigTag is optional, but if specified, it is used to lookup a config with the specified tag.
	// The config id is constructed using the component Desc and the ConfigTag
	ConfigTag() ConfigTag

	// ConfigType is the expected type for the config capnp message
	ConfigType() stdreflect.Type

	// Dependencies lists what other components this component depends on
	Dependencies() Dependencies
}

// Dependencies represents a service's interface dependencies with version constraints
type Dependencies map[Interface]*semver.Constraints

type ConfigTag string

// Interface represents the interface from the client's perspective, i.e., it defines the compoenent's functionality.
type Interface reflect.InterfaceType

func checkInterface(compInterface Interface) Interface {
	if compInterface == nil {
		logger.Panic().Msg("Interface is required")
	}
	switch compInterface.Kind() {
	case stdreflect.Interface:
	default:
		if kind := compInterface.Elem().Kind(); kind != stdreflect.Interface {
			logger.Panic().Msgf("Interface (%T) must be an interface, but was a %v", compInterface, kind)
		}
		compInterface = compInterface.Elem()
	}
	return compInterface
}
