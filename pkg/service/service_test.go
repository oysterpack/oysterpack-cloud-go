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
	stdreflect "reflect"
	"testing"
	"time"

	"bytes"
	"io"
	"strings"

	"github.com/oysterpack/oysterpack.go/pkg/commons/reflect"
	"github.com/oysterpack/oysterpack.go/pkg/metrics"
	"github.com/oysterpack/oysterpack.go/pkg/service"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/rs/zerolog"
)

type FooService interface{}
type Foo struct{}

var foo FooService = Foo{}

type BarService interface{}
type Bar struct{}

var bar BarService = Bar{}

func TestNewService_WithNilLifeCycleFunctions(t *testing.T) {
	server := service.NewService(service.Settings{ServiceInterface: stdreflect.TypeOf(&foo), Version: service.NewVersion("1.0.0")})
	t.Logf("new service : %v", server)
	if !server.State().New() {
		t.Errorf("Service state should be 'New', but instead was : %q", server.State())
	}

	if _, err := startService(server, t); err != nil {
		t.Error(err)
	}

	if _, err := startService(server, t); err == nil {
		t.Errorf("The service should already be running - thus a service.IllegalStateError should have been returned")
	} else {
		switch err.(type) {
		case *service.IllegalStateError:
			t.Logf("IllegalStateError : %v", err)
		default:
			t.Errorf("The error type should be *service.IllegalStateError, but was %T", err)
		}
	}

	if !server.State().Running() {
		t.Errorf("Service state should be 'Running', but instead was : %q", server.State())
	}

	if !stopService(server, t) {
		t.Error("The service should have stopped")
	}
	if !server.State().Terminated() {
		t.Errorf("Service state should be 'Terminated', but instead was : %q", server.State())
	}
	if !server.StopTriggered() {
		t.Errorf("StopTriggered should be true")
	}
	t.Log("stopping a stopped service should cause no issues")
	stopService(server, t)

	if _, err := startService(server, t); err == nil {
		t.Error("Starting a stopped service should fail.")
	} else {
		switch err.(type) {
		case *service.IllegalStateError:
			t.Logf("Restart error: %v", err)
		default:
			t.Errorf("Expected error type is *service.IllegalStateError, but was %T", err)
		}
	}
}

func TestNewService_AwaitRunningForStoppedService(t *testing.T) {
	server := service.NewService(service.Settings{ServiceInterface: stdreflect.TypeOf(&foo), Version: service.NewVersion("1.0.0")})
	server.Stop()

	if err := server.AwaitUntilRunning(); err == nil {
		t.Errorf("An PastStateError should have been returned because the service is stopped")
	} else {
		switch err.(type) {
		case *service.PastStateError:
		default:
			t.Errorf("Unexpected error type : %T : %v", err, err)
		}
	}
}

func TestNewService_StoppingNewService(t *testing.T) {
	server := service.NewService(service.Settings{ServiceInterface: stdreflect.TypeOf(&foo), Version: service.NewVersion("1.0.0")})
	if !stopService(server, t) {
		t.Error("The service should have stopped")
	}
	if !server.State().Terminated() {
		t.Errorf("Service state should be 'Terminated', but instead was : %q", server.State())
	}
	t.Log("The service was never started. Stopping a service that is still in the 'New' state simply transitions it to 'Terminated'")
	if !server.StopTriggered() {
		t.Errorf("StopTriggered should be true")
	}
}

func TestNewService_AwaitBlocking(t *testing.T) {
	server := service.NewService(service.Settings{ServiceInterface: stdreflect.TypeOf(&foo), Version: service.NewVersion("1.0.0")})
	server.StartAsync()
	server.AwaitRunning(0)
	if !server.State().Running() {
		t.Errorf("Service state should be 'Running', but instead was : %q", server.State())
	}
	server.StopAsyc()
	server.AwaitStopped(0)
	if !server.State().Terminated() {
		t.Errorf("Service state should be 'Terminated', but instead was : %q", server.State())
	}
}

func TestNewService_WithNonNilLifeCycleFunctions(t *testing.T) {
	var init service.Init = func(ctx *service.Context) error {
		t.Log("init")
		return nil
	}
	var run service.Run = func(ctx *service.Context) error {
		t.Log("running")
		for {
			select {
			case <-ctx.StopTrigger():
				t.Log("stop triggered")
				if !ctx.StopTriggered() {
					t.Error("StopTriggered should be true")
				}
				return nil
			}
		}
	}
	var destroy service.Destroy = func(ctx *service.Context) error {
		t.Log("destroy")
		return nil
	}

	server := service.NewService(service.Settings{ServiceInterface: stdreflect.TypeOf(&foo), Version: service.NewVersion("1.0.0"), Init: init, Run: run, Destroy: destroy})
	if !server.State().New() {
		t.Errorf("Service state should be 'New', but instead was : %q", server.State())
	}

	if _, err := startService(server, t); err != nil {
		t.Error(err)
	}

	if !server.State().Running() {
		t.Errorf("Service state should be 'Running', but instead was : %q", server.State())
	}

	stopService(server, t)
	if !server.State().Terminated() {
		t.Errorf("Service state should be 'Terminated', but instead was : %q", server.State())
	}
}

