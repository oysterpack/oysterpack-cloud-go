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

	"sync"

	"github.com/Masterminds/semver"
	"github.com/oysterpack/oysterpack.go/pkg/commons"
	"github.com/oysterpack/oysterpack.go/pkg/commons/reflect"
	"github.com/oysterpack/oysterpack.go/pkg/logging"
	"github.com/oysterpack/oysterpack.go/pkg/metrics"
	"github.com/rs/zerolog"
)

// Service represents the backend service.
// use NewService() to create a new instance
type Service interface {
	Desc() *Descriptor

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

	Dependencies() InterfaceDependencies

	HealthChecks

	// MetricOpts define the service related metrics
	MetricOpts() *metrics.MetricOpts
}

// InterfaceDependencies represents a service's interface dependencies with version constraints
type InterfaceDependencies map[Interface]*semver.Constraints

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
	*Descriptor

	lifeCycle

	logger zerolog.Logger

	dependencies InterfaceDependencies

	healthchecks []metrics.HealthCheck

	metricOpts *metrics.MetricOpts
}

// Interface represents the interface from the client's perspective, i.e., it defines the service's functionality.
type Interface reflect.InterfaceType

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

func (a *service) Desc() *Descriptor {
	return a.Descriptor
}

// NewService creates and returns a new Service instance in the 'New' state.
//
// Interface:
// - must be an interface which defines the service's interface
// - if nil or not an interface, then the method panics
// All service life cycle functions are optional.
// Any panic that occurs in the supplied functions is converted to a PanicError.
func NewService(settings Settings) Service {
	serviceInterface := settings.Interface()
	init := settings.Init
	run := settings.Run
	destroy := settings.Destroy

	trapPanics := func(f func(*Context) error, msg string) func(*Context) error {
		if f == nil {
			return func(ctx *Context) error { return nil }
		}
		return func(ctx *Context) (err error) {
			defer func() {
				if p := recover(); p != nil {
					err = &PanicError{Panic: p, Message: msg}
				}
			}()
			return f(ctx)
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

	newService := func() Service {
		svcLog := logger.With().Dict(logging.SERVICE, logging.InterfaceTypeDict(serviceInterface)).Logger()
		if settings.LogOutput != nil {
			svcLog = svcLog.Output(settings.LogOutput)
		}
		if settings.LogLevel != nil {
			svcLog = svcLog.Level(*settings.LogLevel)
		}
		logSettings(svcLog.Info().Str(logging.FUNC, "NewService"), &settings).Msg("")

		svc := &service{
			Descriptor: settings.Descriptor,
			lifeCycle: lifeCycle{
				serviceState: NewServiceState(),
				init:         init,
				run:          run,
				destroy:      destroy,
			},
			logger:       svcLog,
			healthchecks: settings.HealthChecks,
		}
		if len(settings.InterfaceDependencies) > 0 {
			svc.dependencies = settings.InterfaceDependencies
		}
		if settings.Metrics == nil {
			svc.metricOpts = &metrics.MetricOpts{}
		} else {
			svc.metricOpts = settings.Metrics
		}

		return svc
	}

	checkSettings(&settings)
	init = trapPanics(init, "Service.init()")
	instrumentRun()
	destroy = trapPanics(destroy, "Service.destroy()")
	return newService()
}

// panics if settings are invalid
func checkSettings(settings *Settings) {
	settings.Descriptor.serviceInterface = checkInterface(settings.Interface())
	if settings.Version == nil {
		logger.Panic().Str(logging.SERVICE, settings.Interface().String()).Msgf("Version is required")
	}
}

func checkInterface(serviceInterface Interface) Interface {
	if serviceInterface == nil {
		logger.Panic().Msg("Interface is required")
	}
	switch serviceInterface.Kind() {
	case stdreflect.Interface:
	default:
		if kind := serviceInterface.Elem().Kind(); kind != stdreflect.Interface {
			logger.Panic().Msgf("Interface (%T) must be an interface, but was a %v", serviceInterface, kind)
		}
		serviceInterface = serviceInterface.Elem()
	}
	return serviceInterface
}

func logInterfaceDependencies(log *zerolog.Event, settings *Settings) {
	if settings.InterfaceDependencies != nil {
		deps := make([]string, len(settings.InterfaceDependencies))
		i := 0
		for k, v := range settings.InterfaceDependencies {
			deps[i] = fmt.Sprintf("%v : %v", k, v)
			i++
		}
		log.Strs("InterfaceDependencies", deps)
	}
}

func logHealthChecks(log *zerolog.Event, settings *Settings) {
	if len(settings.HealthChecks) > 0 {
		names := make([]string, len(settings.HealthChecks))
		for i, v := range settings.HealthChecks {
			names[i] = v.Key().String()
		}

		log.Strs("HealthChecks", names)
	}
}

func logCounterOpts(dict *zerolog.Event, settings *Settings) {
	if len(settings.Metrics.CounterOpts) > 0 {
		names := make([]string, len(settings.Metrics.CounterOpts))
		for i, v := range settings.Metrics.CounterOpts {
			names[i] = metrics.CounterFQName(v)
		}
		dict.Strs("Counters", names)
	}
}

func logCounterVecOpts(dict *zerolog.Event, settings *Settings) {
	if len(settings.Metrics.CounterVecOpts) > 0 {
		names := make([]string, len(settings.Metrics.CounterVecOpts))
		for i, v := range settings.Metrics.CounterVecOpts {
			names[i] = metrics.CounterFQName(v.CounterOpts)
		}
		dict.Strs("CounterVecss", names)
	}
}

func logGaugeOpts(dict *zerolog.Event, settings *Settings) {
	if len(settings.Metrics.GaugeOpts) > 0 {
		names := make([]string, len(settings.Metrics.GaugeOpts))
		for i, v := range settings.Metrics.GaugeOpts {
			names[i] = metrics.GaugeFQName(v)
		}
		dict.Strs("Gauges", names)
	}
}

func logGaugeVecOpts(dict *zerolog.Event, settings *Settings) {
	if len(settings.Metrics.GaugeVecOpts) > 0 {
		names := make([]string, len(settings.Metrics.GaugeVecOpts))
		for i, v := range settings.Metrics.GaugeVecOpts {
			names[i] = metrics.GaugeFQName(v.GaugeOpts)
		}
		dict.Strs("GaugeVecs", names)
	}
}

func logHistogramOpts(dict *zerolog.Event, settings *Settings) {
	if len(settings.Metrics.HistogramOpts) > 0 {
		names := make([]string, len(settings.Metrics.HistogramOpts))
		for i, v := range settings.Metrics.HistogramOpts {
			names[i] = metrics.HistogramFQName(v)
		}
		dict.Strs("Histograms", names)
	}
}

func logHistogramVecOpts(dict *zerolog.Event, settings *Settings) {
	if len(settings.Metrics.HistogramVecOpts) > 0 {
		names := make([]string, len(settings.Metrics.HistogramVecOpts))
		for i, v := range settings.Metrics.HistogramVecOpts {
			names[i] = metrics.HistogramFQName(v.HistogramOpts)
		}
		dict.Strs("HistogramVecs", names)
	}
}

func logSummaryOpts(dict *zerolog.Event, settings *Settings) {
	if len(settings.Metrics.SummaryOpts) > 0 {
		names := make([]string, len(settings.Metrics.SummaryOpts))
		for i, v := range settings.Metrics.SummaryOpts {
			names[i] = metrics.SummaryFQName(v)
		}
		dict.Strs("Summaries", names)
	}
}

func logSummaryVecOpts(dict *zerolog.Event, settings *Settings) {
	if len(settings.Metrics.SummaryVecOpts) > 0 {
		names := make([]string, len(settings.Metrics.SummaryVecOpts))
		for i, v := range settings.Metrics.SummaryVecOpts {
			names[i] = metrics.SummaryFQName(v.SummaryOpts)
		}
		dict.Strs("SummaryVecs", names)
	}
}

func logMetrics(log *zerolog.Event, settings *Settings) {
	if settings.Metrics != nil {
		dict := zerolog.Dict()

		logCounterOpts(dict, settings)
		logCounterVecOpts(dict, settings)

		logGaugeOpts(dict, settings)
		logGaugeVecOpts(dict, settings)

		logHistogramOpts(dict, settings)
		logHistogramVecOpts(dict, settings)

		logSummaryOpts(dict, settings)
		logSummaryVecOpts(dict, settings)

		log.Dict("Metrics", dict)
	}
}

func logSettings(log *zerolog.Event, settings *Settings) *zerolog.Event {
	log.Str(logging.NAMESPACE, settings.Namespace()).
		Str(logging.SYSTEM, settings.System()).
		Str(logging.COMPONENT, settings.Component()).
		Str(logging.VERSION, settings.Version().String())

	logInterfaceDependencies(log, settings)
	logHealthChecks(log, settings)
	logMetrics(log, settings)

	return log
}

// MetricOpts returns the service related metrics
func (a *service) MetricOpts() *metrics.MetricOpts {
	return a.metricOpts
}

// State returns the current State
func (a *service) State() State {
	state, _ := a.serviceState.State()
	return state
}

// FailureCause returns the error that caused the service to fail.
// The service State should be Failed.
func (a *service) FailureCause() error {
	return a.serviceState.FailureCause()
}

// awaitState blocks until the desired state is reached
// If the wait duration <= 0, then this method blocks until the desired state is reached.
// If the desired state has past, then a PastStateError is returned
func (a *service) awaitState(desiredState State, timeout time.Duration) error {
	matches := func(currentState State) (bool, error) {
		switch {
		case currentState == desiredState:
			return true, nil
		case currentState > desiredState:
			if a.State().Failed() {
				return false, a.FailureCause()
			}
			return false, &PastStateError{Past: desiredState, Current: currentState}
		default:
			return false, nil
		}
	}

	if reachedState, err := matches(a.State()); err != nil {
		return err
	} else if reachedState {
		return nil
	}

	l := a.serviceState.NewStateChangeListener()
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
		if stateChangeChann := a.serviceState.stateChangeChannel(l); stateChangeChann != nil {
			stateChangeChann <- a.State()
		}
	}()
	for state := range l.Channel() {
		if reachedState, err := matches(state); err != nil {
			return err
		} else if reachedState {
			return nil
		}
	}

	return a.FailureCause()
}

// AwaitRunning waits for the Service to reach the running state
func (a *service) AwaitRunning(wait time.Duration) error {
	return a.awaitState(Running, wait)
}

// AwaitUntilRunning waits for the Service to reach the running state
func (a *service) AwaitUntilRunning() error {
	i := 0
	for {
		if err := a.AwaitRunning(10 * time.Second); err != nil {
			return err
		}
		if a.State().Running() {
			return nil
		}
		i++
		a.logger.Info().Str(logging.FUNC, "AwaitUntilRunning").Int("i", i).Msg("")
	}
}

// AwaitStopped waits for the Service to terminate, i.e., reach the Terminated or Failed state
// if the service terminates in a Failed state, then the service failure cause is returned
func (a *service) AwaitStopped(wait time.Duration) error {
	if err := a.awaitState(Terminated, wait); err != nil {
		return a.serviceState.failureCause
	}
	return nil
}

// AwaitUntilStopped waits until the service is stopped
// If the service failed, then the failure cause is returned
func (a *service) AwaitUntilStopped() error {
	if a.State().Stopped() {
		return a.FailureCause()
	}
	i := 0
	for {
		a.AwaitStopped(10 * time.Second)
		if a.State().Stopped() {
			return a.FailureCause()
		}
		i++
		a.logger.Debug().Str(logging.FUNC, "AwaitUntilStopped").Int("i", i).Msg("")
	}
}

// StartAsync initiates service startup.
// If the service state is 'New', this initiates startup and returns immediately.
// Returns an IllegalStateError if the service state is not 'New'.
// A stopped service may not be restarted.
func (a *service) StartAsync() error {
	const FUNC = "StartAsync"

	if !a.serviceState.state.New() {
		err := &IllegalStateError{
			State:   a.serviceState.state,
			Message: "A service can only be started in the 'New' state",
		}
		a.logger.Info().Str(logging.FUNC, FUNC).Err(err).Msg("")
		return err
	}

	// log state changes async- wait for the goroutine to start before proceeding
	wait := sync.WaitGroup{}
	wait.Add(1)
	go func() {
		wait.Done()
		l := a.lifeCycle.serviceState.NewStateChangeListener()
		for stateChange := range l.Channel() {
			a.logger.Info().
				Dict(logging.EVENT, STATE_CHANGED.Dict()).
				Str(logging.STATE, stateChange.String()).
				Msg("")
		}
	}()
	wait.Wait()

	// start up the service
	go func() {
		a.stopTrigger = make(chan struct{})
		ctx := &Context{a}
		a.serviceState.Starting()
		if err := a.init(ctx); err != nil {
			a.destroy(ctx)
			a.serviceState.Failed(&ServiceError{State: Starting, Err: err})
			return
		}
		a.serviceState.Running()
		if err := a.run(ctx); err != nil {
			a.destroy(ctx)
			a.serviceState.Failed(&ServiceError{State: Running, Err: err})
			return
		}
		a.serviceState.Stopping()
		if err := a.destroy(ctx); err != nil {
			a.serviceState.Failed(&ServiceError{State: Stopping, Err: err})
			return
		}
		a.serviceState.Terminated()
	}()

	a.logger.Info().Str(logging.FUNC, FUNC).Msg("")

	return nil
}

// StopAsyc initiates service shutdown.
// If the service is starting or running, this initiates service shutdown and returns immediately.
// If the service is new, it is terminated without having been started nor stopped.
// If the service has already been stopped, this method returns immediately without taking action.
func (a *service) StopAsyc() {
	const FUNC = "StopAsyc"
	if a.serviceState.state.Stopped() {
		a.logger.Info().Str(logging.FUNC, FUNC).Msg("service is already stopped")
		return
	}
	a.stopTriggered = true
	if a.serviceState.state.New() {
		a.serviceState.Terminated()
		a.logger.Info().Str(logging.FUNC, FUNC).Msg("service was never started")
		return
	}
	func() {
		defer commons.IgnorePanic()
		close(a.stopTrigger)
	}()

	a.logger.Info().Str(logging.FUNC, FUNC).Dict(logging.EVENT, STOP_TRIGGERED.Dict()).Msg("")
}

// Stop invokes StopAsync() followed by AwaitUntilStopped()
func (a *service) Stop() {
	if a.State().Stopped() {
		return
	}
	a.StopAsyc()
	a.AwaitUntilStopped()
}

// StopTriggered returns true if the service was triggered to stop.
func (a *service) StopTriggered() bool {
	return a.stopTriggered
}

// StopTrigger returns the channel to listen for the stopping
func (a *service) StopTrigger() StopTrigger {
	return a.stopTrigger
}

// Logger returns the service's logger
func (a *service) Logger() zerolog.Logger {
	return a.logger
}

// Dependencies returns the service's dependencies
func (a *service) Dependencies() InterfaceDependencies {
	return a.dependencies
}

func (a *service) String() string {
	return fmt.Sprintf("Service : %v.%v", a.serviceInterface.PkgPath(), a.serviceInterface.Name())
}

/////////////// HealthChecks interface //////////////////////////

func (a *service) HealthChecks() []metrics.HealthCheck {
	return a.healthchecks
}

// FailedHealthChecks returns all failed health checks based on the last result.
// If a HealthCheck has not yet been run, then it will be run.
func (a *service) FailedHealthChecks() []metrics.HealthCheck {
	failed := []metrics.HealthCheck{}
	for _, healthcheck := range a.healthchecks {
		if healthcheck.LastResult() != nil && !healthcheck.LastResult().Success() {
			failed = append(failed, healthcheck)
		}
	}
	return failed
}

// SucceededHealthChecks returns all health checks that succeeded based on the last result.
// If a HealthCheck has not yet been run, then it will be run.
func (a *service) SucceededHealthChecks() []metrics.HealthCheck {
	succeeded := []metrics.HealthCheck{}
	for _, healthcheck := range a.healthchecks {
		if healthcheck.LastResult() != nil && healthcheck.LastResult().Success() {
			succeeded = append(succeeded, healthcheck)
		}
	}
	return succeeded
}

// RunAllHealthChecks runs all registered health checks.
// After a health check is run it is delivered on the channel.
func (a *service) RunAllHealthChecks() <-chan metrics.HealthCheck {
	count := len(a.healthchecks)
	c := make(chan metrics.HealthCheck, count)
	wait := sync.WaitGroup{}
	wait.Add(count)

	for _, healthcheck := range a.healthchecks {
		go func(healthcheck metrics.HealthCheck) {
			defer func() {
				c <- healthcheck
				wait.Done()
			}()
			healthcheck.Run()
		}(healthcheck)
	}

	go func() {
		wait.Wait()
		close(c)
	}()

	return c
}

// RunAllFailedHealthChecks all failed health checks based on the last result
func (a *service) RunAllFailedHealthChecks() <-chan metrics.HealthCheck {
	failedHealthChecks := a.FailedHealthChecks()
	c := make(chan metrics.HealthCheck, len(failedHealthChecks))
	wait := sync.WaitGroup{}
	wait.Add(len(failedHealthChecks))

	for _, healthcheck := range a.FailedHealthChecks() {
		go func(healthcheck metrics.HealthCheck) {
			defer func() {
				c <- healthcheck
				wait.Done()
			}()
			healthcheck.Run()
		}(healthcheck)
	}

	go func() {
		wait.Wait()
		close(c)
	}()

	return c
}

/////////////// END - HealthChecks interface //////////////////////////

// StopTrigger is used to notify the service to stop.
// Closing the channel is the stop signal
type StopTrigger <-chan struct{}

// ServiceConstructor returns a new instance of a Service in the New state
type ServiceConstructor func() Service

// ServiceReference represents something that has a reference to a service.
type ServiceReference interface {
	Service() Service
}
