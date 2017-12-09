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
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/oysterpack/oysterpack.go/pkg/app/config"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"zombiezen.com/go/capnproto2"
)

// if the metrics HTTP server fails to start, then this is considered a fatal error, which will terminate the process.
func startMetricsHttpReporter() {
	svc := NewService(METRICS_SERVICE_ID)
	if err := Services.Register(svc); err != nil {
		METRICS_HTTP_REPORTER_START_ERROR.Log(Logger().Panic()).Err(err).Msg("")
	}

	metricsService := &metricsHttpReporter{
		Service: svc,
		port:    MetricsServiceSpec().HttpPort(),
	}

	if err := metricsService.start(); err != nil {
		METRICS_HTTP_REPORTER_START_ERROR.Log(Logger().Panic()).Err(err).Msg("")
	}
}

func MetricsServiceSpec() config.MetricsServiceSpec {
	cfg, err := Configs.Config(METRICS_SERVICE_ID)
	if err != nil {
		METRICS_SERVICE_CONFIG_ERROR.Log(Logger().Panic()).Err(err).Msg("")
	}
	if cfg == nil {
		_, s, capnpErr := capnp.NewMessage(capnp.SingleSegment(nil))
		if capnpErr != nil {
			CAPNP_ERR.Log(Logger().Panic()).Err(err).Msg("MetricsServiceSpec() - capnp.NewMessage(capnp.SingleSegment(nil)) failed ")
		}
		spec, capnpErr := config.NewRootMetricsServiceSpec(s)
		if capnpErr != nil {
			CAPNP_ERR.Log(Logger().Panic()).Err(err).Msg("MetricsServiceSpec() - config.NewRootMetricsServiceSpec(s) failed ")
		}
		spec.SetHttpPort(DEFAULT_METRICS_HTTP_PORT)
		return spec
	}
	metricsServiceConfig, err := config.ReadRootMetricsServiceSpec(cfg)
	if err != nil {
		METRICS_SERVICE_CONFIG_ERROR.Log(Logger().Panic()).Err(err).Msg("")
	}
	if metricsServiceConfig.HttpPort() == 0 {
		metricsServiceConfig.SetHttpPort(DEFAULT_METRICS_HTTP_PORT)
	}
	return metricsServiceConfig
}

// metricsHttpReporter is used to expose metrics to Prometheus via HTTP
type metricsHttpReporter struct {
	*Service

	sync.Mutex
	httpServer *http.Server
	port       uint16
}

