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

// to run test coverage, execut the following command
//
// 		go test -covermode=count -coverprofile cover.out -args -app-id 123 -release-id 456 -log-level WARN -service-log-level 1=INFO,2=DISABLED,A=INFO,3
//
//  the test command line args will exercise the command line flag parsing
package app_test

import (
	"testing"

	"time"

	"errors"

	"sync"

	"github.com/oysterpack/oysterpack.go/pkg/app"
	"gopkg.in/tomb.v2"
)

func TestLogger(t *testing.T) {
	wait := sync.WaitGroup{}
	wait.Add(1)

	g := tomb.Tomb{}
	g.Go(func() error {
		logger := app.Logger()
		wait.Wait()
		logger.Debug().Msg("logger.Debug() : DEBUG MSG")
		app.Logger().Debug().Msg("app.Logger().Debug() : DEBUG MSG")
		return nil
	})
	wait.Done()
	g.Wait()
}

func TestNewService_IDZero(t *testing.T) {
	defer func() {
		if p := recover(); p == nil {
			t.Error("a zero ServiceID should have triggered a panic")
		} else {
			t.Log(p)
		}
	}()

	app.NewService(app.ServiceID(0))
}

func TestApp(t *testing.T) {
	app.Reset()

	t.Logf("app id = %v", app.ID())
	if app.StartedOn().IsZero() {
		t.Error("app.StartedOn() should not be the zero value")
	}
	if app.Instance() == 0 {
		t.Error("app instance id should not be blank")
	}

	t.Logf("ReleaseID: %d", app.Release())

	// When a nil service is registered
	// Then app.RegisterService a panic should be triggered
	func() {
		defer func() {
			if p := recover(); p == nil {
				t.Error("panic should have been triggered")
			}
		}()
		app.Services.Register(nil)
	}()

	service := app.NewService(app.ServiceID(1))
	service.Kill(nil)
	// When a killed service is registered
	// Then app.RegisterService should return app.ErrServiceNil

	func() {
		defer func() {
			if p := recover(); p == nil {
				t.Error("registering a dead service is not allowed")
			}
		}()
		app.Services.Register(service)
	}()

	service = app.NewService(app.ServiceID(2))
	if !service.Alive() {
		t.Error("Service should be alive")
	}
	// Given an app that has been killed, i.e., is not alive
	app.Kill()
	if app.Alive() {
		t.Error("app was killed and should not be alive")
	}

	// When a service is registered
	// Then app.RegisterService should return app.ErrAppNotAlive
	func() {
		defer func() {
			if p := recover(); p == nil {
				t.Error("registering a service when the app is dead is not allowed")
			}
		}()
		app.Services.Register(service)
	}()

	// When the app is dead
	// Then services cannot be registered
	func() {
		defer func() {
			if p := recover(); p == nil {
				t.Error("should have panicked because app is dead")
			}
		}()
		app.Services.Register(app.NewService(3))
	}()

	// And LogLevel() can be invoked without hanging
	app.LogLevel()
	// And services can be unregistered
	app.Services.Unregister(service.ID())
	app.Services.ServiceIDs()
	app.Services.Service(service.ID())

	// When the app is reset - after being killed
	app.Reset()
	// Then it should become alive again
	if !app.Alive() {
		t.Error("app was reset and should be alive")
	}
}

func TestRegisterService(t *testing.T) {
	// Given a new app
	app.Reset()

	// Then there should be no registered services
	ids := app.Services.ServiceIDs()

	t.Logf("service ids : %v", ids)
	if len(ids) == 0 {
		t.Errorf("expected the metrics service to be automatically registered")
	}
	initialServiceCount := len(ids)
	metricsService := app.Services.Service(app.METRICS_SERVICE_ID)
	if metricsService == nil {
		t.Error("Metrics service is not registered")
	} else {
		if !metricsService.Alive() {
			t.Error("Metrics service should be alive")
		}
		t.Logf("Metrics Service : %s", metricsService.ID().Hex())
	}

	// When a service is registered
	service := app.NewService(app.ServiceID(1))
	app.Services.Register(service)

	// Then the service id should be returned
	ids = app.Services.ServiceIDs()
	if len(ids) != initialServiceCount+1 {
		t.Errorf("Service count is wrong : %d != %d", len(ids), initialServiceCount+1)
	}
	serviceIDFound := false
	for _, id := range ids {
		if id == service.ID() {
			serviceIDFound = true
			break
		}
	}
	if !serviceIDFound {
		t.Errorf("Returned service id does not match : %v", ids)
	}

	// And the service can be retrieved
	if service2 := app.Services.Service(service.ID()); service2 == nil {
		t.Error("service was not found")
	} else if service2.ID() != service.ID() {
		t.Errorf("wrong service was returned : %v", service2.ID())
	}

	// When the service is killed
	service.Kill(nil)
	service.Wait()
	time.Sleep(time.Millisecond * 20)
	// Then the app will automatically unregister the service
	if service2 := app.Services.Service(service.ID()); service2 != nil {
		t.Error("service should have been unregistered")
	}

	// And no ServiceIDs should be returned except for the app default provided services - Metrics service
	ids = app.Services.ServiceIDs()
	if len(ids) != initialServiceCount {
		t.Errorf("Service count does not match : %d != %d", len(ids), initialServiceCount)
	}

	// When an attempt to register a service that is already registered is made
	// Then service registration should fail with app.ErrServiceAlreadyRegistered
	app.Reset()
	service = app.NewService(app.ServiceID(1))
	app.Services.Register(service)
	func() {
		defer func() {
			if p := recover(); p == nil {
				t.Error("A panic was expected")
			}
		}()
		app.Services.Register(service)
	}()

}

