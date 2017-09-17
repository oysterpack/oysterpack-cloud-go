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
	stdreflect "reflect"
	"time"

	"io"

	"sync"

	"github.com/Masterminds/semver"
	"github.com/oysterpack/oysterpack.go/oysterpack/commons"
	"github.com/oysterpack/oysterpack.go/oysterpack/commons/reflect"
	"github.com/oysterpack/oysterpack.go/oysterpack/logging"
	"github.com/rs/zerolog"
)

// Service represents the backend service.
// use NewService() to create a new instance
type Service interface {
	Interface() ServiceInterface

	Version() *semver.Version

	StartAsync() error

	Stop()
	StopAsyc()

	StopTriggered() bool
	StopTrigger() StopTrigger

	AwaitUntilRunning() error
	AwaitRunning(wait time.Duration) error

	AwaitUntilStopped() error
	AwaitStopped(wait time.Duration) error

	State() State
	FailureCause() error

	Logger() zerolog.Logger

	ServiceDependencies() InterfaceDependencies
}

// InterfaceDependencies represents a service's interface dependencies with version constraints
type InterfaceDependencies map[ServiceInterface]*semver.Constraints

// Service interface implementation
//
// Features
// ========
// 1. Services have a lifecycle defined by its Init, Run, and Destroy functions
// - the service lifecyle state is maintained via ServiceState
// - each state transition is logged as a STATE_CHANGED event async
//   - NOTE: thus, when the application is stopped, the logging goroutines may not get a chance to run before the process terminates.
// - when shut down is triggered the STOP_TRIGGERED event is logged
// 2. Services run in their own separate goroutine
// 3. Each service has their own logger.
//
// Services must be designed to be thread safe. The solution is to design the service as a message processing service
// leveraging channels and goroutines.
//
//
// TODO: Dependencies
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
type service struct {
	serviceInterface ServiceInterface

	version semver.Version

	lifeCycle

	logger zerolog.Logger

	dependencies InterfaceDependencies
}

// ServiceInterface represents the interface from the client's perspective, i.e., it defines the service's functionality.
type ServiceInterface reflect.InterfaceType

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
	Service
}

// Init is a function that is used to initialize the service during startup
type Init func(*Context) error

// Run is responsible to responding to a message on the StopTrigger channel.
// When a message is received from the StopTrigger, then the Run function should stop running ASAP.
type Run func(*Context) error

// Destroy is a function that is used to perform any cleanup during service shutdown.
type Destroy func(*Context) error

// ServiceSettings is used by NewService to create a new service instance
type ServiceSettings struct {
	// REQUIRED - represents the service interface from the client's perspective
	// If the service has no direct client API, e.g., a network based service, then use an empty interface{}
	ServiceInterface

	semver.Version

	// OPTIONAL - functions that define the service lifecycle
	Init
	Run
	Destroy

	// REQUIRED - InterfaceDependencies returns the Service interfaces that this service depends with version constraints
	// It can be used to check if all service Dependencies satisfied by the application.
	InterfaceDependencies

	LogSettings
}

// LogSettings groups the log settings for the service
type LogSettings struct {
	// OPTIONAL - used to specify an alternative writer for the service logger
	LogOutput io.Writer

	// OPTIONAL - if not specified then the global default log level is used
	LogLevel *zerolog.Level
}

// NewService creates and returns a new Service instance in the 'New' state.
//
// ServiceInterface:
// - must be an interface which defines the service's interface
// - if nil or not an interface, then the method panics
// All service life cycle functions are optional.
// Any panic that occurs in the supplied functions is converted to a PanicError.
func NewService(settings ServiceSettings) Service {
	serviceInterface := settings.ServiceInterface
	init := settings.Init
	run := settings.Run
	destroy := settings.Destroy

	checkServiceInterface := func() {
		if serviceInterface == nil {
			panic("ServiceInterface is required")
		}
		switch serviceInterface.Kind() {
		case stdreflect.Interface:
		default:
			if kind := serviceInterface.Elem().Kind(); kind != stdreflect.Interface {
				panic(fmt.Sprintf("ServiceInterface (%T) must be an interface, but was a %v", serviceInterface, kind))
			}
			serviceInterface = serviceInterface.Elem()
		}
	}

	instrumentInit := func() {
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
	}

	instrumentRun := func() {
		if run == nil {
			run = func(ctx *Context) error {
				<-ctx.StopTrigger()
				return nil
			}
		} else {
			_run := run
			run = func(ctx *Context) (err error) {
				defer func() {
					if p := recover(); p != nil {
						err = &PanicError{Panic: p, Message: "Service.run()"}
					}
				}()
				return _run(ctx)
			}
		}
	}

	instrumentDestroy := func() {
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
	}

	newService := func() Service {
		svcLog := logger.With().Dict(logging.SERVICE, logging.InterfaceTypeDict(serviceInterface)).Logger()
		if settings.LogOutput != nil {
			svcLog = svcLog.Output(settings.LogOutput)
		}
		if settings.LogLevel != nil {
			svcLog = svcLog.Level(*settings.LogLevel)
		}
		svcLog.Info().Str(logging.FUNC, "NewService").Msg("")

		svc := &service{
			serviceInterface: serviceInterface,
			version:          settings.Version,
			lifeCycle: lifeCycle{
				serviceState: NewServiceState(),
				init:         init,
				run:          run,
				destroy:      destroy,
			},
			logger: svcLog,
		}
		if len(settings.InterfaceDependencies) > 0 {
			svc.dependencies = settings.InterfaceDependencies
		}
		return svc
	}

	checkServiceInterface()
	instrumentInit()
	instrumentRun()
	instrumentDestroy()
	return newService()
}

