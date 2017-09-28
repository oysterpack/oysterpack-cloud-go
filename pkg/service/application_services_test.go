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

package service_test

import (
	"time"

	"github.com/Masterminds/semver"
	"github.com/oysterpack/oysterpack.go/pkg/commons/reflect"
	"github.com/oysterpack/oysterpack.go/pkg/metrics"
	"github.com/oysterpack/oysterpack.go/pkg/service"
	"github.com/prometheus/client_golang/prometheus"
)

// EchoService defines the Service interface
type EchoService interface {
	Echo(msg interface{}) interface{}
}

var (
	EchoServiceInterface reflect.InterfaceType = echoServiceInterfaceType()
	desc                                       = service.NewDescriptor("oysterpack", "test", "echo", "1.0.0", EchoServiceInterface)
)

func echoServiceInterfaceType() reflect.InterfaceType {
	var echoServicePrototype EchoService = &EchoServiceClient{}
	t, err := reflect.ObjectInterface(&echoServicePrototype)
	if err != nil {
		panic(err)
	}
	return t
}

// EchoServiceClient is the EchoService implementation
type EchoServiceClient struct {
	// embedded service
	*service.RestartableService

	// service channels
	echo chan *EchoRequest
}

////////// service messages //////////////////

type EchoRequest struct {
	Message interface{}
	ReplyTo chan interface{}
}

// message constructor
func NewEchoRequest(msg interface{}) *EchoRequest {
	return &EchoRequest{msg, make(chan interface{}, 1)}
}

////////// service methods //////////////////

// Echo
func (a *EchoServiceClient) Echo(msg interface{}) interface{} {
	req := NewEchoRequest(msg)
	a.echo <- req
	return <-req.ReplyTo
}

//////// service metrics ///////////////

var (
	counterOpts = &prometheus.CounterOpts{
		Name:        "msgs_processed",
		Help:        "The number of messages that have been processed",
		ConstLabels: service.AddServiceMetricLabels(prometheus.Labels{}, desc),
	}
	counter = metrics.GetOrMustRegisterCounter(counterOpts)

	counterVecOpts = metrics.NewCounterVecOpts(
		&prometheus.CounterOpts{
			Name:        "echo_counter_vec",
			Help:        "The number of messages that have been processed",
			ConstLabels: service.AddServiceMetricLabels(prometheus.Labels{}, desc),
		}, "success")
	counterVec = metrics.GetOrMustRegisterCounterVec(counterVecOpts)

	gaugeOpts = &prometheus.GaugeOpts{
		Name:        "echo_gauge",
		Help:        "The number of messages that have been processed",
		ConstLabels: service.AddServiceMetricLabels(prometheus.Labels{}, desc),
	}
	gauge = metrics.GetOrMustRegisterGauge(gaugeOpts)

	gaugeVecOpts = metrics.NewGaugeVecOpts(
		&prometheus.GaugeOpts{
			Name:        "echo_gauge_vec",
			Help:        "The number of messages that have been processed",
			ConstLabels: service.AddServiceMetricLabels(prometheus.Labels{}, desc),
		}, "success")
	gaugeVec = metrics.GetOrMustRegisterCounterVec(counterVecOpts)

	histOpts = &prometheus.HistogramOpts{
		Name:        "echo_hist",
		Help:        "The number of messages that have been processed",
		ConstLabels: service.AddServiceMetricLabels(prometheus.Labels{}, desc),
		Buckets:     prometheus.DefBuckets,
	}
	hist = metrics.GetOrMustRegisterGauge(gaugeOpts)

	histVecOpts = metrics.NewHistogramVecOpts(
		&prometheus.HistogramOpts{
			Name:        "echo_hist_vec",
			Help:        "The number of messages that have been processed",
			ConstLabels: service.AddServiceMetricLabels(prometheus.Labels{}, desc),
			Buckets:     prometheus.DefBuckets,
		}, "success")
	histVec = metrics.GetOrMustRegisterHistogramVec(histVecOpts)

	summaryOpts = &prometheus.SummaryOpts{
		Name:        "echo_summary",
		Help:        "The number of messages that have been processed",
		ConstLabels: service.AddServiceMetricLabels(prometheus.Labels{}, desc),
		Objectives:  prometheus.DefObjectives,
	}
	summary = metrics.GetOrMustRegisterSummary(summaryOpts)

	summaryVecOpts = metrics.NewSummaryVecOpts(
		&prometheus.SummaryOpts{
			Name:        "echo_summary_vec",
			Help:        "The number of messages that have been processed",
			ConstLabels: service.AddServiceMetricLabels(prometheus.Labels{}, desc),
			Objectives:  prometheus.DefObjectives,
		}, "success")
	summaryVec = metrics.GetOrMustRegisterSummaryVec(summaryVecOpts)

	metricOpts = &metrics.MetricOpts{
		CounterOpts:    []*prometheus.CounterOpts{counterOpts},
		CounterVecOpts: []*metrics.CounterVecOpts{counterVecOpts},

		GaugeOpts:    []*prometheus.GaugeOpts{gaugeOpts},
		GaugeVecOpts: []*metrics.GaugeVecOpts{gaugeVecOpts},

		HistogramOpts:    []*prometheus.HistogramOpts{histOpts},
		HistogramVecOpts: []*metrics.HistogramVecOpts{histVecOpts},

		SummaryOpts:    []*prometheus.SummaryOpts{summaryOpts},
		SummaryVecOpts: []*metrics.SummaryVecOpts{summaryVecOpts},
	}
)