func TestNewService_InitPanics(t *testing.T) {
	var init service.Init = func(ctx *service.Context) error {
		panic("Init is panicking")
	}
	var run service.Run = func(ctx *service.Context) error {
		t.Log("running")
		for {
			select {
			case <-ctx.StopTrigger():
				t.Log("stop triggered")
				return nil
			}
		}
	}
	var destroy service.Destroy = func(ctx *service.Context) error {
		t.Log("destroy")
		return nil
	}

	server := service.NewService(service.Settings{ServiceInterface: stdreflect.TypeOf(&foo), Version: service.NewVersion("1.0.0"), Init: init, Run: run, Destroy: destroy})
	if !server.State().New() {
		t.Errorf("Service state should be 'New', but instead was : %q", server.State())
	}

	if started, err := startService(server, t); !started && err != nil {
		switch err.(type) {
		case *service.ServiceError:
		default:
			t.Errorf("Expected a service.ServiceError to be returned, but was %T : %v", err, err)
		}
		expectedError := err.(*service.ServiceError)
		t.Logf("expected error : %v", expectedError)
	} else {
		if started {
			t.Errorf("Expected server to fail to start")
		}

		if err == nil {
			t.Errorf("Expected a service.ServiceError to be returned")
		}
	}

	if !server.State().Failed() {
		t.Errorf("Service state should be 'Failed', but instead was : %q", server.State())
	}

	// stopping a server that is already stopped should be ok
	if !stopService(server, t) {
		t.Errorf("Service should already be in a stopped state, but we timed out waiting for the server to terminate")
	}
	if !server.State().Failed() {
		t.Errorf("Service state should be 'Failed', but instead was : %q", server.State())
	}
}

func TestNewService_RunPanics(t *testing.T) {
	var init service.Init = func(ctx *service.Context) error {
		t.Log("init")
		return nil
	}
	var run service.Run = func(ctx *service.Context) error {
		panic("Run is panicking")
	}
	var destroy service.Destroy = func(ctx *service.Context) error {
		t.Log("destroy")
		return nil
	}

	server := service.NewService(service.Settings{ServiceInterface: stdreflect.TypeOf(&foo), Version: service.NewVersion("1.0.0"), Init: init, Run: run, Destroy: destroy})
	if !server.State().New() {
		t.Errorf("Service state should be 'New', but instead was : %q", server.State())
	}

	if started, err := startService(server, t); !started && err != nil {
		switch err.(type) {
		case *service.ServiceError:
		default:
			t.Errorf("Expected a service.ServiceError to be returned, but was %T : %v", err, err)
		}
		expectedError := err.(*service.ServiceError)
		t.Logf("expected error : %v", expectedError)
	} else {
		// there is a possible timing issue where the state was set to Running right before the Run func panics
		server.AwaitUntilStopped()
		switch server.FailureCause().(type) {
		case *service.ServiceError:
			t.Logf("As expected, service failed : %v", server.FailureCause())
		default:
			t.Errorf("Expected a service.ServiceError to be returned, but was %T : %v", err, err)
		}
	}

	if !server.State().Failed() {
		t.Errorf("Service state should be 'Failed', but instead was : %q", server.State())
	}

	// stopping a server that is already stopped should be ok
	if !stopService(server, t) {
		t.Errorf("Service should already be in a stopped state, but we timed out waiting for the server to terminate")
	}
	if !server.State().Failed() {
		t.Errorf("Service state should be 'Failed', but instead was : %q", server.State())
	}
}

func TestNewService_DestroyPanics(t *testing.T) {
	var init service.Init = func(ctx *service.Context) error {
		t.Log("init")
		return nil
	}
	var run service.Run = func(ctx *service.Context) error {
		t.Log("running")
		for {
			select {
			case <-ctx.StopTrigger():
				t.Log("stop triggered")
				return nil
			}
		}
	}
	var destroy service.Destroy = func(ctx *service.Context) error {
		panic("Destroy is panicking")
	}

	server := service.NewService(service.Settings{ServiceInterface: stdreflect.TypeOf(&foo), Version: service.NewVersion("1.0.0"), Init: init, Run: run, Destroy: destroy})
	if !server.State().New() {
		t.Errorf("Service state should be 'New', but instead was : %q", server.State())
	}

	if _, err := startService(server, t); err != nil {
		t.Error(err)
	}

	if !server.State().Running() {
		t.Errorf("Service state should be 'Running', but instead was : %q", server.State())
	}

	stopService(server, t)
	if !server.State().Failed() {
		t.Errorf("Service state should be 'Terminated', but instead was : %q", server.State())
	}
	t.Log(server.FailureCause())
}

func TestService_ServiceDependencies(t *testing.T) {
	serviceDependency, _ := reflect.ObjectInterface(&bar)
	service := service.NewService(service.Settings{
		ServiceInterface:      stdreflect.TypeOf(&foo),
		Version:               service.NewVersion("1.0.0"),
		InterfaceDependencies: service.InterfaceDependencies{serviceDependency: nil},
	})
	if len(service.Dependencies()) != 1 {
		t.Errorf("Bar should have been added as a service dependency")
	}
	var barInterface reflect.InterfaceType
	for dependency := range service.Dependencies() {
		if dependency == serviceDependency {
			barInterface = dependency
		}
	}
	if barInterface != serviceDependency {
		t.Errorf("barInterface should == Bar, but was : %v", barInterface)
	}
}

