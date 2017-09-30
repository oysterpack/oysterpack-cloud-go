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
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/oysterpack/oysterpack.go/pkg/metrics"
	"github.com/oysterpack/oysterpack.go/pkg/service"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// reporter reports prometheus metrics via HTTP
// endpoint : /metrics
type reporter struct {
	*service.RestartableService

	httpServer *http.Server
}

func (a *reporter) Registry() *prometheus.Registry {
	return metrics.Registry
}

func (a *reporter) init(ctx *service.Context) error {
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

func (a *reporter) destroy(ctx *service.Context) error {
	background := context.Background()
	shutdownContext, cancel := context.WithTimeout(background, time.Second*30)
	defer cancel()
	if err := a.httpServer.Shutdown(shutdownContext); err != nil {
		a.Service().Logger().Error().Err(err).Msg("")
	}
	a.httpServer = nil
	return nil
}

// Println implements promhttp.Logger interface.
// It is used to log any errors reported by the prometheus http handler
func (a *reporter) Println(v ...interface{}) {
	a.Service().Logger().Error().Msg(fmt.Sprint(v))
}
