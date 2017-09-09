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
	"github.com/oysterpack/oysterpack.go/oysterpack/internal/utils"
	"github.com/oysterpack/oysterpack.go/oysterpack/logging"
	"github.com/rs/zerolog"
	"reflect"
	"time"
)

var STOP_TRIGGERED = &logging.Event{0, "STOP_TRIGGERED"}
var STATE_CHANGED = &logging.Event{1, "STATE_CHANGED"}

// Service represents an object with an operational state, with methods to start and stop.
// The service runs in its own goroutine.
// The service has a life cycle linked to its ServiceState.
// It must be created using NewService
//
// Services must be designed to be concurrent, but also concurrent safe.
// The solution is to design the service as a message processing service leveraging channels and goroutines.
// There would be a channel for each type of message a service can handle. Goroutines are leveraged for concurrent message processing.
// The design pattern is for the backend service Run function to process messages sent via channels.
// The Run function should be designed to multiplex on channels - the StopTrigger channel and a channel for each type of
// message that the service can handle.
//
// Functions relay messages to the service backend via messages. If a reply is required, then the message will have a reply channel.
//
// Design Options"
// 1. The package represents the service. Each service will live in its own package. The service will be defined by the
// functions that are exposed by the package.
// 2. A struct that implements ServiceComposite encapsulates the service's state and behaivor
//
// 		type ConfigService struct {
// 			svc service.Service
//
//			// service specific state
// 		}
//
//
// TODO: dependencies
// TODO: events
// TODO: logging
// TODO: metrics
// TODO: healthchecks
// TODO: alarms
// TODO: error / panic logging
// TODO: error / panic handling
// TODO: supervision / restart policy
// TODO: config (JSON)
// TODO: readiness probe
// TODO: liveliness probe
// TODO: devops
// TODO: security
// TODO: gRPC - frontend
type Service struct {
	serviceInterface commons.InterfaceType

	lifeCycle

	zerolog.Logger
}

// ServiceComposite
type ServiceComposite interface {
	Service() Service
}

// LifeCycle encapsulates the service lifecycle, including the service's backend functions, i.e., Init, Run, Destroy
type lifeCycle struct {
	serviceState *ServiceState

	stopTriggered bool
	// closing the channel signals to the Run function that stop has been triggered
	stopTrigger chan struct{}

	init    Init
	run     Run
	destroy Destroy
}

// Context represents the service context that is exposed to the Init and Destroy funcs
type Context struct {
	service *Service
}

// Init is a function that is used to initialize the service during startup
type Init func(*Context) error

// Run is responsible to responding to a message on the StopTrigger channel.
// When a message is received from the StopTrigger, then the Run function should stop running ASAP.
type Run func(*RunContext) error

// RunContext represents the service context that is exposed to the Run func
type RunContext struct {
	service *Service
}

// StopTriggered returns true if the service was triggered to stop
func (ctx *RunContext) StopTriggered() bool {
	return ctx.service.stopTriggered
}

// StopTrigger returns the channel that the Run func should use to listen on for the stop trigger.
// Closing the channel will signal that service has been triggered to stop.
func (ctx *RunContext) StopTrigger() StopTrigger {
	return ctx.service.stopTrigger
}

// Destroy is a function that is used to perform any cleanup during service shutdown.
type Destroy func(*Context) error

// NewService creates and returns a new Service instance in the 'New' state.
//
// serviceInterface:
// - must be an interface which defines the service's interface
// - if nil or not an interface, then the method panics
// All service life cycle functions are optional.
// Any panic that occur in the supplied functions is converted to a PanicError.
func NewService(serviceInterface commons.InterfaceType, init Init, run Run, destroy Destroy) *Service {
	if serviceInterface == nil {
		panic("serviceInterface is required")
	}
	switch serviceInterface.Kind() {
	case reflect.Interface:
	default:
		if kind := serviceInterface.Elem().Kind(); kind != reflect.Interface {
			panic(fmt.Sprintf("serviceInterface (%T) must be an interface, but was a %v", serviceInterface, kind))
		}
		serviceInterface = serviceInterface.Elem()
	}

	if init == nil {
		init = func(ctx *Context) error { return nil }
	} else {
		_init := init
		init = func(ctx *Context) (err error) {
			defer func() {
				if p := recover(); p != nil {
					err = &PanicError{Panic: p, Message: "Service.init()"}
				}
			}()
			return _init(ctx)
		}
	}

	if run == nil {
		run = func(ctx *RunContext) error {
			<-ctx.StopTrigger()
			return nil
		}
	} else {
		_run := run
		run = func(ctx *RunContext) (err error) {
			defer func() {
				if p := recover(); p != nil {
					err = &PanicError{Panic: p, Message: "Service.run()"}
				}
			}()
			return _run(ctx)
		}
	}

	if destroy == nil {
		destroy = func(ctx *Context) error { return nil }
	} else {
		_destroy := destroy
		destroy = func(ctx *Context) (err error) {
			defer func() {
				if p := recover(); p != nil {
					err = &PanicError{Panic: p, Message: "Service.destroy()"}
				}
			}()
			return _destroy(ctx)
		}
	}

	svcLog := logger.With().
		Dict(logging.SERVICE, zerolog.Dict().
			Str(logging.PACKAGE, serviceInterface.PkgPath()).
			Str(logging.TYPE, serviceInterface.Name())).
		Logger()

	svcLog.Info().
		Str(logging.FUNC, "NewService").
		Msg("")
	return &Service{
		serviceInterface: serviceInterface,
		lifeCycle: lifeCycle{
			serviceState: NewServiceState(),
			init:         init,
			run:          run,
			destroy:      destroy,
		},
		Logger: svcLog,
	}
}

