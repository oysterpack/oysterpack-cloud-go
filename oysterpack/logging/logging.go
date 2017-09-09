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

package logging

import (
	"github.com/oysterpack/oysterpack.go/oysterpack/commons"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"reflect"
	"time"
)

// logger fields
const (
	PACKAGE = "pkg"
	TYPE    = "type"
	FUNC    = "func"
	SERVICE = "svc"
	NAME    = "name"
	EVENT   = "event"
	ID      = "id"
	CODE    = "code"
	STATE   = "state"
)

// NewTypeLogger returns a new logger with pkg={pkg}, type={type}
// where {pkg} is o's package path and {type} is o's type name
// o must be a struct - the pattern is to use an empty struct
func NewTypeLogger(o interface{}) zerolog.Logger {
	if t, err := commons.Struct(reflect.TypeOf(o)); err != nil {
		panic("NewTypeLogger can only be created for a struct")
	} else {
		return log.With().
			Str(PACKAGE, string(commons.TypePackage(t))).
			Str(TYPE, string(commons.TypePackage(t))).
			Logger()
	}
}

// NewPackageLogger returns a new logger with pkg={pkg}
// where {pkg} is o's package path
// o must be a struct - the pattern is to use an empty struct
func NewPackageLogger(o interface{}) zerolog.Logger {
	if t, err := commons.Struct(reflect.TypeOf(o)); err != nil {
		panic("NewPackageLogger can only be created for a struct")
	} else {
		return log.With().Str(PACKAGE, string(commons.TypePackage(t))).Logger()
	}
}

func init() {
	zerolog.TimeFieldFormat = time.RFC3339Nano
}