// State returns the current State
func (svc *service) State() State {
	state, _ := svc.serviceState.State()
	return state
}

// FailureCause returns the error that caused the service to fail.
// The service State should be Failed.
func (svc *service) FailureCause() error {
	return svc.serviceState.FailureCause()
}

// awaitState blocks until the desired state is reached
// If the wait duration <= 0, then this method blocks until the desired state is reached.
// If the desired state has past, then a PastStateError is returned
func (svc *service) awaitState(desiredState State, timeout time.Duration) error {
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
		defer commons.IgnorePanic()
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
func (svc *service) AwaitRunning(wait time.Duration) error {
	return svc.awaitState(Running, wait)
}

// AwaitUntilRunning waits for the Service to reach the running state
func (svc *service) AwaitUntilRunning() error {
	i := 0
	for {
		if err := svc.AwaitRunning(10 * time.Second); err != nil {
			return err
		}
		if svc.State().Running() {
			return nil
		}
		i++
		svc.logger.Info().Str(logging.FUNC, "AwaitUntilRunning").Int("i", i).Msg("")
	}
}

// AwaitStopped waits for the Service to terminate, i.e., reach the Terminated or Failed state
// if the service terminates in a Failed state, then the service failure cause is returned
func (svc *service) AwaitStopped(wait time.Duration) error {
	if err := svc.awaitState(Terminated, wait); err != nil {
		return svc.serviceState.failureCause
	}
	return nil
}

// AwaitUntilStopped waits until the service is stopped
// If the service failed, then the failure cause is returned
func (svc *service) AwaitUntilStopped() error {
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
		svc.logger.Debug().Str(logging.FUNC, "AwaitUntilStopped").Int("i", i).Msg("")
	}
}

// StartAsync initiates service startup.
// If the service state is 'New', this initiates startup and returns immediately.
// Returns an IllegalStateError if the service state is not 'New'.
// A stopped service may not be restarted.
func (svc *service) StartAsync() error {
	const FUNC = "StartAsync"

	if !svc.serviceState.state.New() {
		err := &IllegalStateError{
			State:   svc.serviceState.state,
			Message: "A service can only be started in the 'New' state",
		}
		svc.logger.Info().Str(logging.FUNC, FUNC).Err(err).Msg("")
		return err
	}

	// log state changes async- wait for the goroutine to start before proceeding
	wait := sync.WaitGroup{}
	wait.Add(1)
	go func() {
		wait.Done()
		l := svc.lifeCycle.serviceState.NewStateChangeListener()
		for stateChange := range l.Channel() {
			svc.logger.Info().
				Dict(logging.EVENT, STATE_CHANGED.Dict()).
				Str(logging.STATE, stateChange.String()).
				Msg("")
		}
	}()
	wait.Wait()

	// start up the service
	go func() {
		svc.stopTrigger = make(chan struct{})
		ctx := &Context{svc}
		svc.serviceState.Starting()
		if err := svc.init(ctx); err != nil {
			svc.destroy(ctx)
			svc.serviceState.Failed(&ServiceError{State: Starting, Err: err})
			return
		}
		svc.serviceState.Running()
		if err := svc.run(ctx); err != nil {
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

	svc.logger.Info().Str(logging.FUNC, FUNC).Msg("")

	return nil
}

// StopAsyc initiates service shutdown.
// If the service is starting or running, this initiates service shutdown and returns immediately.
// If the service is new, it is terminated without having been started nor stopped.
// If the service has already been stopped, this method returns immediately without taking action.
func (svc *service) StopAsyc() {
	const FUNC = "StopAsyc"
	if svc.serviceState.state.Stopped() {
		svc.logger.Info().Str(logging.FUNC, FUNC).Msg("service is already stopped")
		return
	}
	svc.stopTriggered = true
	if svc.serviceState.state.New() {
		svc.serviceState.Terminated()
		svc.logger.Info().Str(logging.FUNC, FUNC).Msg("service was never started")
		return
	}
	func() {
		defer commons.IgnorePanic()
		close(svc.stopTrigger)
	}()

	svc.logger.Info().Str(logging.FUNC, FUNC).Dict(logging.EVENT, STOP_TRIGGERED.Dict()).Msg("")
}

// Stop invokes StopAsync() followed by AwaitUntilStopped()
func (svc *service) Stop() {
	if svc.State().Stopped() {
		return
	}
	svc.StopAsyc()
	svc.AwaitUntilStopped()
}

// StopTriggered returns true if the service was triggered to stop.
func (svc *service) StopTriggered() bool {
	return svc.stopTriggered
}

// Interface returns the service interface which defines the service functionality
func (svc *service) Interface() ServiceInterface {
	return svc.serviceInterface
}

// Version returns the service version
func (a *service) Version() *semver.Version {
	return &a.version
}

// StopTrigger returns the channel to listen for the stopping
func (svc *service) StopTrigger() StopTrigger {
	return svc.stopTrigger
}

// Logger returns the service's logger
func (svc *service) Logger() zerolog.Logger {
	return svc.logger
}

// ServiceDependencies returns the service's dependencies
func (svc *service) ServiceDependencies() InterfaceDependencies {
	return svc.dependencies
}

func (svc *service) String() string {
	return fmt.Sprintf("Service : %v.%v", svc.serviceInterface.PkgPath(), svc.serviceInterface.Name())
}

// StopTrigger is used to notify the service to stop.
// Closing the channel is the stop signal
type StopTrigger <-chan struct{}

// ServiceConstructor returns a new instance of a Service in the New state
type ServiceConstructor func() Service

// ServiceReference represents something that has a reference to a service.
type ServiceReference interface {
	Service() Service
}
