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
	"testing"

	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/oysterpack/oysterpack.go/oysterpack/metrics"
	metricsService "github.com/oysterpack/oysterpack.go/oysterpack/metrics/service"
	"github.com/oysterpack/oysterpack.go/oysterpack/service"
	"github.com/prometheus/client_golang/prometheus"
)

func TestNewClient(t *testing.T) {
	app := service.NewApplication(service.ApplicationSettings{})
	app.Start()
	defer app.Stop()

	client := app.MustRegisterService(metricsService.NewClient)
	client.Service().AwaitUntilRunning()

	service := client.(metricsService.Interface)

	var ping metrics.RunHealthCheck = func() error {
		return nil
	}

	opts := prometheus.GaugeOpts{
		Name: "ping",
		Help: "ping always succeeds",
	}
	pingCheck := metrics.NewHealthCheck(opts, 0, ping)
	pingCheck.MustRegister(service.Registry())
	pingCheck.Run()

	body, err := getMetrics()
	if err != nil {
		t.Error(err)
	} else {
		t.Logf("metrics:\n%s", string(body))
	}

}

func getMetrics() (string, error) {
	url := fmt.Sprintf("http://localhost:%d/metrics", metricsService.Config.HTTPPort())
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(body), nil
}
