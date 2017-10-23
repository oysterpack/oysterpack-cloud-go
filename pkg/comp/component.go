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

	"time"

	"fmt"

	"errors"

	"github.com/Masterminds/semver"
	"github.com/oysterpack/oysterpack.go/pkg/commons/reflect"
	"github.com/oysterpack/oysterpack.go/pkg/logging"
	"github.com/oysterpack/oysterpack.go/pkg/metrics"
	"github.com/rs/zerolog"
	"zombiezen.com/go/capnproto2"
)

type ComponentConstructor func() Component

// Component represents a component that provides the functionality defined by the Interface
type Component interface {
	// Desc provides component's name and version, as well as the namespace and system it belongs to.
	Desc() *Descriptor

	// Interface is the interface the component implements, i.e., the component instance can be safely casted to this interface
	Interface() Interface

	// State the current component lifecycle state
	State() State

	// StateChan returns a channel that can be used to monitor component lifecycle state transitions.
	// The channel is closed once the state reaches a terminal state.
	StateChan() <-chan StateChanged

	// FailureCause if the component is in FAILED state, then this is the error that caused it
	FailureCause() error

	// Start the component, if not already started
	// As the component transitions states, the new state is sent on the channel.
	// The channel will be closed once the component is running or reaches a FAILED state
	// If the component requires config to start, then it is passed in as a capnp message.
	// The ctx is provided for the component to lookup dependencies.
	Start(config *capnp.Message, ctx Context) <-chan State

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

	Logger() *zerolog.Logger
}

// StateChanged represents a state change event for a Component.
// NOTE: based on timing, the current state of the Component could have changed since this event was fired.
type StateChanged struct {
	Component Component
	State     State
	Time      time.Time
}

// Dependencies represents a component's interface dependencies with version constraints
type Dependencies map[Interface]*semver.Constraints

type ConfigTag string

// Interface represents the interface from the client's perspective, i.e., it defines the compoenent's functionality.
type Interface reflect.InterfaceType

func checkInterface(compInterface Interface) (Interface, error) {
	if compInterface == nil {
		return nil, errors.New("Interface is required")
	}
	switch compInterface.Kind() {
	case stdreflect.Interface:
	default:
		if kind := compInterface.Elem().Kind(); kind != stdreflect.Interface {
			return nil, fmt.Errorf("Interface (%T) must be an interface, but was a %v", compInterface, kind)
		}
		compInterface = compInterface.Elem()
	}
	return compInterface, nil
}

type component struct {
	desc          *Descriptor
	compInterface Interface

	state        State
	failureCause error

	configTag  ConfigTag
	configType stdreflect.Type

	metrics      *metrics.MetricOpts
	healthchecks HealthChecks
	logEvents    []logging.Event
	dependencies Dependencies

	logger *zerolog.Logger

	init      func(comp *component, config *capnp.Message, ctx Context)
	handleMsg func(comp *component, msg interface{}) error
	destroy   func(comp *component)

	serverChan chan interface{}
}

func (a *component) server(stopping <-chan struct{}) error {
	for {
		select {
		case <-stopping:
			a.destroy(a)
			return a.failureCause
		case msg := <-a.serverChan:
			if err := a.handleComponentMessage(msg); err != nil {
				switch err.(type) {
				case *UnsupportedMessageError:
					if err := a.handleMsg(a, msg); err != nil {
						return err
					}
				}
			}
		}
	}
}

func (a *component) handleComponentMessage(msg interface{}) error {
	// TODO
	return &UnsupportedMessageError{nil}
}

func (a *component) Desc() *Descriptor {
	return a.desc
}

func (a *component) Interface() Interface {
	return a.compInterface
}

func (a *component) State() State {
	return a.state
}

func (a *component) StateChan() <-chan StateChanged {
	c := make(chan StateChanged, 4)
	//TODO
	return c
}

func (a *component) FailureCause() error {
	return a.failureCause
}

func (a *component) Start(config *capnp.Message, ctx Context) <-chan State {
	c := make(chan State, 4)
	//TODO
	return c
}

// MetricOpts returns the metrics that are collected for this component
func (a *component) MetricOpts() *metrics.MetricOpts {
	return a.metrics
}

func (a *component) HealthChecks() HealthChecks {
	return a.healthchecks
}

func (a *component) LogEvents() []logging.Event {
	return a.logEvents
}

func (a *component) ConfigTag() ConfigTag {
	return a.configTag
}

func (a *component) ConfigType() stdreflect.Type {
	return a.configType
}

func (a *component) Dependencies() Dependencies {
	return a.dependencies
}

func (a *component) Logger() *zerolog.Logger {
	return a.logger
}
