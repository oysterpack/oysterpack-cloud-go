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

import "gopkg.in/tomb.v2"

// Reset is exposed only for testing purposes.
// Reset will :
// 	1. kill the app and then resurrect it with a new Tomb
// 	2. start the app server.
//	3. start the app RPC server
//	4. reset metrics
//	5. start the metrics HTTP reporter
//	6. reset healthchecks
func Reset() {
	app.Kill(nil)
	app.Wait()

	app = tomb.Tomb{} // resurrection
	runAppServer()
	runRPCAppServer()

	resetMetrics()
	startMetricsHttpReporter()

	initHealthCheckService()

	APP_RESET.Log(logger.Info()).Msg("reset")
}

func ResetWithConfigDir(configDir string) {
	Configs.SetConfigDir(configDir)
	Reset()
}

func resetMetrics() {
	metricsServiceMutex.Lock()
	defer metricsServiceMutex.Unlock()
	metricsRegistry = newMetricsRegistry(true)
	counters = make(map[ServiceID]map[MetricID]*CounterMetric)
	counterVectors = make(map[ServiceID]map[MetricID]*CounterVectorMetric)
	gauges = make(map[ServiceID]map[MetricID]*GaugeMetric)
	gaugeVectors = make(map[ServiceID]map[MetricID]*GaugeVectorMetric)
	histograms = make(map[ServiceID]map[MetricID]*HistogramMetric)
	histogramVectors = make(map[ServiceID]map[MetricID]*HistogramVectorMetric)
}