func (a *metricsHttpReporter) start() error {
	if !a.Alive() {
		return ServiceNotAliveError(METRICS_SERVICE_ID)
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

// config errors are treated as fatal errors
func (a *metricsHttpReporter) bootstrap() {
	registerMetricsHandler := func() {
		metricsServiceMutex.Lock()
		defer metricsServiceMutex.Unlock()
		metricsHandler := promhttp.HandlerFor(
			metricsRegistry,
			promhttp.HandlerOpts{
				ErrorLog:      a,
				ErrorHandling: promhttp.ContinueOnError,
			},
		)
		http.Handle("/metrics", metricsHandler)
	}

	registerMetrics := func() {
		toCounterMetricSpec := func(spec config.CounterMetricSpec) CounterMetricSpec {
			counterSpec, err := NewCounterMetricSpec(spec)
			if err != nil {
				METRICS_SERVICE_CONFIG_ERROR.Log(Logger().Panic()).Err(err).
					Uint64("svc", spec.ServiceId()).
					Uint64("metric", spec.MetricId()).
					Msg("Failed to read Help field")
			}
			return counterSpec
		}

		toCounterVectorMetricSpec := func(spec config.CounterVectorMetricSpec) *CounterVectorMetricSpec {
			counterSpec, err := NewCounterVectorMetricSpec(spec)
			if err != nil {
				METRICS_SERVICE_CONFIG_ERROR.Log(Logger().Panic()).Err(err).Msg("Failed to read CounterVectorMetricSpec.MetricSpec field")
			}
			return counterSpec
		}

		toGaugeMetricSpec := func(spec config.GaugeMetricSpec) GaugeMetricSpec {
			gaugeSpec, err := NewGaugeMetricSpec(spec)
			if err != nil {
				METRICS_SERVICE_CONFIG_ERROR.Log(Logger().Panic()).Err(err).
					Uint64("svc", spec.ServiceId()).
					Uint64("metric", spec.MetricId()).
					Msg("Failed to read Help field")
			}
			return gaugeSpec
		}

		toGaugeVectorMetricSpec := func(spec config.GaugeVectorMetricSpec) *GaugeVectorMetricSpec {
			gaugeSpec, err := NewGaugeVectorMetricSpec(spec)
			if err != nil {
				METRICS_SERVICE_CONFIG_ERROR.Log(Logger().Panic()).Err(err).Msg("Failed to read GaugeVectorMetricSpec.MetricSpec field")
			}
			return gaugeSpec
		}

		toHistogramMetricSpec := func(spec config.HistogramMetricSpec) HistogramMetricSpec {
			histogramSpec, err := NewHistogramMetricSpec(spec)
			if err != nil {
				METRICS_SERVICE_CONFIG_ERROR.Log(Logger().Panic()).Err(err).
					Uint64("svc", spec.ServiceId()).
					Uint64("metric", spec.MetricId()).
					Msg("")
			}
			return histogramSpec
		}

		toHistogramVectorMetricSpec := func(spec config.HistogramVectorMetricSpec) *HistogramVectorMetricSpec {
			histogramMetricSpec, err := NewHistogramVectorMetricSpec(spec)
			if err != nil {
				METRICS_SERVICE_CONFIG_ERROR.Log(Logger().Panic()).Err(err).Msg("")
			}
			return histogramMetricSpec
		}

		spec := MetricsServiceSpec()
		metricSpecs, err := spec.MetricSpecs()
		if err != nil {
			METRICS_SERVICE_CONFIG_ERROR.Log(Logger().Panic()).Err(err).Msg("")
		}

		counterSpecs, err := metricSpecs.CounterSpecs()
		if err != nil {
			METRICS_SERVICE_CONFIG_ERROR.Log(Logger().Panic()).Err(err).Msg("")
		}
		for i := 0; i < counterSpecs.Len(); i++ {
			metricSpec := toCounterMetricSpec(counterSpecs.At(i))
			metric := &CounterMetric{&metricSpec, prometheus.NewCounter(metricSpec.CounterOpts())}
			metric.register()
		}

		counterVectorSpecs, err := metricSpecs.CounterVectorSpecs()
		if err != nil {
			METRICS_SERVICE_CONFIG_ERROR.Log(Logger().Panic()).Err(err).Msg("")
		}
		for i := 0; i < counterVectorSpecs.Len(); i++ {
			metricSpec := toCounterVectorMetricSpec(counterVectorSpecs.At(i))
			metric := &CounterVectorMetric{metricSpec, prometheus.NewCounterVec(metricSpec.CounterOpts(), metricSpec.DynamicLabels)}
			metric.register()
		}

		gaugeSpecs, err := metricSpecs.GaugeSpecs()
		if err != nil {
			METRICS_SERVICE_CONFIG_ERROR.Log(Logger().Panic()).Err(err).Msg("")
		}
		for i := 0; i < gaugeSpecs.Len(); i++ {
			metricSpec := toGaugeMetricSpec(gaugeSpecs.At(i))
			metric := &GaugeMetric{&metricSpec, prometheus.NewGauge(metricSpec.GaugeOpts())}
			metric.register()
		}

		gaugeVectorSpecs, err := metricSpecs.GaugeVectorSpecs()
		if err != nil {
			METRICS_SERVICE_CONFIG_ERROR.Log(Logger().Panic()).Err(err).Msg("")
		}
		for i := 0; i < gaugeVectorSpecs.Len(); i++ {
			metricSpec := toGaugeVectorMetricSpec(gaugeVectorSpecs.At(i))
			metric := &GaugeVectorMetric{metricSpec, prometheus.NewGaugeVec(metricSpec.GaugeOpts(), metricSpec.DynamicLabels)}
			metric.register()
		}

		histogramSpecs, err := metricSpecs.HistogramSpecs()
		if err != nil {
			METRICS_SERVICE_CONFIG_ERROR.Log(Logger().Panic()).Err(err).Msg("")
		}
		for i := 0; i < histogramSpecs.Len(); i++ {
			metricSpec := toHistogramMetricSpec(histogramSpecs.At(i))
			metric := &HistogramMetric{&metricSpec, prometheus.NewHistogram(metricSpec.HistogramOpts())}
			metric.register()
		}

		histogramVectorSpecs, err := metricSpecs.HistogramVectorSpecs()
		if err != nil {
			METRICS_SERVICE_CONFIG_ERROR.Log(Logger().Panic()).Err(err).Msg("")
		}
		for i := 0; i < histogramVectorSpecs.Len(); i++ {
			metricSpec := toHistogramVectorMetricSpec(histogramVectorSpecs.At(i))
			metric := &HistogramVectorMetric{metricSpec, prometheus.NewHistogramVec(metricSpec.HistogramOpts(), metricSpec.DynamicLabels)}
			metric.register()
		}
	}

	registerMetricsHandlerOnce.Do(registerMetricsHandler)
	registerMetrics()
}

func (a *metricsHttpReporter) startHttpServer() {
	a.bootstrap()
	a.httpServer = &http.Server{
		Addr: fmt.Sprintf(":%d", a.port),
	}

	listenerWait := sync.WaitGroup{}
	listenerWait.Add(1)
	a.Go(func() error {
		METRICS_HTTP_SERVER_STARTING.Log(a.Logger().Info()).Str("addr", a.httpServer.Addr).Msg("Metrics HTTP server starting ...")
		listenerWait.Done()
		err := a.httpServer.ListenAndServe()
		logEvent := METRICS_HTTP_SERVER_STOPPED.Log(a.Logger().Info())
		if err != nil {
			logEvent.Str("reason", err.Error())
		}
		logEvent.Msg("Metrics http server has stopped")
		return nil
	})
	listenerWait.Wait()
	METRICS_HTTP_SERVER_STARTED.Log(a.Logger().Info()).Str("addr", a.httpServer.Addr).Msg("Metrics HTTP server started")
}

func (a *metricsHttpReporter) stop() {
	a.Lock()
	defer a.Unlock()
	background := context.Background()
	shutdownContext, cancel := context.WithTimeout(background, time.Second*5)
	defer cancel()
	// try graceful shutdown
	if err := a.httpServer.Shutdown(shutdownContext); err != nil {
		METRICS_HTTP_REPORTER_SHUTDOWN_ERROR.Log(a.Logger().Warn()).Err(ServiceShutdownError(METRICS_SERVICE_ID, err)).Msg("Error during HTTP server shutdown.")
	}
	// always ensure the listener is closed
	a.httpServer.Close()
	a.httpServer = nil
}

// Println implements promhttp.Logger interface.
// It is used to log any errors reported by the prometheus http handler
func (a *metricsHttpReporter) Println(v ...interface{}) {
	a.Logger().Error().Msg(fmt.Sprint(v))
}
