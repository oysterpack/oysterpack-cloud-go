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

package app

import (
	"os"
	"sync"

	"fmt"
	"net/http"

	"context"
	"time"

	"github.com/oysterpack/oysterpack.go/pkg/app/config"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

const (
	METRICS_SERVICE_ID               = ServiceID(0xe3054017c1b1d214)
	DEFAULT_METRICS_HTTP_PORT uint16 = 4444
)

var (
	metricsMutex sync.RWMutex

	// Registry is the global registry
	registry = NewMetricsRegistry(true)
)

// NewRegistry creates a new registry.
// If collectProcessMetrics = true, then the prometheus GoCollector and ProcessCollectors are registered.
func NewMetricsRegistry(collectProcessMetrics bool) *prometheus.Registry {
	registry := prometheus.NewRegistry()
	if collectProcessMetrics {
		registry.MustRegister(
			prometheus.NewGoCollector(),
			prometheus.NewProcessCollector(os.Getpid(), ""),
		)
	}
	return registry
}

func MetricsRegistry() *prometheus.Registry {
	metricsMutex.RLock()
	defer metricsMutex.RUnlock()
	return registry
}

// ResetMetricsRegistry creates a new registry.
// The main use case is for unit testing.
func ResetMetricsRegistry() {
	metricsMutex.Lock()
	defer metricsMutex.Unlock()
	registry = NewMetricsRegistry(true)
}

// if the metrics HTTP server fails to start, then this is considered a fatal error, which will terminate the process.
func startMetricsHttpReporter() {
	svc := NewService(METRICS_SERVICE_ID)
	if err := RegisterService(svc); err != nil {
		METRICS_HTTP_REPORTER_START_ERROR.Log(Logger().Fatal()).Err(err).Msg("")
	}
	metricsService := &metricsHttpReporter{
		Service: svc,
		port:    metricsReporterHttpPort(),
	}

	if err := metricsService.start(); err != nil {
		METRICS_HTTP_REPORTER_START_ERROR.Log(Logger().Fatal()).Err(err).Msg("")
	}
}

func metricsReporterHttpPort() uint16 {
	cfg, err := Config(METRICS_SERVICE_ID)
	if err != nil {
		switch err := err.(type) {
		case ServiceConfigNotExistError:
			return DEFAULT_METRICS_HTTP_PORT
		default:
			METRICS_HTTP_REPORTER_CONFIG_ERROR.Log(Logger().Fatal()).Err(err).Msg("")
		}
	}
	metricsServiceConfig, err := config.ReadRootMetricsServiceSpec(cfg)
	if err != nil {
		METRICS_HTTP_REPORTER_CONFIG_ERROR.Log(Logger().Fatal()).Err(err).Msg("")
	}
	if metricsServiceConfig.HttpPort() == 0 {
		return DEFAULT_METRICS_HTTP_PORT
	}
	return metricsServiceConfig.HttpPort()
}

// metricsHttpReporter is used to expose metrics to Prometheus via HTTP
type metricsHttpReporter struct {
	sync.Mutex
	*Service
	httpServer *http.Server
	port       uint16
}

func (a *metricsHttpReporter) start() error {
	if !a.Alive() {
		return ErrServiceNotAlive
	}
	a.Lock()
	defer a.Unlock()
	if a.httpServer != nil {
		return nil
	}
	a.startHttpServer()
	a.httpServer.RegisterOnShutdown(func() {
		if a.Alive() {
			METRICS_HTTP_REPORTER_SHUTDOWN_WHILE_SERVICE_RUNNING.Log(a.Logger().Error()).Msg("Metrics HTTP server shutdown, but service is still alive.")
			a.Lock()
			defer a.Unlock()
			a.httpServer.Close()
			a.startHttpServer()
		}
	})
	a.Go(func() error {
		<-a.Dying()
		a.stop()
		return nil
	})
	return nil
}

func (a *metricsHttpReporter) startHttpServer() {
	metricsHandler := promhttp.HandlerFor(
		MetricsRegistry(),
		promhttp.HandlerOpts{
			ErrorLog:      a,
			ErrorHandling: promhttp.ContinueOnError,
		},
	)
	http.Handle("/metrics", metricsHandler)
	a.httpServer = &http.Server{
		Addr: fmt.Sprintf(":%d", a.port),
	}
	go a.httpServer.ListenAndServe()
	a.Logger().Info().Str("addr", a.httpServer.Addr).Msg("Metrics HTTP server started")
}

func (a *metricsHttpReporter) stop() {
	a.Lock()
	defer a.Unlock()
	background := context.Background()
	shutdownContext, cancel := context.WithTimeout(background, time.Second*10)
	defer cancel()
	if err := a.httpServer.Shutdown(shutdownContext); err != nil {
		METRICS_HTTP_REPORTER_SHUTDOWN_ERROR.Log(a.Logger().Warn()).Err(NewMetricsServiceError(err)).Msg("Error during HTTP server shutdown.")
		a.httpServer.Close()
	}
	a.httpServer = nil
}

// Println implements promhttp.Logger interface.
// It is used to log any errors reported by the prometheus http handler
func (a *metricsHttpReporter) Println(v ...interface{}) {
	a.Logger().Error().Msg(fmt.Sprint(v))
}
