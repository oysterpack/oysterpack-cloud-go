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
	"io"

	"github.com/Masterminds/semver"
	"github.com/oysterpack/oysterpack.go/pkg/metrics"
	"github.com/rs/zerolog"
)

// Settings is used by NewService to create a new service instance
type Settings struct {
	// REQUIRED - represents the service interface from the client's perspective
	// If the service has no direct client API, e.g., a network based service, then use an empty interface{}
	ServiceInterface

	*semver.Version

	// OPTIONAL - functions that define the service lifecycle
	Init
	Run
	Destroy

	// REQUIRED - InterfaceDependencies returns the Service interfaces that this service depends with version constraints
	// It can be used to check if all service Dependencies satisfied by the application.
	InterfaceDependencies

	LogSettings

	HealthChecks []metrics.HealthCheck

	// service metrics
	Metrics *metrics.MetricOpts
}

// LogSettings groups the log settings for the service
type LogSettings struct {
	// OPTIONAL - used to specify an alternative writer for the service logger
	LogOutput io.Writer

	// OPTIONAL - if not specified then the global default log level is used
	LogLevel *zerolog.Level
}
