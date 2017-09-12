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

// Service represents an object with an operational state.
//
// Features
// ========
// 1. Services have a lifecycle defined by its Init, Run, and Destroy functions
// - the service lifecyle state is maintained via ServiceState
// - each state transition is logged as a STATE_CHANGED event
// - services must be created using NewService()
// - when shut down is triggered the STOP_TRIGGERED event is logged
// 2. Services run in their own separate goroutine
// 3. Each service has their own logger.
//
// Services must be designed to be concurrency safe. The solution is to design the service as a message processing service
// leveraging channels and goroutines. There are several approaches :
// 1. one channel per service function - the service would use a select within an infinite for loop to process messages sent via channels
// 2. one channel per service - the service would use a type switch to map message handlers to messages
// 3. a hybrid of the above approaches
//
// NOTE: If a reply is required, then the message will have a reply channel.
//
// The recommended approach is for the service implementation to implement the ClientServer interface :
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
// TODO: health checks
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

	svcLog := logger.With().Dict(logging.SERVICE, logging.ServiceDict(serviceInterface)).Logger()
	svcLog.Info().Str(logging.FUNC, "NewService").Msg("")

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
func (svc *Service) awaitState(desiredState State, timeout time.Duration) error {
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
	if timeout > 0 {
		timer := time.AfterFunc(timeout, l.Cancel)
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

// AwaitUntilRunning waits for the Service to reach the running state
func (svc *Service) AwaitUntilRunning() error {
	i := 0
	for {
		if err := svc.AwaitRunning(10 * time.Second); err != nil {
			return err
		}
		if svc.State().Running() {
			return nil
		}
		i++
		svc.Logger.Info().Str(logging.FUNC, "AwaitUntilRunning").Int("i", i).Msg("")
	}
}

// AwaitStopped waits for the Service to terminate, i.e., reach the Terminated or Failed state
// if the service terminates in a Failed state, then the service failure cause is returned
func (svc *Service) AwaitStopped(wait time.Duration) error {
	if err := svc.awaitState(Terminated, wait); err != nil {
		return svc.serviceState.failureCause
	}
	return nil
}

// AwaitUntilStopped waits until the service is stopped
// If the service failed, then the failure cause is returned
func (svc *Service) AwaitUntilStopped() error {
	if svc.State().Stopped() {
		return svc.FailureCause()
	}
	i := 0
	for {
		svc.AwaitStopped(10 * time.Second)
		if svc.State().Stopped() {
			return svc.FailureCause()
		}
		i++
		svc.Logger.Info().Str(logging.FUNC, "AwaitUntilStopped").Int("i", i).Msg("")
	}
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
	svc.Logger.Info().Str(logging.FUNC, FUNC).Dict(logging.EVENT, STOP_TRIGGERED.Dict()).Msg("")
}

// Stop invokes StopAsync() followed by AwaitUntilStopped()
func (svc *Service) Stop() {
	if svc.State().Stopped() {
		return
	}
	svc.StopAsyc()
	svc.AwaitUntilStopped()
}

// StopTriggered returns true if the service was triggered to stop.
func (svc *Service) StopTriggered() bool {
	return svc.stopTriggered
}

// Interface returns the service interface which defines the service functionality
func (svc *Service) Interface() reflect.Type {
	return svc.serviceInterface
}

func (svc *Service) String() string {
	return fmt.Sprintf("Service : %v.%v", svc.serviceInterface.PkgPath(), svc.serviceInterface.Name())
}

// StopTrigger is used to notify the service to stop.
// Closing the channel is the stop signal
type StopTrigger <-chan struct{}

// ServiceConstructor returns a new instance of a Service in the New state
type ServiceConstructor func() *Service
