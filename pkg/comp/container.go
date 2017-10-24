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
	"errors"
	"fmt"
	stdreflect "reflect"

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

// NewContainer creates and returns a new Container
func NewContainer(desc *Descriptor, config func(Component) *capnp.Message) *Container {
	c := &Container{
		desc:   desc,
		config: config,

		serverChan: make(chan interface{}),

		comps:       make(map[Interface]Component),
		compLookups: make(map[Interface][]chan Component),
	}
	registerContainerHealthChecks(c)
	c.Go(c.server)
	return c
}

func registerContainerHealthChecks(c *Container) {
	gauge := prometheus.GaugeOpts{
		Namespace:   metrics.METRIC_NAMESPACE_OYSTERPACK,
		Subsystem:   "application",
		Name:        "check_service_dependencies",
		Help:        "The healthcheck fails if any dependencies are not satisfied.",
		ConstLabels: c.desc.AddMetricLabels(prometheus.Labels{}),
	}
	runInterval := time.Minute
	healthcheck := func() error { return CheckAllDependencies(c) }
	metrics.NewHealthCheck(gauge, runInterval, healthcheck)
}

// Container is used to manage a set of Component(s). The container runs in its own goroutine.
// Its lifetime is the goroutine's lifetime. When the container is killed, its goroutine and all sub-goroutines are notified
// to terminate.
//
// The container shutdown can be triggered via:
//	1. Container.Stop()
//  2. OS signals : syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT
//  3. One of the registered compononets fails.
//
// If a ShutdownHook returns an error, then it is simply logged and moves on.
// The container is designed to be concurrency safe. It uses the classic client-server design pattern using channels to communicate between goroutines.
// The container state is managed by the backend server goroutine. All mutable requests are sent to the server via the serverChan.
// The container's lifetime is defined by the server goroutine's lifetime.
// All container goroutines are tracked and managed by the Tomb. When the Tomb is killed, the container is shutdown.
//
// Container goroutines:
//  1. server goroutine - main backend goroutine
//  2. component goroutines :
//     1. a goroutine to start the component
//     2. after the component is started, a goroutine is used to monitor the component's lifecycle - if the component fails, then trigger the container shutdown
//     3. a goroutine that will shutdown the component after it is notified that the container is stopping
//  3. A goroutine per shutdown hook that waits until the container is stopping to run the shutdown hook
type Container struct {
	desc   *Descriptor
	config func(Component) *capnp.Message

	tomb.Tomb
	serverChan chan interface{}

	comps map[Interface]Component

	compLookups map[Interface][]chan Component
}

// Desc returns the container Descriptor, which is used to name the set of components that are packaged into this container.
func (a *Container) Desc() *Descriptor {
	return a.desc
}

// NotifyStopped returns a channel that can be used to be notified when
func (a *Container) NotifyStopped() <-chan struct{} {
	return a.Dead()
}

// Kill performs the following :
//  1. Cancels all scheduled HealthCheck(s)
//  2. The metrics and health check registries are reset.
//  3. Kills the container
func (a *Container) Kill(reason error) {
	if !a.Alive() {
		return
	}
	defer metrics.ResetRegistry()
	// stop all healthchecks before shutting down because if healthchecks fail during container shutdown, then false healthcheck failure may be reported
	metrics.HealthChecks.StopAllHealthCheckTickers()
	a.Tomb.Kill(nil)
}

func (a *Container) Go(f func(stopping <-chan struct{}) error) {
	a.Tomb.Go(func() error {
		return f(a.Dying())
	})
}

func (a *Container) server(stopping <-chan struct{}) error {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)

	for {
		select {
		case <-stopping:
			CONTAINER_KILLED.Log(logger.Info()).Msg("")
			return nil
		case sig := <-sigs:
			OS_SIGNAL.Log(logger.Info()).Str(logging.SIGNAL, sig.String()).Msg("")
			a.Kill(nil)
		case msg := <-a.serverChan:
			switch req := msg.(type) {
			case registerComponent:
				comp, err := a.register(req.newComp)
				if err != nil {
					req.err <- err
				} else {
					req.comp <- comp
				}
			case getComponent:
				req.response <- a.component(req.comp)
			case getComponents:
				req.response <- a.components()
			default:
				logger.Panic().Msgf("Unhandled message : %T", req)
			}
		}
	}
}

