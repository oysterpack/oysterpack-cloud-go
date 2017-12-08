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
	"time"

	"runtime/debug"

	"github.com/oysterpack/oysterpack.go/pkg/app/uid"
	"github.com/rs/zerolog"
)

const (
	// BUG is for "unknown" errors, i.e., for errors that we did not anticipate
	ErrorID_BUG = ErrorID(0)

	ERR_MSG_FORMAT = "ErrorID(0x%x) UID(0x%x) %s"
)

var (
	ErrSpec_BUG = ErrSpec{ErrorID_BUG, ErrorType_BUG, ErrorSeverity_HIGH}
)

// ErrorType is used to classify error types
type ErrorType uint8

func (a ErrorType) UInt8() uint8 {
	return uint8(a)
}

// ErrorType enum values
const (
	// for errors that have not been anticipated or not yet customized to your system
	ErrorType_BUG = ErrorType(iota)
	// e.g. broken network connections, IO errors, service not available, etc
	ErrorType_KNOWN_EDGE_CASE
)

// ErrorSeverity is used to classify error severity levels
type ErrorSeverity uint8

func (a ErrorSeverity) UInt8() uint8 {
	return uint8(a)
}

// ErrorSeverity enum values
const (
	ErrorSeverity_LOW = ErrorSeverity(iota)
	ErrorSeverity_MEDIUM
	ErrorSeverity_HIGH
	// should cause the app to terminate
	ErrorSeverity_FATAL
)

func NewError(cause error, message string, errSpec ErrSpec, serviceID ServiceID, ctx interface{}, tags ...string) *Error {
	return &Error{
		Cause:         cause,
		Message:       message,
		ErrorID:       errSpec.ErrorID,
		ErrorType:     errSpec.ErrorType,
		ErrorSeverity: errSpec.ErrorSeverity,
		Stack:         debug.Stack(),
		Tags:          tags,
		Context:       ctx,
		UIDHash:       uid.NextUIDHash(),
		Time:          time.Now(),
		DomainID:      Domain(),
		AppID:         ID(),
		ServiceID:     serviceID,
		PID:           PID,
		Hostname:      HOSTNAME,
	}
}

func NewBug(cause error, message string, errSpec ErrSpec, serviceID ServiceID, ctx interface{}, tags ...string) *Error {
	return NewError(cause, message, ErrSpec_BUG, serviceID, ctx, tags...)
}

type ErrSpec struct {
	ErrorID
	ErrorType
	ErrorSeverity
}

type ServiceErrSpec struct {
	ServiceID
	ErrSpec
}

// Error is used to model all application errors purposely.
//
// Many developers make the mistake of error propagation as secondary to the flow of their system. Careful consideration
// is given to how data flows through the system, but errors are something that are tolerated and ferried up the stack
// without much thought, and ultimately dumped in front of the user. With just a little forethought, and minimal overhead,
// you can make your error handling an asset to your system.
//
// Errors indicate that your system has entered a state in which it cannot fulfill an operation that a user explicitly or
// implicitly  requested. because of this, it needs to relay a few pieces of critical information:
//
// 	- What happened
//	- When and where it happened
type Error struct {
	////////////////////
	// What happened //
	//////////////////

	Cause error
	// user friendly message
	Message string
	// Used to identify the exact error type
	ErrorID
	ErrorType
	ErrorSeverity
	Stack []byte
	// tags can be used to categorize errors, e.g., UI, DB, ES, SECURITY
	Tags []string
	// should support JSON marshalling - it will be stored as a JSON clob
	Context interface{}

	/////////////////////////////////
	// When and where it happened //
	///////////////////////////////

	// used for error tracking, i.e., used to lookup this specific error
	uid.UIDHash
	time.Time
	DomainID
	AppID
	ServiceID
	// PID and Hostname are part of the error because errors need to be centrally reported
	PID      int
	Hostname string
}

func (a *Error) Error() string {
	if a.Message != "" {
		return fmt.Sprintf(ERR_MSG_FORMAT, a.ErrorID, a.UIDHash, a.Message)
	}
	return fmt.Sprintf(ERR_MSG_FORMAT, a.ErrorID, a.UIDHash, a.Cause.Error())
}

func (a *Error) Log(logger zerolog.Logger) {
	var event *zerolog.Event
	if a.ErrorSeverity == ErrorSeverity_FATAL {
		event = logger.Error()
	} else {
		event = logger.Error()
	}
	dict := zerolog.Dict().
		Uint64("id", a.ErrorID.UInt64()).
		Time("time", a.Time).
		Uint64("uid", a.UIDHash.UInt64()).
		Uint8("type", a.ErrorType.UInt8()).
		Uint8("sev", a.ErrorSeverity.UInt8()).
		Uint64("domain", a.DomainID.UInt64()).
		Uint64("app", a.AppID.UInt64()).
		Uint64("service", a.ServiceID.UInt64())
	if len(a.Stack) > 0 {
		dict.Str("stack", string(a.Stack))
	}
	if len(a.Tags) > 0 {
		dict.Strs("tags", a.Tags)
	}
	event.Dict("err", dict).Msg(a.Error())
}

//////////////////// DEPRECATED ////////////////////////////////

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