func TestNewService_WithHealthChecks(t *testing.T) {
	var ping metrics.RunHealthCheck = func() error {
		return nil
	}
	opts := prometheus.GaugeOpts{
		Name: "ping",
		Help: "ping always succeeds",
	}
	registry_backup := metrics.Registry
	metrics.Registry = prometheus.NewPedanticRegistry()
	defer func() {
		// restore registry
		metrics.Registry = registry_backup
	}()

	service.NewService(service.Settings{
		ServiceInterface: stdreflect.TypeOf(&foo), Version: service.NewVersion("1.0.0"),
		HealthChecks: []metrics.HealthCheck{metrics.NewHealthCheck(opts, 0, ping)},
	})

	gatheredMetrics, _ := metrics.Registry.Gather()
	if m := metrics.FindMetricFamilyByName(gatheredMetrics, opts.Name); m == nil {
		t.Errorf("Metric was not found : %v", opts.Name)
		t.Error("Metrics should have been registered when the service was created")
	}
}

func TestNewService_WithDupHealthChecks(t *testing.T) {
	var ping metrics.RunHealthCheck = func() error {
		return nil
	}
	opts := prometheus.GaugeOpts{
		Name: "ping",
		Help: "ping always succeeds",
	}
	registry_backup := metrics.Registry
	metrics.Registry = prometheus.NewPedanticRegistry()
	defer func() {
		// restore registry
		metrics.Registry = registry_backup
	}()

	func() {
		defer func() {
			t.Helper()
			if p := recover(); p == nil {
				t.Error("registering dup metric should have triggered a panic")
			} else {
				t.Log(p)
			}
		}()
		service.NewService(service.Settings{
			ServiceInterface: stdreflect.TypeOf(&foo), Version: service.NewVersion("1.0.0"),
			HealthChecks: []metrics.HealthCheck{
				metrics.NewHealthCheck(opts, 0, ping),
				metrics.NewHealthCheck(opts, 0, ping),
			},
		})
	}()
}

func TestNewService_WithNilServiceInterface(t *testing.T) {
	defer func() {
		if p := recover(); p == nil {
			t.Error("registering service wilth nil ServiceInterface should have triggered a panic")
		} else {
			t.Log(p)
		}
	}()
	service.NewService(service.Settings{Version: service.NewVersion("1.0.0")})
}

func TestNewService_WithInvalidServiceInterface(t *testing.T) {
	defer func() {
		if p := recover(); p == nil {
			t.Error("registering service wilth nil ServiceInterface should have triggered a panic")
		} else {
			t.Log(p)
		}
	}()

	t.Logf("stdreflect.TypeOf(struct{}{}).Kind : %v", stdreflect.TypeOf(struct{}{}).Kind())

	service.NewService(service.Settings{ServiceInterface: stdreflect.TypeOf(&struct{}{}), Version: service.NewVersion("1.0.0")})
}

func TestNewService_WithLogSettings(t *testing.T) {

	buf := &bytes.Buffer{}
	var logOutput io.Writer = buf

	logLevel := zerolog.WarnLevel

	server := service.NewService(service.Settings{
		ServiceInterface: stdreflect.TypeOf(&foo), Version: service.NewVersion("1.0.0"),
		LogSettings: service.LogSettings{
			LogLevel:  &logLevel,
			LogOutput: logOutput,
		},
	})

	server.Logger().Info().Msg("INFO")
	server.Logger().Warn().Msg("WARN")

	if !strings.Contains(buf.String(), "WARN") {
		t.Errorf("Log level should be set to warn : %v", buf)
	}

	if strings.Contains(buf.String(), "INFO") {
		t.Errorf("INFO level message should not be logged : %v", buf)
	}
}

// startService waits up to 3 seconds for the server to start - checking every second
// returns true is the server started
// returns false if we timed out waiting for the server to start
func startService(server service.Service, t *testing.T) (bool, error) {
	if err := server.StartAsync(); err != nil {
		return false, err
	}
	for i := 1; i <= 3; i++ {
		if err := server.AwaitRunning(time.Second); err != nil {
			return false, err
		}
		if server.State().Running() {
			return true, nil
		}
		t.Logf("Waiting for server to run for %d sec ...", i)
	}
	return false, nil
}

// stopService waits up to 3 seconds for the server to stop
// returns true is the server stopped
// returns false if we timed out waiting for the server to stop
func stopService(server service.Service, t *testing.T) bool {
	server.StopAsyc()
	for i := 1; i <= 3; i++ {
		server.AwaitStopped(time.Second)
		if server.State().Stopped() {
			return true
		}
		t.Logf("Waiting for server (%v) to run for %d sec ...", server, i)
	}
	return false
}

type ConfigService struct {
	svc service.Service
}
