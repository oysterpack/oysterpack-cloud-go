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
	"errors"
	"fmt"
)

// Err maps an ErrorID to an error
type Err struct {
	ErrorID ErrorID
	Err     error
}

func (a *Err) Error() string {
	return fmt.Sprintf("ErrorID(0x%x) : %v", a.ErrorID, a.Err)
}

// UnrecoverableError is a marker interface for errors that cannot be recovered from automatically, i.e., manual intervention is required
type UnrecoverableError interface {
	UnrecoverableError()
}

// Errors
var (
	ErrAppNotAlive = &Err{ErrorID: ErrorID(0xdf76e1927f240401), Err: errors.New("App is not alive")}

	ErrServiceNotAlive      = &Err{ErrorID: ErrorID(0x9cb3a496d32894d2), Err: errors.New("Service is not alive")}
	ErrServiceNotRegistered = &Err{ErrorID: (0xf34b64bac786f536), Err: errors.New("Service is not registered")}
	ErrServiceNotAvailable  = &Err{ErrorID: ErrorID(0x8aae12f3016b7f50), Err: errors.New("Service is not available")}

	ErrServiceAlreadyRegistered = &Err{ErrorID: ErrorID(0xcfd879a478f9c733), Err: errors.New("Service already registered")}
	ErrServiceNil               = &Err{ErrorID: ErrorID(0x9d95c5fac078b82c), Err: errors.New("Service is nil")}

	ErrDomainIDZero      = &Err{ErrorID: ErrorID(0xb808d46722559577), Err: errors.New("DomainID(0) is not allowed")}
	ErrAppIDZero         = &Err{ErrorID: ErrorID(0xd5f068b2636835bb), Err: errors.New("AppID(0) is not allowed")}
	ErrServiceIDZero     = &Err{ErrorID: ErrorID(0xd33c54b382368d97), Err: errors.New("ServiceID(0) is not allowed")}
	ErrHealthCheckIDZero = &Err{ErrorID: ErrorID(0x9e04840a7fbac5ae), Err: errors.New("HealthCheckID(0) is not allowed")}

	ErrUnknownLogLevel = &Err{ErrorID(0x814a17666a94fe39), errors.New("Unknown log level")}

	ErrHealthCheckAlreadyRegistered = &Err{ErrorID: ErrorID(0xdbfd6d9ab0049876), Err: errors.New("HealthCheck already registered")}
	ErrHealthCheckNil               = &Err{ErrorID: ErrorID(0xf3a9b5c8afb8a698), Err: errors.New("HealthCheck is nil")}
	ErrHealthCheckNotRegistered     = &Err{ErrorID: ErrorID(0xefb3ffddac690f37), Err: errors.New("HealthCheck is not registered")}
	ErrHealthCheckNotAlive          = &Err{ErrorID: ErrorID(0xe1972916f1c18dae), Err: errors.New("HealthCheck is not alive")}
)

func NewServiceInitError(serviceId ServiceID, err error) ServiceInitError {
	return ServiceInitError{ServiceID: serviceId, Err: &Err{ErrorID: ErrorID(0xec1bf26105c1a895), Err: err}}
}

type ServiceInitError struct {
	ServiceID
	*Err
}

func (a ServiceInitError) Error() string {
	return fmt.Sprintf("ErrorID(0x%x)  : ServiceID(0x%x) : %v", a.ErrorID, a.ServiceID, a.Err)
}

// NewConfigError wraps an error as a ConfigError
func NewConfigError(err error) ConfigError {
	return ConfigError{
		&Err{ErrorID: ErrorID(0xe75f1a73534f382d), Err: err},
	}
}

// ConfigError indicates there was an error while trying to load a config
type ConfigError struct {
	*Err
}

func (a ConfigError) UnrecoverableError() {}

// NewMetricsServiceError wraps the error as a MetricsServiceError
func NewMetricsServiceError(err error) MetricsServiceError {
	return MetricsServiceError{&Err{ErrorID: ErrorID(0xc24ac892db47da9f), Err: err}}
}

// MetricsServiceError indicates an error occurred with in the MetricsHttpReporter
type MetricsServiceError struct {
	*Err
}

func NewHealthCheckTimeoutError(id HealthCheckID) HealthCheckTimeoutError {
	return HealthCheckTimeoutError{ErrorID(0x8257a572526e13f4), id}
}

type HealthCheckTimeoutError struct {
	ErrorID
	HealthCheckID
}

func (a HealthCheckTimeoutError) Error() string {
	return fmt.Sprintf("%x : HealthCheck timed out : %x", a.ErrorID, a.HealthCheckID)
}

func NewHealthCheckKillTimeoutError(id HealthCheckID) HealthCheckKillTimeoutError {
	return HealthCheckKillTimeoutError{ErrorID(0xf4ad6052397f6858), id}
}

type HealthCheckKillTimeoutError struct {
	ErrorID
	HealthCheckID
}

func (a HealthCheckKillTimeoutError) Error() string {
	return fmt.Sprintf("%x : HealthCheck timed out while dying : %x", a.ErrorID, a.HealthCheckID)
}