func TestGetService(t *testing.T) {
	// Given a new app
	app.Reset()

	// When a service is registered
	service := app.NewService(app.ServiceID(1))
	app.Services.Register(service)

	//Then the service can be retrieved
	service2 := app.Services.Service(service.ID())
	if service2 == nil {
		t.Fatal("The service should have been returned")
	}
	if service2 == nil || service2.ID() != service.ID() {
		t.Errorf("Service was not returned : %v", service2)
	}

	// When the app is reset, it is effectively restarted
	app.Reset()
	// And no ServiceIDs should be returned except for the app default provided services - Metrics service
	ids := app.Services.ServiceIDs()
	if len(ids) != len(app.INFRASTRUCTURE_SERVICE_IDS) {
		t.Errorf("There should only be app infrastructure services registered : %v != %v", len(ids), len(app.INFRASTRUCTURE_SERVICE_IDS))
	} else {
		svc := app.Services.Service(app.METRICS_SERVICE_ID)
		if svc == nil {
			t.Error("MetricsService should always be available")
		}
		svc = app.Services.Service(app.HEALTHCHECK_SERVICE_ID)
		if svc == nil {
			t.Error("HealthCheck service should always be available")
		}
	}
}

func TestServiceDiedWithError(t *testing.T) {
	app.Reset()

	service := app.NewService(app.ServiceID(1))
	app.Services.Register(service)

	service.Kill(errors.New("TestServiceDiedWithError"))
}

const FooServiceID = app.ServiceID(0xc8947921de337098)

type FooService struct {
	*app.Service
	fooRequestChan  chan fooRequest
	barChan         chan struct{}
	barBufferedChan chan struct{}

	fooRequests chan fooRequest
}

func (a *FooService) startServer() {
	a.fooRequestChan = make(chan fooRequest)
	a.barChan = make(chan struct{})
	a.barBufferedChan = make(chan struct{}, 1024)

	a.fooRequests = make(chan fooRequest, 20)

	a.Go(func() error {
		for {
			select {
			case <-a.Dying():
				return nil
			case a.fooRequests <- fooRequest{make(chan struct{})}:
			}
		}
	})

	a.Go(func() error {
		for {
			select {
			case <-a.Dying():
				return nil
			case req := <-a.fooRequestChan:
				select {
				case <-a.Dying():
					return nil
				case req.response <- struct{}{}:
				}
			case <-a.barChan:
			case <-a.barBufferedChan:
			}
		}
	})
}

func (a *FooService) Foo() error {
	request := fooRequest{make(chan struct{})}
	select {
	case <-a.Dying():
		return app.ServiceNotAliveError(a.Service.ID())
	case a.fooRequestChan <- request:
		select {
		case <-a.Dying():
			return app.ServiceNotAliveError(a.Service.ID())
		case <-request.response:
			return nil
		}
	}
}

func (a *FooService) FooUsingCachedFooRequests() error {
	request := <-a.fooRequests
	select {
	case <-a.Dying():
		return app.ServiceNotAliveError(a.Service.ID())
	case a.fooRequestChan <- request:
		select {
		case <-a.Dying():
			return app.ServiceNotAliveError(a.Service.ID())
		case <-request.response:
			return nil
		}
	}
}

func (a *FooService) Bar() error {
	select {
	case <-a.Dying():
		return app.ServiceNotAliveError(a.Service.ID())
	case a.barChan <- struct{}{}:
		return nil
	}
}

func (a *FooService) BarUsingBufferedChan() error {
	select {
	case <-a.Dying():
		return app.ServiceNotAliveError(a.Service.ID())
	case a.barBufferedChan <- struct{}{}:
		return nil
	}
}

type fooRequest struct {
	response chan struct{}
}

// These benchmarks provide a baseline for best case scenarios.
//
// BenchmarkService/REQUEST-RESPONSE-8              				1000000              2369 ns/op              96 B/op          1 allocs/op
// BenchmarkService/REQUEST-RESPONSE_using_cached_FooRequests-8      500000              2824 ns/op              96 B/op          1 allocs/op
// BenchmarkService/REQUEST_-_Unbuffered_Chan-8     				1000000              1154 ns/op               0 B/op          0 allocs/op
// BenchmarkService/REQUEST_on_Buffered_Chan-8      				5000000               345 ns/op               0 B/op          0 allocs/op
//
//
func BenchmarkService(b *testing.B) {
	app.Reset()
	defer app.Reset()

	foo := FooService{Service: app.NewService(FooServiceID)}
	foo.startServer()

	app.Services.Register(foo.Service)
	b.ResetTimer()

	b.Run("REQUEST-RESPONSE", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			if err := foo.Foo(); err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("REQUEST-RESPONSE using cached FooRequests", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			if err := foo.FooUsingCachedFooRequests(); err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("REQUEST - Unbuffered Chan", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			if err := foo.Bar(); err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("REQUEST on Buffered Chan", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			if err := foo.BarUsingBufferedChan(); err != nil {
				b.Fatal(err)
			}
		}
	})
}
