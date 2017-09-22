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

package http

import (
	"net/http"

	"fmt"

	"context"
	"time"

	"github.com/Masterminds/semver"
	"github.com/oysterpack/oysterpack.go/pkg/commons/reflect"
	"github.com/oysterpack/oysterpack.go/pkg/metrics"
	"github.com/oysterpack/oysterpack.go/pkg/service"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Service interface
type Service interface {
	Registry() *prometheus.Registry
}

// MetricsServiceInterface service interface instance that can be used to lookup the registered service
var MetricsServiceInterface service.ServiceInterface = func() service.ServiceInterface {
	var c Service = &server{}
	serviceInterface, err := reflect.ObjectInterface(&c)
	if err != nil {
		panic(err)
	}
	return serviceInterface
}()

type server struct {
	*service.RestartableService

	httpServer *http.Server
}

func (a *server) Registry() *prometheus.Registry {
	return metrics.Registry
}

func (a *server) init(ctx *service.Context) error {
	metricsHandler := promhttp.HandlerFor(
		a.Registry(),
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

func (a *server) destroy(ctx *service.Context) error {
	background := context.Background()
	shutdownContext, cancel := context.WithTimeout(background, time.Second*30)
	defer cancel()
	if err := a.httpServer.Shutdown(shutdownContext); err != nil {
		a.Service().Logger().Error().Err(err).Msg("")
	}
	a.httpServer = nil
	return nil
}

// Println implements promhttp.Logger interface
func (a *server) Println(v ...interface{}) {
	a.Service().Logger().Error().Msg(fmt.Sprint(v))
}

func (a *server) newService() service.Service {
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

// NewClient service.ClientConstructor
func NewClient(app service.Application) service.Client {
	c := &server{}
	c.RestartableService = service.NewRestartableService(c.newService)
	return c
}

func init() {
	service.App().MustRegisterService(NewClient)
}
