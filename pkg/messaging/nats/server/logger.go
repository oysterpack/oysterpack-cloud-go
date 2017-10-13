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

package server

import (
	natsserver "github.com/nats-io/gnatsd/server"
	"github.com/oysterpack/oysterpack.go/pkg/logging"
	"github.com/rs/zerolog"
)

type pkgobject struct{}

var logger = logging.NewPackageLogger(pkgobject{})

// NewNATSLogger creates a new NATS logger that delegates to the provided logger
func NewNATSLogger(logger *zerolog.Logger) natsserver.Logger {
	return natsLogger{logger}
}

type natsLogger struct {
	*zerolog.Logger
}

func (a natsLogger) Noticef(format string, v ...interface{}) {
	a.Info().Msgf(format, v...)
}

// Log a fatal error
func (a natsLogger) Fatalf(format string, v ...interface{}) {
	a.Info().Msgf(format, v...)
}

// Log an error
func (a natsLogger) Errorf(format string, v ...interface{}) {
	a.Info().Msgf(format, v...)
}

// Log a debug statement
func (a natsLogger) Debugf(format string, v ...interface{}) {
	a.Info().Msgf(format, v...)
}

// Log a trace statement
func (a natsLogger) Tracef(format string, v ...interface{}) {
	a.Info().Msgf(format, v...)
}
