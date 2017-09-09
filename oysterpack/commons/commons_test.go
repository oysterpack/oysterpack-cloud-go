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

package commons_test

import (
	"github.com/oysterpack/oysterpack.go/oysterpack/commons"
	"testing"
)

func TestObjectPackage(t *testing.T) {
	expectedPackage := commons.PackagePath("github.com/oysterpack/oysterpack.go/oysterpack/commons_test")

	type IFoo interface{}
	type Foo struct{}
	foo := Foo{}
	t.Logf("foo's package: %v", commons.ObjectPackage(foo))
	if commons.ObjectPackage(foo) != expectedPackage {
		t.Errorf("Foo's package should be %q, but was %q", expectedPackage, commons.ObjectPackage(foo))
	}

	var ifoo IFoo = Foo{}
	t.Logf("ifoo's package: %v", commons.ObjectPackage(ifoo))
	if commons.ObjectPackage(ifoo) != expectedPackage {
		t.Errorf("&ifoo's package should be %q, but was %q", expectedPackage, commons.ObjectPackage(foo))
	}

	t.Logf("&ifoo's package: %v", commons.ObjectPackage(&ifoo))
	if commons.ObjectPackage(&ifoo) != expectedPackage {
		t.Errorf("&ifoo's package should be %q, but was %q", expectedPackage, commons.ObjectPackage(foo))
	}
}
