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
	"io"

	"errors"
	"fmt"
	stdreflect "reflect"

	"runtime"

	"github.com/oysterpack/oysterpack.go/pkg/logging"
	"gopkg.in/tomb.v2"
	"zombiezen.com/go/capnproto2"

	"os"
	"os/signal"
	"syscall"

	"time"

	"github.com/oysterpack/oysterpack.go/pkg/metrics"
	"github.com/prometheus/client_golang/prometheus"
)

// Container is used to manage a set of Component(s).
//
// The container shutdown can be triggered via:
//	1. Container.Stop()
//  2. OS signals : syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT
type Container interface {
	Desc() *Descriptor

	Registry

	// DumpAllStacks dumps all goroutine stacks to the specified writer.
	// Used for troubleshooting purposes.
	DumpAllStacks(out io.Writer)

	// Stop performs the following :
	//  1. Cancels all scheduled HealthCheck(s)
	//  2. Stops all registered components. Components will be stopped in the reverse order they were started in, i.e.,
	//     the first component that is started will be the last one that is stopped.
	//  3. Runs all registered ShutdownHook(s) in the reverse order they were registered.
	//  4. Cancels all registered goroutines
	//  5. The metrics and health check registries are reset.
	Stop() error

	// State the current container lifecycle state
	Alive() bool

	NotifyStopped() <-chan struct{}
}

// NewContainer creates and returns a new Container
func NewContainer(desc *Descriptor, config func(Component) *capnp.Message) Container {
	c := &container{
		desc:   desc,
		config: config,

		serverChan: make(chan interface{}),

		comps:       make(map[Interface]Component),
		compLookups: make(map[Interface][]chan Component),
	}
	registerContainerHealthChecks(c)
	c.t.Go(c.server)
	return c
}

func registerContainerHealthChecks(c *container) {
	gauge := prometheus.GaugeOpts{
		Namespace:   metrics.METRIC_NAMESPACE_OYSTERPACK,
		Subsystem:   "application",
		Name:        "check_service_dependencies",
		Help:        "The healthcheck fails if any dependencies are not satisfied.",
		ConstLabels: c.desc.AddMetricLabels(prometheus.Labels{}),
	}
	runInterval := time.Minute
	healthcheck := func() error { return c.CheckAllDependencies() }
	c.dependencyHealthCheck = metrics.NewHealthCheck(gauge, runInterval, healthcheck)
}

type container struct {
	desc   *Descriptor
	config func(Component) *capnp.Message

	t          tomb.Tomb
	serverChan chan interface{}

	comps map[Interface]Component

	dependencyHealthCheck metrics.HealthCheck

	compLookups map[Interface][]chan Component
}

func (a *container) Alive() bool {
	return a.t.Alive()
}

func (a *container) Desc() *Descriptor {
	return a.desc
}

func (a *container) CheckAllDependenciesHealthCheck() metrics.HealthCheck {
	return a.dependencyHealthCheck
}

func (a *container) NotifyStopped() <-chan struct{} {
	return a.t.Dead()
}

func (a *container) Stop() error {
	defer metrics.ResetRegistry()

	if !a.Alive() {
		return a.t.Err()
	}
	// stop all healthchecks before shutting down because if healthchecks fail during container shutdown, then false healthcheck failure may be reported
	metrics.HealthChecks.StopAllHealthCheckTickers()
	a.t.Kill(nil)
	return a.t.Wait()
}

func (a *container) server() error {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)

	for {
		select {
		case <-a.t.Dying():
			CONTAINER_KILLED.Log(logger.Info()).Msg("")
			return nil
		case sig := <-sigs:
			OS_SIGNAL.Log(logger.Info()).Str(logging.SIGNAL, sig.String()).Msg("")
			go a.Stop()
		case msg := <-a.serverChan:
			switch req := msg.(type) {
			case RegisterComponent:
				comp, err := a.register(req.newComp)
				if err != nil {
					req.err <- err
				} else {
					req.comp <- comp
				}
			default:
				logger.Panic().Msgf("Unhandled message : %T", req)
			}
		}
	}
}

type RegisterComponent struct {
	newComp ComponentConstructor
	comp    chan Component
	err     chan error
}

func (a *container) MustRegister(newComp ComponentConstructor) Component {
	comp, err := a.Register(newComp)
	if err != nil {
		logger.Panic().Str(logging.TYPE, TYPE_REGISTRY).Str(logging.FUNC, "MustRegister").Err(err).Msg("")
	}
	return comp
}

func (a *container) Register(newComp ComponentConstructor) (Component, error) {
	req := RegisterComponent{newComp, make(chan Component), make(chan error)}
	a.serverChan <- req
	select {
	case comp := <-req.comp:
		return comp, nil
	case err := <-req.err:
		return nil, err
	}
}

func (a *container) register(newComp ComponentConstructor) (Component, error) {
	comp := newComp()

	if comp.Desc() == nil {
		return nil, errors.New("Component has no Descriptor")
	}

	if comp.Interface() == nil {
		desc := comp.Desc()
		return nil, fmt.Errorf("Component has no Interface : %s/%s/%s/%s", desc.namespace, desc.system, desc.component, desc.version)
	}

	if !stdreflect.TypeOf(comp).AssignableTo(comp.Interface()) {
		return nil, fmt.Errorf("The component does not implement the Interface : %T is not assignable to %v", comp, comp.Interface())
	}

	if c, exists := a.comps[comp.Interface()]; exists {
		desc := c.Desc()
		return nil, fmt.Errorf("Another Component is already registered for Interface : %s/%s/%s/%s", desc.namespace, desc.system, desc.component, desc.version)
	}
	a.comps[comp.Interface()] = comp
	a.startComponent(comp)
	return comp, nil
}