// State returns the current State
func (svc *Service) State() State {
	state, _ := svc.serviceState.State()
	return state
}

// FailureCause returns the error that caused the service to fail.
// The service State should be Failed.
func (svc *Service) FailureCause() error {
	return svc.serviceState.FailureCause()
}

// awaitState blocks until the desired state is reached
// If the wait duration <= 0, then this method blocks until the desired state is reached.
// If the desired state has past, then a PastStateError is returned
func (svc *Service) awaitState(desiredState State, wait time.Duration) error {
	matches := func(currentState State) (bool, error) {
		switch {
		case currentState == desiredState:
			return true, nil
		case currentState > desiredState:
			if svc.State().Failed() {
				return false, svc.FailureCause()
			}
			return false, &PastStateError{Past: desiredState, Current: currentState}
		default:
			return false, nil
		}
	}

	if reachedState, err := matches(svc.State()); err != nil {
		return err
	} else if reachedState {
		return nil
	}
	l := svc.serviceState.NewStateChangeListener()
	if wait > 0 {
		timer := time.AfterFunc(wait, l.Cancel)
		defer func() {
			timer.Stop()
			l.Cancel()
		}()
	} else {
		defer l.Cancel()
	}
	// in case the service started matches in the meantime, seed the messages with the current state
	go func() {
		// ignore panics caused by sending on a closed messages
		// the messages might be closed if the service failed
		defer utils.IgnorePanic()
		if stateChangeChann := svc.serviceState.stateChangeChannel(l); stateChangeChann != nil {
			stateChangeChann <- svc.State()
		}
	}()
	for state := range l.Channel() {
		if reachedState, err := matches(state); err != nil {
			return err
		} else if reachedState {
			return nil
		}
	}

	return svc.FailureCause()
}

// AwaitRunning waits for the Service to reach the running state
func (svc *Service) AwaitRunning(wait time.Duration) error {
	return svc.awaitState(Running, wait)
}

// AwaitTerminated waits for the Service to terminate, i.e., reach the Terminated or Failed state
// if the service terminates in a Failed state, then the service failure cause is returned
func (svc *Service) AwaitTerminated(wait time.Duration) error {
	if err := svc.awaitState(Terminated, wait); err != nil {
		return svc.serviceState.failureCause
	}
	return nil
}

// StartAsync initiates service startup.
// If the service state is 'New', this initiates startup and returns immediately.
// Returns an IllegalStateError if the service state is not 'New'.
// A stopped service may not be restarted.
func (svc *Service) StartAsync() error {
	const FUNC = "StartAsync"
	if svc.serviceState.state.New() {
		go func() {
			svc.stopTrigger = make(chan struct{})
			ctx := &Context{
				service: svc,
			}
			svc.serviceState.Starting()
			if err := svc.init(ctx); err != nil {
				svc.destroy(ctx)
				svc.serviceState.Failed(&ServiceError{State: Starting, Err: err})
				return
			}
			runCtx := &RunContext{
				service: svc,
			}
			svc.serviceState.Running()
			if err := svc.run(runCtx); err != nil {
				svc.destroy(ctx)
				svc.serviceState.Failed(&ServiceError{State: Running, Err: err})
				return
			}
			svc.serviceState.Stopping()
			if err := svc.destroy(ctx); err != nil {
				svc.serviceState.Failed(&ServiceError{State: Stopping, Err: err})
				return
			}
			svc.serviceState.Terminated()
		}()
		// log state changes
		go func() {
			l := svc.lifeCycle.serviceState.NewStateChangeListener()
			for stateChange := range l.Channel() {
				svc.Logger.Info().
					Dict(logging.EVENT, STATE_CHANGED.Dict()).
					Str(logging.STATE, stateChange.String()).
					Msg("")
			}
		}()
		svc.Logger.Info().Str(logging.FUNC, FUNC).Msg("")

		return nil
	}
	err := &IllegalStateError{
		State:   svc.serviceState.state,
		Message: "A service can only be started in the 'New' state",
	}
	svc.Logger.Info().Str(logging.FUNC, FUNC).Err(err).Msg("")
	return err
}

// StopAsyc initiates service shutdown.
// If the service is starting or running, this initiates service shutdown and returns immediately.
// If the service is new, it is terminated without having been started nor stopped.
// If the service has already been stopped, this method returns immediately without taking action.
func (svc *Service) StopAsyc() {
	const FUNC = "StopAsyc"
	if svc.serviceState.state.Stopped() {
		svc.Logger.Info().Str(logging.FUNC, FUNC).Msg("service is already stopped")
		return
	}
	svc.stopTriggered = true
	if svc.serviceState.state.New() {
		svc.serviceState.Terminated()
		svc.Logger.Info().Str(logging.FUNC, FUNC).Msg("service was never started")
		return
	}
	close(svc.stopTrigger)
	svc.Logger.Info().
		Str(logging.FUNC, FUNC).
		Dict(logging.EVENT, STOP_TRIGGERED.Dict()).
		Msg("")
}

// StopTriggered returns true if the service was triggered to stop.
func (svc *Service) StopTriggered() bool {
	return svc.stopTriggered
}

// Interface returns the service interface which defines the service functionality
func (svc *Service) Interface() reflect.Type {
	return svc.serviceInterface
}

// StopTrigger is used to notify the service to stop.
// Closing the channel is the stop signal
type StopTrigger <-chan struct{}