type registerComponent struct {
	newComp ComponentConstructor
	comp    chan Component
	err     chan error
}

func (a *Container) MustRegister(newComp ComponentConstructor) Component {
	comp, err := a.Register(newComp)
	if err != nil {
		logger.Panic().Str(logging.TYPE, TYPE_REGISTRY).Str(logging.FUNC, "MustRegister").Err(err).Msg("")
	}
	return comp
}

func (a *Container) Register(newComp ComponentConstructor) (Component, error) {
	req := registerComponent{newComp, make(chan Component), make(chan error)}
	a.serverChan <- req
	select {
	case comp := <-req.comp:
		return comp, nil
	case err := <-req.err:
		return nil, err
	}
}

func (a *Container) register(newComp ComponentConstructor) (Component, error) {
	comp := newComp()

	if comp.Desc() == nil {
		return nil, errors.New("Component has no Descriptor")
	}

	if comp.Interface() == nil {
		desc := comp.Desc()
		return nil, fmt.Errorf("Component has no Interface : %s/%s/%s/%s", desc.namespace, desc.system, desc.component, desc.version)
	}

	checkInterface(comp.Interface())

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

func (a *Container) startComponent(comp Component) {
	a.Go(func(stopping <-chan struct{}) error {
		for {
			select {
			case <-stopping:
				return nil
			case <-comp.Start(a.config(comp), a):
				a.compStarted(comp)

				// monitor the component life cycle
				// if the component fails, then fail fast and kill the container
				a.Go(func(stopping <-chan struct{}) error {
					for {
						select {
						case <-stopping:
							return nil
						case stateChange := <-comp.StateChan():
							COMP_STATE_CHANGE.Log(comp.Logger().Info()).Str("state", stateChange.State.String()).Msg("")
							if comp.State().Terminal() {
								err := comp.FailureCause()
								if err != nil {
									COMP_FAILED.Log(comp.Logger().Error()).Err(err).Msg("")
								}
								return err
							}
						}
					}
				})
				return nil
			}
		}
	})

	// when the container is killed, then stop the registered component
	//a.Go(func(stopping <-chan struct{}) error {
	//	<-stopping
	//	<-comp.Stop()
	//	return nil
	//})
}

func (a *Container) compStarted(comp Component) {
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

type getComponent struct {
	comp     Interface
	response chan chan Component
}

func (a *Container) Component(comp Interface) <-chan Component {
	req := &getComponent{comp, make(chan chan Component)}
	a.serverChan <- req
	return <-req.response
}

func (a *Container) component(comp Interface) chan Component {
	c := make(chan Component, 1)

	if component := a.comps[comp]; component.State().Running() {
		c <- component
	} else {
		lookups := a.compLookups[comp]
		a.compLookups[comp] = append(lookups, c)
	}

	return c
}

type getComponents struct {
	response chan []Component
}

func (a *Container) Components() []Component {
	req := &getComponents{make(chan []Component)}
	a.serverChan <- req
	return <-req.response
}

func (a *Container) components() []Component {
	comps := make([]Component, len(a.comps))
	for _, comp := range a.comps {
		comps = append(comps, comp)
	}
	return comps
}

func (a *Container) RegisterShutdownHook(name string, f func() error) {
	const SHUTDOWNHOOK = "ShutdownHook"
	hook := func(stopping <-chan struct{}) error {
		defer func() {
			if p := recover(); p != nil {
				logger.Error().Err(fmt.Errorf("%v", p)).Str(SHUTDOWNHOOK, name).Msg("panic")
			}
		}()

		<-stopping

		if err := f(); err != nil {
			logger.Error().Err(err).Str(SHUTDOWNHOOK, name).Msg("")
		} else {
			logger.Info().Str(SHUTDOWNHOOK, name).Msg("success")
		}
		return nil
	}
	a.Go(hook)
	logger.Info().Str(SHUTDOWNHOOK, name).Msg("registered")
}
