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

package app_test

import (
	"testing"

	"time"

	"errors"

	"github.com/oysterpack/oysterpack.go/pkg/app"
	"github.com/rs/zerolog"
)

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
	if len(app.InstanceId()) == 0 {
		t.Error("app instance id should not be blank")
	}

	// When a nil service is registered
	// Then app.RegisterService should return app.ErrServiceNil
	if err := app.RegisterService(nil); err == nil {
		t.Error("registering a nill service is not allowed")
	} else if err != app.ErrServiceNil {
		t.Errorf("Wrong error ttype was returned : %T", err)
	}

	service := app.NewService(app.ServiceID(1))
	service.Kill(nil)
	// When a killed service is registered
	// Then app.RegisterService should return app.ErrServiceNil
	if err := app.RegisterService(service); err == nil {
		t.Error("registering a dead service is not allowed")
	} else if err != app.ErrServiceNotAlive {
		t.Errorf("Wrong error ttype was returned : %T", err)
	}

	service = app.NewService(app.ServiceID(1))
	// Given an app that has been killed, i.e., is not alive
	app.Kill()
	// When a service is registered
	// Then app.RegisterService should return app.ErrAppNotAlive
	if err := app.RegisterService(service); err == nil {
		t.Error("registering a dead service is not allowed")
	} else if err != app.ErrAppNotAlive {
		t.Errorf("Wrong error ttype was returned : %T", err)
	}

	// When the app is dead
	// Then all app functions should fail and return ErrAppNotAlive
	if _, err := app.RegisteredServiceIds(); err == nil {
		t.Error("should have failed because app is dead")
	} else if err != app.ErrAppNotAlive {
		t.Errorf("Wrong error ttype was returned : %T", err)
	}
	if err := app.UnregisterService(service.ID()); err == nil {
		t.Error("should have failed because app is dead")
	} else if err != app.ErrAppNotAlive {
		t.Errorf("Wrong error ttype was returned : %T", err)
	}
	if _, err := app.GetService(service.ID()); err == nil {
		t.Error("should have failed because app is dead")
	} else {
		t.Log(err.Error())
		if err != app.ErrAppNotAlive {
			t.Errorf("Wrong error ttype was returned : %T", err)
		}
	}

	// And LogLevel() can be invoked without hanging
	app.LogLevel()
	// And SetServiceLogLevel should return an ErrAppNotAlive
	if app.SetServiceLogLevel(service.ID(), zerolog.ErrorLevel) != app.ErrAppNotAlive {
		t.Errorf("ErrAppNotAlive should have been returned : %v", app.SetServiceLogLevel(service.ID(), zerolog.ErrorLevel))
	}

	app.Reset()

	app.SetLogLevel(zerolog.InfoLevel)
	if app.LogLevel() != zerolog.InfoLevel {
		t.Errorf("log level did not match : %v", app.LogLevel())
	}
	app.SetLogLevel(zerolog.DebugLevel)
	if app.LogLevel() != zerolog.DebugLevel {
		t.Errorf("log level did not match : %v", app.LogLevel())
	}
}

