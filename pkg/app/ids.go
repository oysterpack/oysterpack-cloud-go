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
	"github.com/rs/zerolog"
)

// AppID unique application ID
type AppID uint64

// InstanceID is the unique id for an app instance. There may be multiple instances, i.e., processes,  of an app running.
// The instance id is used to differentiate the different app instances. For examples, logs and metrics will contain the instance id.
type InstanceID string

// ReleaseID is a unique ID assigned to an application release
type ReleaseID uint64

// ServiceID unique service ID
type ServiceID uint64

// ErrorID unique error id
type ErrorID uint64

// LogEventID
type LogEventID uint64

func (a LogEventID) Log(event *zerolog.Event) *zerolog.Event {
	return event.Uint64("event", uint64(a))
}

// MetricID unique error id
type MetricID uint64

// HealthCheckID unique healthcheck id
type HealthCheckID uint64
