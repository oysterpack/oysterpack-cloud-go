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
	service := app.NewService(app.ServiceID(1), app.Log())
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
	service := app.NewService(app.ServiceID(1), app.Log())
	if err := app.RegisterService(service); err != nil {
		t.Fatal(err)
	}

	service2, err := app.GetService(service.ID())
	if err != nil {
		t.Fatal(err)
	}
	if service2 == nil || service2.ID() != service.ID() {
		t.Errorf("Service was not returned : %v", service2)
	}
}
