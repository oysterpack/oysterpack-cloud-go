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

package app

import "reflect"

// PackagePath represents a go package path
type PackagePath string

// NoPackage represents the "" package, which is for predeclared types and unnamed types
const NoPackage PackagePath = ""

// TypeName represents a go type name
type TypeName string

// InterfaceType is a type for an interface
type InterfaceType reflect.Type

// StructType is a type for a struct
type StructType reflect.Type

// ObjectPackage returns the package the the specified object belongs to
// It only supports named types or pointers to named types.
// If the type was predeclared (string, error) or unnamed (*T, struct{}, []int),
// then the package path will be the empty string.
func ObjectPackage(o interface{}) PackagePath {
	return TypePackage(reflect.TypeOf(o))
}

// TypePackage returns the package path that the type belongs to
// If the type is a pointer, then it returns the package for the pointer's element type
func TypePackage(t reflect.Type) PackagePath {
	switch {
	case t.Kind() == reflect.Ptr:
		return TypePackage(t.Elem())
	default:
		return PackagePath(t.PkgPath())
	}
}
