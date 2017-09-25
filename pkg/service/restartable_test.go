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
	"reflect"
	"testing"

	"github.com/oysterpack/oysterpack.go/pkg/service"
)

func TestRestartableService_RestartService(t *testing.T) {

	service := service.NewRestartableService(func() service.Service {
		return service.NewService(service.Settings{Interface: reflect.TypeOf(&foo), Version: service.NewVersion("1.0.0")})
	})

	if !service.Service().State().New() {
		t.Fatalf("service state should be New, but is %q", service.Service().State())
	}
	if service.RestartCount() != 0 {
		t.Fatalf("restart count should be 0, but is %d", service.RestartCount())
	}

	service.Service().StartAsync()
	service.Service().AwaitUntilRunning()
	if !service.Service().State().Running() {
		t.Fatalf("service state should be Running, but is %q", service.Service().State())
	}
	service1 := service.Service()
	service.Restart()
	service.Service().AwaitUntilRunning()
	t.Logf("service state after restarting : %q", service.Service().State())
	if !service.Service().State().Running() {
		t.Fatalf("service state should be Running, but is %q", service.Service().State())
	}
	if service.RestartCount() != 1 {
		t.Fatalf("restart count should be 1, but is %d", service.RestartCount())
	}
	if service1 == service.Service() {
		t.Fatal("A new service instance should have been created")
	}
	service.Service().Stop()
	if !service.Service().State().Stopped() {
		t.Fatalf("service state should be stopped, but is %q", service.Service().State())
	}
	service1 = service.Service()
	service.Restart()
	if service1 == service.Service() {
		t.Fatal("A new service instance should have been created")
	}
	service.Service().AwaitUntilRunning()
	if !service.Service().State().Running() {
		t.Fatalf("service state should be Running, but is %q", service.Service().State())
	}
	if service.RestartCount() != 2 {
		t.Fatalf("restart count should be 2, but is %d", service.RestartCount())
	}
}