func TestRegisterService(t *testing.T) {
	// Given a new app
	app.Reset()

	// Then there should be no registered services
	ids, err := app.RegisteredServiceIds()
	if err != nil {
		t.Fatal(err)
	}
	if len(ids) != 0 {
		t.Errorf("There should not be any services registered.")
	}

	// When a service is registered
	service := app.NewService(app.ServiceID(1))
	if err = app.RegisterService(service); err != nil {
		t.Fatal(err)
	}

	// Then the service id should be returned
	ids, err = app.RegisteredServiceIds()
	if err != nil {
		t.Error(err)
	}
	if len(ids) != 1 {
		t.Errorf("There should only be 1 service registered at this point")
	}
	if ids[0] != service.ID() {
		t.Errorf("Returned service id does not match : %v", ids)
	}
	// And the service can be retrieved
	if service2, err := app.GetService(service.ID()); err != nil {
		t.Error(err)
	} else if service2.ID() != service.ID() {
		t.Errorf("wrong service was returned : %v", service2.ID())
	}

	// When the service is killed
	service.Kill(nil)
	service.Wait()
	// Then the app will automatically unregister the service
	if service2, err := app.GetService(service.ID()); err != nil {
		t.Error(err)
	} else if service2 != nil {
		time.Sleep(time.Millisecond * 20)
		if service2, err = app.GetService(service.ID()); err != nil {
			t.Error(err)
		} else if service2 != nil {
			t.Error("service should have been unregistered by now")
		}
	}

	// And no ServiceIDs should be returned
	ids, err = app.RegisteredServiceIds()
	if err != nil {
		t.Fatal(err)
	}
	if len(ids) != 0 {
		t.Errorf("There should not be any services registered.")
	}

	// When an attempt to register a service that is already registered is made
	// Then service registration should fail with app.ErrServiceAlreadyRegistered
	app.Reset()
	service = app.NewService(app.ServiceID(1))
	if err := app.RegisterService(service); err != nil {
		t.Error(err)
	}
	if err := app.RegisterService(service); err == nil {
		t.Error(err)
	} else if err != app.ErrServiceAlreadyRegistered {
		t.Errorf("the wrong error type was returned : %T", err)
	}

	// When the service log level is changed
	// Then it is updated
	app.SetServiceLogLevel(service.ID(), zerolog.DebugLevel)
	if service.LogLevel() != zerolog.DebugLevel {
		t.Error("log level did not match : %v", service.LogLevel())
	}
	service, _ = app.GetService(service.ID())
	if service.LogLevel() != zerolog.DebugLevel {
		t.Error("log level did not match : %v", service.LogLevel())
	}

	if err := app.SetServiceLogLevel(service.ID(), zerolog.ErrorLevel); err != nil {
		t.Error(err)
	} else {
		if service.LogLevel() != zerolog.ErrorLevel {
			t.Error("log level did not match : %v", service.LogLevel())
		}
		service, _ = app.GetService(service.ID())
		if service.LogLevel() != zerolog.ErrorLevel {
			t.Error("log level did not match : %v", service.LogLevel())
		}
	}

	if err := app.SetServiceLogLevel(app.ServiceID(99), zerolog.ErrorLevel); err == nil {
		t.Error("The service should not be registered")
	} else if err != app.ErrServiceNotRegistered {
		t.Errorf("The wrong error type was returned : %T", err)
	}

}

func TestGetService(t *testing.T) {
	// Given a new app
	app.Reset()

	// When a service is registered
	service := app.NewService(app.ServiceID(1))
	if err := app.RegisterService(service); err != nil {
		t.Fatal(err)
	}

	//Then the service can be retrieved
	service2, err := app.GetService(service.ID())
	if err != nil {
		t.Fatal(err)
	}
	if service2 == nil || service2.ID() != service.ID() {
		t.Errorf("Service was not returned : %v", service2)
	}

	// When the app is reset, it is effectively restarted
	app.Reset()
	// And no ServiceIDs should be returned
	ids, err := app.RegisteredServiceIds()
	if err != nil {
		t.Fatal(err)
	}
	if len(ids) != 0 {
		t.Errorf("There should not be any services registered.")
	}
}

func TestServiceDiedWithError(t *testing.T) {
	app.Reset()

	service := app.NewService(app.ServiceID(1))
	if err := app.RegisterService(service); err != nil {
		t.Fatal(err)
	}

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
		return app.ErrServiceNotAlive
	case a.fooRequestChan <- request:
		select {
		case <-a.Dying():
			return app.ErrServiceNotAlive
		case <-request.response:
			return nil
		}
	}
}

func (a *FooService) FooUsingCachedFooRequests() error {
	request := <-a.fooRequests
	select {
	case <-a.Dying():
		return app.ErrServiceNotAlive
	case a.fooRequestChan <- request:
		select {
		case <-a.Dying():
			return app.ErrServiceNotAlive
		case <-request.response:
			return nil
		}
	}
}

func (a *FooService) Bar() error {
	select {
	case <-a.Dying():
		return app.ErrServiceNotAlive
	case a.barChan <- struct{}{}:
		return nil
	}
}

func (a *FooService) BarUsingBufferedChan() error {
	select {
	case <-a.Dying():
		return app.ErrServiceNotAlive
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

	if err := app.RegisterService(foo.Service); err != nil {
		b.Fatal(err)
	}
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
