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
	return fmt.Sprintf("%x : %v", a.ErrorID, a.Err)
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

	ErrListenerFactoryNil           = &Err{ErrorID: ErrorID(0xbded147157abdee4), Err: errors.New("ListenerFactory is nil")}
	ErrRPCMainInterfaceNil          = &Err{ErrorID: ErrorID(0x9730bacf079fb410), Err: errors.New("RPCMainInterface is nil")}
	ErrRPCPortZero                  = &Err{ErrorID: ErrorID(0x9580be146625218d), Err: errors.New("RPC port cannot be 0")}
	ErrRPCServiceMaxConnsZero       = &Err{ErrorID: ErrorID(0xea3df278b2992429), Err: errors.New("The max number of connection must be > 0")}
	ErrRPCServiceUnknownMessageType = &Err{ErrorID: ErrorID(0xdff7dbf6f092058e), Err: errors.New("Unknown message type")}
	ErrRPCListenerNotStarted        = &Err{ErrorID: ErrorID(0xdb45f1d3176ecad3), Err: errors.New("RPC server listener is not started")}

	ErrPEMParsing = &Err{ErrorID: ErrorID(0xa7b59b95250c2789), Err: errors.New("Failed to parse PEM encoded cert(s)")}

	ErrHealthCheckAlreadyRegistered = &Err{ErrorID: ErrorID(0xdbfd6d9ab0049876), Err: errors.New("HealthCheck already registered")}
	ErrHealthCheckNil               = &Err{ErrorID: ErrorID(0xf3a9b5c8afb8a698), Err: errors.New("HealthCheck is nil")}
	ErrHealthCheckNotRegistered     = &Err{ErrorID: ErrorID(0xefb3ffddac690f37), Err: errors.New("HealthCheck is not registered")}
	ErrHealthCheckNotAlive          = &Err{ErrorID: ErrorID(0xe1972916f1c18dae), Err: errors.New("HealthCheck is not alive")}
)

// NewListenerFactoryError wraps the specified error as a ListenerFactoryError
func NewListenerFactoryError(err error) ListenerFactoryError {
	return ListenerFactoryError{
		&Err{ErrorID: ErrorID(0x828e2024b2a12526), Err: err},
	}
}

// ListenerFactoryError for ListenerFactory related errors
type ListenerFactoryError struct {
	*Err
}

// UnrecoverableError - if we can't start a listener, then that is something we cannot recover from automatically.
func (a ListenerFactoryError) UnrecoverableError() {}

// NewTLSConfigError wraps the specified error as a ListenerFactoryError
func NewTLSConfigError(err error) ListenerFactoryError {
	return ListenerFactoryError{
		&Err{ErrorID: ErrorID(0xb67cbd821c0ab946), Err: err},
	}
}

// TLSConfigError for TLS configuration related issues
type TLSConfigError struct {
	*Err
}

// UnrecoverableError required manual intervention to resolve the misconfiguration
func (a TLSConfigError) UnrecoverableError() {}

// NewNetListenError wraps the specified error as a NetListenError
func NewNetListenError(err error) NetListenError {
	return NetListenError{
		&Err{ErrorID: ErrorID(0xa1dcc954855732fc), Err: err},
	}
}

// NetListenError indicates there was an error when trying to start a network listener
type NetListenError struct {
	*Err
}

func (a NetListenError) UnrecoverableError() {}

// NewRPCServerFactoryError wraps an error as an RPCServerFactoryError
func NewRPCServerFactoryError(err error) RPCServerFactoryError {
	return RPCServerFactoryError{
		&Err{ErrorID: ErrorID(0x954d1590f06ffee5), Err: err},
	}
}

// RPCServerFactoryError indicates an error trying to create an RPC server
type RPCServerFactoryError struct {
	*Err
}

func (a RPCServerFactoryError) UnrecoverableError() {}

// NewRPCServerSpecError wraps the error as an RPCServerSpecError
func NewRPCServerSpecError(err error) RPCServerSpecError {
	return RPCServerSpecError{
		&Err{ErrorID: ErrorID(0x9394e42b4cf30b1b), Err: err},
	}
}

// RPCServerSpecError indicates the RPCServerSpec is invalid
type RPCServerSpecError struct {
	*Err
}

func (a RPCServerSpecError) UnrecoverableError() {}

// NewRPCClientSpecError wraps the error as an RPCClientSpecError
func NewRPCClientSpecError(err error) RPCClientSpecError {
	return RPCClientSpecError{
		&Err{ErrorID: ErrorID(0xebcb20d1b8ffd569), Err: err},
	}
}

// RPCClientSpecError indicates the RPCClientSpec is invalid
type RPCClientSpecError struct {
	*Err
}

func (a RPCClientSpecError) UnrecoverableError() {}

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
