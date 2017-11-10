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

import (
	"os"
	"strings"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// NewPackageLogger returns a new logger with pkg={pkg}
// where {pkg} is o's package path
// o must be for a named type because the package path can only be obtained for named types
func NewPackageLogger(o interface{}) zerolog.Logger {
	if ObjectPackage(o) == NoPackage {
		panic("only objects for named types are supported")
	}
	return log.With().
		Str("pkg", string(ObjectPackage(o))).
		Logger().
		Output(os.Stderr).
		Level(LoggingLevel())
}

// LoggingLevel returns the application log level.
// If the command line is parsed, then the -loglevel flag will be inspected. Valid values for -loglevel are : [DEBUG,INFO,WARN,ERROR]
// If not specified on the command line, then the defauly value is INFO.
// The log level is used to configure the log level for loggers returned via NewTypeLogger() and NewPackageLogger().
// It is also used to initialize zerolog's global logger level.
func LoggingLevel() zerolog.Level {
	if logLevel == nil {
		return zerolog.WarnLevel
	}
	switch strings.ToUpper(*logLevel) {
	case "DEBUG":
		return zerolog.DebugLevel
	case "INFO":
		return zerolog.InfoLevel
	case "WARN":
		return zerolog.WarnLevel
	case "ERROR":
		return zerolog.ErrorLevel
	default:
		return zerolog.WarnLevel
	}
}

// Level is the logging level
type Level string

// log levels
const (
	DEBUG Level = "debug"
	INFO  Level = "info"
	WARN  Level = "warn"
	ERROR Level = "error"
)
