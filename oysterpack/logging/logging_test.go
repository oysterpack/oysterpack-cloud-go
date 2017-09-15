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

package logging_test

import (
	"bytes"
	"encoding/json"
	"io"
	"testing"
	"time"

	"github.com/oysterpack/oysterpack.go/oysterpack/commons"
	"github.com/oysterpack/oysterpack.go/oysterpack/logging"
	"github.com/rs/zerolog"
)

type A struct{}

func TestNewPackageLogger(t *testing.T) {
	now := time.Now()
	t.Logf("now = %v", now.Format(zerolog.TimeFieldFormat))

	event := logging.Event{0, "RUNNING"}
	logger := logging.NewPackageLogger(A{})
	logger.Info().Dict(logging.EVENT, event.Dict()).Msg("")

	var buf bytes.Buffer
	logger = logger.Output(io.Writer(&buf))
	logger.Info().Dict(logging.EVENT, event.Dict()).Msg("")
	t.Log(buf.String())

	logEvent := &logging.LogEvent{}
	json.Unmarshal(buf.Bytes(), logEvent)
	t.Logf("logEvent : %v", logEvent)
	if logEvent.Time.Before(now) {
		t.Errorf("Time was not parsed correctly : %v : %v", buf.String(), logEvent.Time)
	}
	if logEvent.Level != logging.INFO {
		t.Errorf("Level was not parsed correctly : %v", logEvent.Level)
	}
	if logEvent.Package != commons.ObjectPackage(A{}) {
		t.Errorf("Package was not parsed correctly : %v", logEvent.Package)
	}
	if *logEvent.Event != event {
		t.Errorf("Event was not parsed correctly : %v", logEvent.Package)
	}
}

func TestNewPackageLogger_ForUnnamedType(t *testing.T) {
	func() {
		defer func() {
			if p := recover(); p == nil {
				t.Error("logging.NewPackageLogger(1) should have panicked because a named type is required")
			}
		}()
		logging.NewPackageLogger(1)
	}()
}

func TestNewTypeLogger(t *testing.T) {
	now := time.Now()
	t.Logf("now = %v", now.Format(zerolog.TimeFieldFormat))

	logger := logging.NewTypeLogger(A{})
	logger.Info().Msg("")

	var buf bytes.Buffer
	logger = logger.Output(io.Writer(&buf))
	logger.Info().Msg("")
	t.Log(buf.String())

	logEvent := &logging.LogEvent{}
	json.Unmarshal(buf.Bytes(), logEvent)
	t.Logf("unmarshalled logEvent JSON : %v", logEvent)
	t.Logf("unmarshalled logEvent : %v", *logEvent)
	if logEvent.Time.Before(now) {
		t.Errorf("Time was not parsed correctly : %v : %v", buf.String(), logEvent.Time)
	}
	if logEvent.Level != logging.INFO {
		t.Errorf("Level was not parsed correctly : %v", logEvent.Level)
	}
	if logEvent.Package != commons.ObjectPackage(A{}) {
		t.Errorf("Package was not parsed correctly : %v", logEvent.Package)
	}
	if logEvent.Type != "A" {
		t.Errorf("Type was not parsed correctly : %v", logEvent.Type)
	}
}

func TestNewTypeLogger_ForUnnamedType(t *testing.T) {
	func() {
		defer func() {
			if p := recover(); p == nil {
				t.Error("logging.NewTypeLogger(1) should have panicked because a named type is required")
			}
		}()
		logging.NewTypeLogger(1)
	}()
}
