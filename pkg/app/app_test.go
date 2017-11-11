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

	"github.com/oysterpack/oysterpack.go/pkg/app"
)

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
