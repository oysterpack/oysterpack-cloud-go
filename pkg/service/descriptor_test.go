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

func TestNewDescriptor(t *testing.T) {

	desc := service.NewDescriptor(" Oysterpack ", " TEST ", " Echo_service ", "1.2.3", EchoServiceInterface)

	t.Logf("%v", desc)

	if desc.Namespace() != "oysterpack" {
		t.Errorf("wrong Namespace")
	}
	if desc.System() != "test" {
		t.Errorf("wrong System")
	}
	if desc.Component() != "echo_service" {
		t.Errorf("wrong Component")
	}
	if desc.Version().Major() != 1 && desc.Version().Minor() != 2 && desc.Version().Patch() != 3 {
		t.Errorf("wrong version")
	}
}

func TestNewDescriptor_RequiredFields(t *testing.T) {

	shouldPanic := func(f func()) {
		t.Helper()
		defer func() {
			t.Helper()
			if p := recover(); p == nil {
				t.Error("should have panicked")
			} else {
				t.Logf("%v", p)
			}
		}()
		f()
	}

	shouldPanic(func() { service.NewDescriptor("  ", " TEST ", " Echo_service ", "1.2.3", EchoServiceInterface) })
	shouldPanic(func() { service.NewDescriptor(" Oysterpack ", "  ", " Echo_service ", "1.2.3", EchoServiceInterface) })
	shouldPanic(func() { service.NewDescriptor(" Oysterpack ", " TEST ", "  ", "", EchoServiceInterface) })
	shouldPanic(func() { service.NewDescriptor(" Oysterpack ", " TEST ", " Echo_service ", "1.2.3", nil) })
	shouldPanic(func() {
		service.NewDescriptor(" Oysterpack ", " TEST ", " Echo_service ", "dfsddf", reflect.TypeOf(&Zoo{}))
	})

}

type Zoo struct{}
