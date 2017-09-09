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

package commons

import (
	"fmt"
	"reflect"
)

// PackagePath represents a go package path
type PackagePath string

// TypeName represents a go type name
type TypeName string

type InterfaceType reflect.Type

type StructType reflect.Type

// ObjectPackage returns the package the the specified object belongs to
func ObjectPackage(o interface{}) PackagePath {
	return TypePackage(reflect.TypeOf(o))
}

func TypePackage(t reflect.Type) PackagePath {
	switch {
	case t.Kind() == reflect.Ptr:
		return TypePackage(t.Elem())
	default:
		return PackagePath(t.PkgPath())
	}
}

// Interface will check that t is either an interface or an interface pointer.
// If it is an interface pointer, then the interface that is pointed to is returned.
// If it is not an interface, then an error is returned describing the actual type.
func Interface(t reflect.Type) (InterfaceType, error) {
	switch t.Kind() {
	case reflect.Interface:
		return t, nil
	case reflect.Ptr:
		return Interface(t.Elem())
	default:
		return nil, fmt.Errorf("not an interface (package: %v, name: %v, kind: %v)", t.PkgPath(), t.Name(), t.Kind())
	}
}

func Struct(t reflect.Type) (StructType, error) {
	switch t.Kind() {
	case reflect.Struct:
		return t, nil
	case reflect.Ptr:
		return Struct(t.Elem())
	default:
		return nil, fmt.Errorf("not a struct (package: %v, name: %v, kind: %v)", t.PkgPath(), t.Name(), t.Kind())
	}
}
