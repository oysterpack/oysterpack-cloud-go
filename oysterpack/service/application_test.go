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

package service

import (
	"github.com/oysterpack/oysterpack.go/oysterpack/commons"
	"testing"
)

func TestApplication_MapWithTypeKey(t *testing.T) {
	type A interface{}
	type B interface{}

	var a A = "a"
	var b B = "b"
	interfaceA, _ := commons.ObjectInterface(&a)
	interfaceB, _ := commons.ObjectInterface(&b)

	m := make(map[commons.InterfaceType]interface{})
	m[interfaceA] = a
	m[interfaceB] = b
	t.Logf("len(m) = %d", len(m))
	if len(m) != 2 {
		t.Error("invalid map size")
	}

	if m[interfaceA] != "a" {
		t.Error("a was not found")
	}

	if m[interfaceB] != "b" {
		t.Error("b was not found")
	}
}

type Foo interface {
	F()
}

type Bar interface {
	B()
}

type FooImpl struct{}

func (a *FooImpl) F() {}

func (a *FooImpl) B() {}

func TestApplication_TypeSwitch(t *testing.T) {
	var foo Foo = &FooImpl{}
	switch foo := foo.(type) {
	case Foo:
		t.Logf("Foo : %T", foo)
		foo.F()
	case Bar:
		t.Logf("Bar : %T", foo)
		foo.B()
	default:
		t.Errorf("%T", foo)
	}

	bar := foo.(Bar)
	t.Logf("converted Foo to Bar : %T", bar)
}
