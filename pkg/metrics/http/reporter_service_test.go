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

package http_test

import (
	"testing"

	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/oysterpack/oysterpack.go/pkg/metrics"
	metricsService "github.com/oysterpack/oysterpack.go/pkg/metrics/http"
	"github.com/oysterpack/oysterpack.go/pkg/service"
	"github.com/prometheus/client_golang/prometheus"
)

func TestService(t *testing.T) {
	// service is auto registered
	client := <-service.App().ServiceByTypeAsync(metricsService.ReporterInterface).Channel()
	client.Service().AwaitUntilRunning()

	var ping metrics.RunHealthCheck = func() error {
		return nil
	}

	opts := prometheus.GaugeOpts{
		Name: "ping",
		Help: "ping always succeeds",
	}
	pingCheck := metrics.NewHealthCheck(opts, 0, ping)
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
