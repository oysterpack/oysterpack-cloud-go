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
	"fmt"

	"github.com/rs/zerolog"
)

// DomainID is unique ID assigned to a domain
type DomainID uint64

func (a DomainID) Hex() string { return hex(uint64(a)) }

func (a DomainID) UInt64() uint64 { return uint64(a) }

// AppID unique application ID
type AppID uint64

func (a AppID) Hex() string { return hex(uint64(a)) }

func (a AppID) UInt64() uint64 { return uint64(a) }

// InstanceID is the unique id for an app instance. There may be multiple instances, i.e., processes,  of an app running.
// The instance id is used to differentiate the different app instances. For examples, logs and metrics will contain the instance id.
type InstanceID uint64

func (a InstanceID) UInt64() uint64 { return uint64(a) }

func (a InstanceID) Hex() string { return hex(uint64(a)) }

// ReleaseID is a unique ID assigned to an application release
type ReleaseID uint64

func (a ReleaseID) Hex() string { return hex(uint64(a)) }

func (a ReleaseID) UInt64() uint64 { return uint64(a) }

// ServiceID unique service ID
type ServiceID uint64

func (a ServiceID) Hex() string { return hex(uint64(a)) }

func (a ServiceID) UInt64() uint64 { return uint64(a) }

// ErrorID unique error id
type ErrorID uint64

func (a ErrorID) Hex() string { return hex(uint64(a)) }

func (a ErrorID) UInt64() uint64 { return uint64(a) }

// LogEventID
type LogEventID uint64

// Log sets the log event id for a new log event
func (a LogEventID) Log(event *zerolog.Event) *zerolog.Event {
	return event.Uint64("event", uint64(a))
}

func (a LogEventID) Hex() string { return hex(uint64(a)) }

// MetricID unique error id
type MetricID uint64

func (a MetricID) Hex() string { return hex(uint64(a)) }

func (a MetricID) UInt64() uint64 { return uint64(a) }

// Name returns the metric formatted name for Prometheus. The name must match the regex : [a-zA-Z_:][a-zA-Z0-9_:]*
func (a MetricID) PrometheusName(serviceId ServiceID) string {
	return fmt.Sprintf("op_%x_%x", serviceId, a)
}

// HealthCheckID unique healthcheck id
type HealthCheckID uint64

func (a HealthCheckID) Hex() string { return hex(uint64(a)) }

func hex(id uint64) string { return fmt.Sprintf("%x", id) }
