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
	"encoding/json"
	"fmt"
	"reflect"
	"time"

	"github.com/oysterpack/oysterpack.go/oysterpack/commons"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// logger fields
const (
	PACKAGE = "pkg"
	TYPE    = "type"
	FUNC    = "func"
	SERVICE = "svc"
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
			Str(TYPE, t.Name()).
			Logger()
	}
}

// NewPackageLogger returns a new logger with pkg={pkg}
// where {pkg} is o's package path
// o must be for a named type because the package path can only be obtained for named types
func NewPackageLogger(o interface{}) zerolog.Logger {
	if commons.ObjectPackage(o) == commons.NoPackage {
		panic("only objects for named types are supported")
	}
	return log.With().
		Str(PACKAGE, string(commons.ObjectPackage(o))).
		Logger()
}

func init() {
	// log with nanosecond precision time
	zerolog.TimeFieldFormat = time.RFC3339Nano
}

// Level is the logging level
type Level string

// log levels
const (
	DEBUG Level = "debug"
	INFO  Level = "info"
	WARN  Level = "warn"
	ERROR Level = "error"
	FATAL Level = "fatal"
)

// LogEvent contains the common fields for log events
type LogEvent struct {
	Time    time.Time           `json:"time"`
	Level   Level               `json:"level"`
	Package commons.PackagePath `json:"pkg"`
	Type    string              `json:"type,omitempty"`
	Func    string              `json:"func,omitempty"`
	Event   *Event              `json:"event,omitempty"`
	Service string              `json:"svc,omitempty"`
	State   string              `json:"state,omitempty"`
}

func (e *LogEvent) String() string {
	if bytes, err := json.Marshal(*e); err == nil {
		return string(bytes)
	}
	return fmt.Sprint(*e)
}

// Event represents some event
type Event struct {
	Id   int    `json:"id"`
	Code string `json:"code"`
}

// Dict converts the Event to a zerolog sub dictionary
func (e *Event) Dict() *zerolog.Event {
	return zerolog.Dict().
		Int("id", e.Id).
		Str("code", e.Code)
}

// InterfaceTypeDict returns a standard zerolog dictionary for an interface type
// It contains the Package and Type fields.
func InterfaceTypeDict(interfaceType commons.InterfaceType) *zerolog.Event {
	return zerolog.Dict().
		Str(PACKAGE, interfaceType.PkgPath()).
		Str(TYPE, interfaceType.Name())
}