////////// service constructor //////////////////

// ServiceConstructor
func (a *EchoServiceClient) newService() service.Service {
	return service.NewService(service.Settings{
		Descriptor: desc,
		Run:        a.run,
		Metrics:    metricOpts,
	})
}

func (a *EchoServiceClient) init(ctx *service.Context) error {
	if a.echo == nil {
		a.echo = make(chan *EchoRequest)
	}

	return nil
}

// Service Run func
func (a *EchoServiceClient) run(ctx *service.Context) error {
	for {
		select {
		case req := <-a.echo:
			counter.Inc()
			counterVec.WithLabelValues("true")
			go func() {
				req.ReplyTo <- req.Message
			}()
		case <-ctx.StopTrigger():
			return nil
		}
	}
}

// EchoServiceClientConstructor is a ClientConstructor
func EchoServiceClientConstructor(application service.Application) service.Client {
	serviceClient := &EchoServiceClient{echo: make(chan *EchoRequest)}
	serviceClient.RestartableService = service.NewRestartableService(serviceClient.newService)
	return serviceClient
}

///////////////////////////

type HeartbeatService interface {
	Ping() time.Duration
}

var HeartbeatServiceInterface reflect.InterfaceType = heartbeatServiceInterfaceType()

func heartbeatServiceInterfaceType() reflect.InterfaceType {
	var prototype HeartbeatService = &HeartbeatServiceClient{}
	t, err := reflect.ObjectInterface(&prototype)
	if err != nil {
		panic(err)
	}
	return t
}

type HeartbeatServiceClient struct {
	*service.RestartableService

	pingChan chan *PingRequest
}

type PingRequest struct {
	time.Time
	replyTo chan time.Duration
}

func (a *HeartbeatServiceClient) Ping() time.Duration {
	req := &PingRequest{time.Now(), make(chan time.Duration, 1)}
	a.pingChan <- req
	return <-req.replyTo
}

func (a *HeartbeatServiceClient) run(ctx *service.Context) error {
	for {
		select {
		case <-ctx.StopTrigger():
			return nil
		case req := <-a.pingChan:
			req.replyTo <- time.Since(req.Time)
		}
	}
}

func (a *HeartbeatServiceClient) newService() service.Service {
	return service.NewService(service.Settings{
		Descriptor: service.NewDescriptor("oysterpack", "test", "heartbeat", "1.0.0", HeartbeatServiceInterface),
		Run:        a.run,
	})
}

func HeartbeatServiceClientConstructor(application service.Application) service.Client {
	serviceClient := &HeartbeatServiceClient{
		pingChan: make(chan *PingRequest),
	}
	serviceClient.RestartableService = service.NewRestartableService(serviceClient.newService)
	return serviceClient
}

////////////////////////////

type AService interface{}

type AServiceClient struct {
	*service.RestartableService
}

var AServiceInterface reflect.InterfaceType = aServiceInterfaceType()

func aServiceInterfaceType() reflect.InterfaceType {
	var prototype AService = &AServiceClient{}
	t, err := reflect.ObjectInterface(&prototype)
	if err != nil {
		panic(err)
	}
	return t
}

func AServiceClientConstructorFactory(version string) service.ClientConstructor {
	return func(application service.Application) service.Client {
		serviceClient := &AServiceClient{}
		serviceClient.RestartableService = service.NewRestartableService(func() service.Service {
			return service.NewService(service.Settings{Descriptor: service.NewDescriptor("oysterpack", "test", "aservice", version, AServiceInterface)})
		})
		return serviceClient
	}
}

////////////////////////////////

type BService interface{}

type BServiceClient struct {
	*service.RestartableService
}

var BServiceInterface reflect.InterfaceType = bServiceInterfaceType()

func bServiceInterfaceType() reflect.InterfaceType {
	var prototype BService = &BServiceClient{}
	t, err := reflect.ObjectInterface(&prototype)
	if err != nil {
		panic(err)
	}
	return t
}

func BServiceClientConstructor(app service.Application) service.Client {
	serviceClient := &AServiceClient{}
	serviceClient.RestartableService = service.NewRestartableService(func() service.Service {

		var init service.Init = func(ctx *service.Context) error {
			select {
			case <-app.ServiceByTypeAsync(AServiceInterface).Channel():
				return nil
			case <-ctx.StopTrigger():
				return nil
			}
		}

		aServiceVersionConstraint, err := semver.NewConstraint(">= 1.0, < 2")
		if err != nil {
			panic(err)
		}

		return service.NewService(service.Settings{
			Descriptor:            service.NewDescriptor("oysterpack", "test", "bservice", "1.0.0", BServiceInterface),
			InterfaceDependencies: service.InterfaceDependencies{AServiceInterface: aServiceVersionConstraint},
			Init: init,
		})
	})
	return serviceClient
}