func (a *container) startComponent(comp Component) {
	a.t.Go(func() error {
		for {
			select {
			case <-a.t.Dying():
				return nil
			case <-comp.Start(a.config(comp), a):
				a.componentStateChanged(comp)
			}
		}
	})

	// when the container is killed, then stop the registered component
	a.t.Go(func() error {
		<-a.t.Dying()
		<-comp.Stop()
		return comp.FailureCause()
	})
}

func (a *container) componentStateChanged(comp Component) {
	if comp.State().RunningOrLater() {
		for _, c := range a.compLookups[comp.Interface()] {
			c <- comp
			close(c)
		}
		delete(a.compLookups, comp.Interface())
	}

	// since we are here, check all others waiting for Components
	started := []Interface{}
	for i, chans := range a.compLookups {
		if a.comps[i].State().RunningOrLater() {
			for _, ch := range chans {
				ch <- a.comps[i]
				close(ch)
			}
			started = append(started, i)
		}
	}
	for _, i := range started {
		delete(a.compLookups, i)
	}
}

func (a *container) Component(comp Interface) <-chan Component {
	c := make(chan Component, 1)

	if component := a.comps[comp]; component.State().Running() {
		c <- component
	} else {
		lookups := a.compLookups[comp]
		a.compLookups[comp] = append(lookups, c)
	}

	return c
}

func (a *container) Components() <-chan Component {
	c := make(chan Component, len(a.comps))
	go func() {
		for _, comp := range a.comps {
			c <- comp
		}
	}()
	return c
}

func (a *container) RegisterShutdownHook(name string, f func() error) {
	const SHUTDOWNHOOK = "ShutdownHook"
	hook := func() error {
		defer func() {
			if p := recover(); p != nil {
				logger.Error().Err(fmt.Errorf("%v", p)).Str(SHUTDOWNHOOK, name).Msg("panic")
			}
		}()

		<-a.t.Dying()

		err := f()
		if err != nil {
			logger.Error().Err(err).Str(SHUTDOWNHOOK, name).Msg("")
		} else {
			logger.Info().Str(SHUTDOWNHOOK, name).Msg("success")
		}
		return err
	}
	a.t.Go(hook)
	logger.Info().Str(SHUTDOWNHOOK, name).Msg("registered")
}

func (a *container) DumpAllStacks(out io.Writer) {
	size := 1024 * 8
	for {
		buf := make([]byte, size)
		if len := runtime.Stack(buf, true); len <= size {
			out.Write(buf[:len])
			return
		}
		size = size + (1024 * 8)
	}
}

// CheckAllDependenciesRegistered checks that are  Dependencies are currently satisfied.
func (a *container) CheckAllDependenciesRegistered() []*DependencyMissingError {
	errors := []*DependencyMissingError{}
	for _, comp := range a.comps {
		if missingDependencies := a.CheckDependenciesRegistered(comp); missingDependencies != nil {
			errors = append(errors, missingDependencies)
		}
	}
	return errors
}

// CheckDependenciesRegistered checks that the 's Dependencies are currently satisfied
// nil is returned if there is no error
func (a *container) CheckDependenciesRegistered(comp Component) *DependencyMissingError {
	missingDependencies := &DependencyMissingError{&DependencyMappings{Interface: comp.Interface()}}
	for dependency, constraints := range comp.Dependencies() {
		b, exists := a.comps[dependency]
		if !exists {
			missingDependencies.AddMissingDependency(dependency)
		} else {
			if constraints != nil {
				if !constraints.Check(b.Desc().Version()) {
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

// CheckAllDependenciesRunning checks that are  Dependencies are currently satisfied.
func (a *container) CheckAllDependenciesRunning() []*DependencyNotRunningError {
	errors := []*DependencyNotRunningError{}
	for _, comp := range a.comps {
		if notRunning := a.CheckDependenciesRunning(comp); notRunning != nil {
			errors = append(errors, notRunning)
		}
	}
	return errors
}

// CheckDependenciesRunning checks that the 's Dependencies are currently satisfied
// nil is returned if there is no error
func (a *container) CheckDependenciesRunning(comp Component) *DependencyNotRunningError {
	notRunning := &DependencyNotRunningError{&DependencyMappings{Interface: comp.Interface()}}
	for dependency, constraints := range comp.Dependencies() {
		if compDependency, exists := a.comps[dependency]; !exists ||
			(constraints != nil && !constraints.Check(compDependency.Desc().Version())) ||
			!compDependency.State().Running() {
			notRunning.AddDependencyNotRunning(dependency)
		}
	}
	if notRunning.HasNotRunning() {
		return notRunning
	}
	return nil
}

// CheckAllDependencies checks that all Dependencies for each  are available
// nil is returned if there are no errors
func (a *container) CheckAllDependencies() *DependencyErrors {
	errors := []error{}
	for _, err := range a.CheckAllDependenciesRegistered() {
		errors = append(errors, err)
	}

	for _, err := range a.CheckAllDependenciesRunning() {
		errors = append(errors, err)
	}

	if len(errors) > 0 {
		return &DependencyErrors{errors}
	}
	return nil
}

// CheckDependencies checks that the  Dependencies are available
func (a *container) CheckDependencies(comp Component) *DependencyErrors {
	errors := []error{}
	if err := a.CheckDependenciesRegistered(comp); err != nil {
		errors = append(errors, err)
	}

	if err := a.CheckDependenciesRunning(comp); err != nil {
		errors = append(errors, err)
	}

	if len(errors) > 0 {
		return &DependencyErrors{errors}
	}
	return nil
}
