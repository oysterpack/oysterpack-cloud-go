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
	"net/http"

	"fmt"

	"context"
	"time"

	"github.com/Masterminds/semver"
	"github.com/oysterpack/oysterpack.go/oysterpack/commons/os"
	"github.com/oysterpack/oysterpack.go/oysterpack/commons/reflect"
	"github.com/oysterpack/oysterpack.go/oysterpack/service"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type Interface interface {
	Registry() *prometheus.Registry
}

var MetricsServiceInterface service.ServiceInterface = func() service.ServiceInterface {
	var c Interface = &client{}
	serviceInterface, err := reflect.ObjectInterface(&c)
	if err != nil {
		panic(err)
	}
	return serviceInterface
}()

type client struct {
	*service.RestartableService

	metricRegistry *prometheus.Registry
	httpServer     *http.Server
}

func (a *client) Registry() *prometheus.Registry {
	return a.metricRegistry
}

func (a *client) init(ctx *service.Context) error {
	if a.metricRegistry == nil {
		a.metricRegistry = prometheus.NewRegistry()
		a.metricRegistry.MustRegister(
			prometheus.NewGoCollector(),
			prometheus.NewProcessCollector(os.PID(), ""),
		)
	}

	metricsHandler := promhttp.HandlerFor(
		a.metricRegistry,
		promhttp.HandlerOpts{
			ErrorLog:      a,
			ErrorHandling: promhttp.ContinueOnError,
		},
	)
	http.Handle("/metrics", metricsHandler)
	a.httpServer = &http.Server{
		Addr: fmt.Sprintf(":%d", Config.HTTPPort()),
	}
	go a.httpServer.ListenAndServe()

	return nil
}

func (a *client) destroy(ctx *service.Context) error {
	shutdownContext := context.Background()
	shutdownContext, _ = context.WithTimeout(shutdownContext, time.Second*30)
	if err := a.httpServer.Shutdown(shutdownContext); err != nil {
		a.Service().Logger().Error().Err(err).Msg("")
	}
	a.httpServer = nil
	return nil
}

// Println implements promhttp.Logger interface
func (a *client) Println(v ...interface{}) {
	a.Service().Logger().Error().Msg(fmt.Sprint(v))
}

func (a *client) newService() service.Service {
	version, err := semver.NewVersion("1.0.0")
	if err != nil {
		panic(err)
	}

	settings := service.ServiceSettings{
		ServiceInterface: MetricsServiceInterface,
		Version:          version,
		Init:             a.init,
		Destroy:          a.destroy,
	}
	return service.NewService(settings)
}

func NewClient(app service.Application) service.Client {
	c := &client{}
	c.RestartableService = service.NewRestartableService(c.newService)
	return c
}
