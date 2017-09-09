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
	"reflect"
	"testing"
)

var thisPackage = commons.PackagePath("github.com/oysterpack/oysterpack.go/oysterpack/commons_test")

func TestObjectPackage(t *testing.T) {
	type IFoo interface{}
	type Foo struct{}

	foo := Foo{}
	if pkg := commons.ObjectPackage(foo); pkg != thisPackage {
		t.Errorf("foo's package should be %q, but was %q", thisPackage, pkg)
	}
	if pkg := commons.ObjectPackage(&foo); pkg != thisPackage {
		t.Errorf("&foo's package should be %q, but was %q", thisPackage, pkg)
	}

	var ifoo IFoo = Foo{}
	if pkg := commons.ObjectPackage(ifoo); pkg != thisPackage {
		t.Errorf("ifoo's package should be %q, but was %q", thisPackage, pkg)
	}

	if pkg := commons.ObjectPackage(&ifoo); pkg != thisPackage {
		t.Errorf("&ifoo's package should be %q, but was %q", thisPackage, pkg)
	}

	type Count int
	var count Count
	if pkg := commons.ObjectPackage(count); pkg != thisPackage {
		t.Errorf("count's package should be %q, but was %q", thisPackage, pkg)
	}
	if pkg := commons.ObjectPackage(&count); pkg != thisPackage {
		t.Errorf("&count's package should be %q, but was %q", thisPackage, pkg)
	}

	if commons.ObjectPackage(struct{}{}) != commons.NoPackage {
		t.Errorf("unnamed types should have no package")
	}
	if commons.ObjectPackage(1) != commons.NoPackage {
		t.Errorf("literal types should have no package")
	}
}

func TestInterface(t *testing.T) {
	type IFoo interface{}
	type Foo struct{}

	var ifoo IFoo = Foo{}
	if interfaceType, err := commons.Interface(reflect.TypeOf(&ifoo)); err != nil {
		t.Error(err)
	} else {
		if interfaceType.Name() != "IFoo" {
			t.Errorf("%v != IFoo", interfaceType.Name())
		}
	}

	if _, err := commons.Interface(reflect.TypeOf(ifoo)); err == nil {
		t.Error("Expected an error because the type passed in is for a struct")
	}

}

func TestObjectInterface(t *testing.T) {
	type IFoo interface{}
	type Foo struct{}

	var ifoo IFoo = Foo{}
	if interfaceType, err := commons.ObjectInterface(&ifoo); err != nil {
		t.Error(err)
	} else {
		if interfaceType.Name() != "IFoo" {
			t.Errorf("%v != IFoo", interfaceType.Name())
		}
	}

	if _, err := commons.ObjectInterface(ifoo); err == nil {
		t.Error("Expected an error because the type passed in is for a struct")
	}
}

func TestStruct(t *testing.T) {
	type Foo struct{}
	var foo = Foo{}
	if structType, err := commons.Struct(reflect.TypeOf(foo)); err != nil {
		t.Error(err)
	} else {
		if structType.Name() != "Foo" {
			t.Errorf("%v != Foo", structType.Name())
		}
	}

	if structType, err := commons.Struct(reflect.TypeOf(&foo)); err != nil {
		t.Error(err)
	} else {
		if structType.Name() != "Foo" {
			t.Errorf("%v != Foo", structType.Name())
		}
	}

	type IFoo interface{}
	var iFoo IFoo = Foo{}
	if _, err := commons.Struct(reflect.TypeOf(&iFoo)); err == nil {
		t.Error("Expected an error because the type passed in references an interface")
	}

	if structType, err := commons.Struct(reflect.TypeOf(iFoo)); err != nil {
		t.Error(err)
	} else {
		if structType.Name() != "Foo" {
			t.Errorf("%v != Foo", structType.Name())
		}
	}
}
