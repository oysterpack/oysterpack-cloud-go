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

	"sort"

	"github.com/oysterpack/oysterpack.go/pkg/app/uid"
	"github.com/rs/zerolog"
)

const (
	// BUG is for "unknown" errors, i.e., for errors that we did not anticipate
	ErrorID_BUG = ErrorID(0)
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
	// misconfiguration or missing configuration
	ErrorType_Config
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
	// should cause the app to terminate, i.e. trigger a panic
	ErrorSeverity_FATAL
)

func NewError(cause error, message string, errSpec ErrSpec, serviceID ServiceID, ctx interface{}, tags ...string) *Error {
	if len(tags) > 0 {
		tagSet := make(map[string]struct{})
		for _, t := range tags {
			tagSet[t] = struct{}{}
		}
		if len(tagSet) != len(tags) {
			i := 0
			for t := range tagSet {
				tags[i] = t
				i++
			}
			tags = tags[0:i]
		}
		sort.Strings(tags)
	}

	return &Error{
		Cause:         cause,
		Message:       message,
		ErrorID:       errSpec.ErrorID,
		ErrorType:     errSpec.ErrorType,
		ErrorSeverity: errSpec.ErrorSeverity,
		Stack:         string(debug.Stack()),
		Tags:          tags,
		Context:       ctx,
		UIDHash:       uid.NextUIDHash(),
		Time:          time.Now(),
		DomainID:      Domain(),
		AppID:         ID(),
		ReleaseID:     Release(),
		InstanceID:    Instance(),
		ServiceID:     serviceID,
		PID:           PID,
		Hostname:      HOSTNAME,
	}
}

func NewBug(cause error, message string, serviceID ServiceID, ctx interface{}, tags ...string) *Error {
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
//
// Even though the Error fields are directly exposed, Error should be treated as immutable - i.e., once created it should not be changed.
//
// TODO: create capnp message
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
	Stack string
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
	InstanceID
	ReleaseID
	// PID and Hostname are part of the error because errors need to be centrally reported
	PID      int
	Hostname string

	ServiceID
}

func (a *Error) Error() string {
	if a.Message != "" {
		return a.Message
	}
	return a.Cause.Error()
}

func (a *Error) ErrSpec() ErrSpec {
	return ErrSpec{a.ErrorID, a.ErrorType, a.ErrorSeverity}
}

// WithTag returns a new Error instance with the specified tag added.
// If this Error already has the tag, then the same Error is returned.
func (a *Error) WithTag(tag string) *Error {
	if a.HasTag(tag) {
		return a
	}
	return NewError(a.Cause, a.Message, a.ErrSpec(), a.ServiceID, a.Context, append(a.Tags, tag)...)
}

func (a *Error) Log(logger zerolog.Logger) {
	event := logger.Error()
	dict := zerolog.Dict().
		Uint64("id", a.ErrorID.UInt64()).
		Time("time", a.Time).
		Uint64("uid", a.UIDHash.UInt64()).
		Uint8("type", a.ErrorType.UInt8()).
		Uint8("sev", a.ErrorSeverity.UInt8()).
		Uint64("domain", a.DomainID.UInt64()).
		Uint64("app", a.AppID.UInt64()).
		Str("release", a.ReleaseID.Hex()).
		Uint64("service", a.ServiceID.UInt64()).
		Uint64("release", Release().UInt64()).
		Uint64("instance", Instance().UInt64())
	if len(a.Stack) > 0 {
		dict.Str("stack", string(a.Stack))
	}
	if len(a.Tags) > 0 {
		dict.Strs("tags", a.Tags)
	}
	event.Dict("err", dict).Msg(a.Error())
}

func (a *Error) HasTag(tag string) bool {
	for _, t := range a.Tags {
		if t == tag {
			return true
		}
	}
	return false
}

// IsError returns true if the error type is *Error and the ErrorID matches.
// Returns false if error is nil.
func IsError(err error, errorID ErrorID) bool {
	if err == nil {
		return false
	}
	e, ok := err.(*Error)
	if !ok {
		return false
	}
	return e.ErrorID == errorID
}

var (
	ErrSpec_AppNotAlive   = ErrSpec{ErrorID: ErrorID(0xdf76e1927f240401), ErrorType: ErrorType_KNOWN_EDGE_CASE, ErrorSeverity: ErrorSeverity_FATAL}
	ErrSpec_ConfigFailure = ErrSpec{ErrorID: ErrorID(0xe75f1a73534f382d), ErrorType: ErrorType_Config, ErrorSeverity: ErrorSeverity_FATAL}

	ErrSpec_ServiceInitFailed        = ErrSpec{ErrorID: ErrorID(0xec1bf26105c1a895), ErrorType: ErrorType_KNOWN_EDGE_CASE, ErrorSeverity: ErrorSeverity_FATAL}
	ErrSpec_ServiceShutdownFailed    = ErrSpec{ErrorID: ErrorID(0xc24ac892db47da9f), ErrorType: ErrorType_KNOWN_EDGE_CASE, ErrorSeverity: ErrorSeverity_MEDIUM}
	ErrSpec_ServiceNotAlive          = ErrSpec{ErrorID: ErrorID(0x9cb3a496d32894d2), ErrorType: ErrorType_KNOWN_EDGE_CASE, ErrorSeverity: ErrorSeverity_HIGH}
	ErrSpec_ServiceNotRegistered     = ErrSpec{ErrorID: ErrorID(0xf34b64bac786f536), ErrorType: ErrorType_KNOWN_EDGE_CASE, ErrorSeverity: ErrorSeverity_HIGH}
	ErrSpec_ServiceNotAvailable      = ErrSpec{ErrorID: ErrorID(0x8aae12f3016b7f50), ErrorType: ErrorType_KNOWN_EDGE_CASE, ErrorSeverity: ErrorSeverity_HIGH}
	ErrSpec_ServiceAlreadyRegistered = ErrSpec{ErrorID: ErrorID(0xcfd879a478f9c733), ErrorType: ErrorType_KNOWN_EDGE_CASE, ErrorSeverity: ErrorSeverity_HIGH}
	ErrSpec_IllegalArgument          = ErrSpec{ErrorID: ErrorID(0x9d95c5fac078b82c), ErrorType: ErrorType_KNOWN_EDGE_CASE, ErrorSeverity: ErrorSeverity_HIGH}

	ErrSpec_InvalidLogLevel = ErrSpec{ErrorID: ErrorID(0x814a17666a94fe39), ErrorType: ErrorType_Config, ErrorSeverity: ErrorSeverity_HIGH}

	ErrSpec_HealthCheckAlreadyRegistered = ErrSpec{ErrorID: ErrorID(0xdbfd6d9ab0049876), ErrorType: ErrorType_KNOWN_EDGE_CASE, ErrorSeverity: ErrorSeverity_MEDIUM}
	ErrSpec_HealthCheckNotRegistered     = ErrSpec{ErrorID: ErrorID(0xefb3ffddac690f37), ErrorType: ErrorType_KNOWN_EDGE_CASE, ErrorSeverity: ErrorSeverity_LOW}
	ErrSpec_HealthCheckNotAlive          = ErrSpec{ErrorID: ErrorID(0xe1972916f1c18dae), ErrorType: ErrorType_KNOWN_EDGE_CASE, ErrorSeverity: ErrorSeverity_HIGH}
	ErrSpec_HealthCheckKillTimeout       = ErrSpec{ErrorID: ErrorID(0xf4ad6052397f6858), ErrorType: ErrorType_KNOWN_EDGE_CASE, ErrorSeverity: ErrorSeverity_HIGH}
	ErrSpec_HealthCheckTimeout           = ErrSpec{ErrorID: ErrorID(0x8257a572526e13f4), ErrorType: ErrorType_KNOWN_EDGE_CASE, ErrorSeverity: ErrorSeverity_HIGH}
)

func AppNotAliveError() *Error {
	return NewError(
		errors.New("App is not alive"),
		"",
		ErrSpec_AppNotAlive,
		APP_SERVICE,
		nil,
	)
}

func ConfigError(serviceID ServiceID, err error, message string) *Error {
	return NewError(
		err,
		message,
		ErrSpec_ConfigFailure,
		serviceID,
		nil,
	)
}

func ServiceInitError(serviceID ServiceID, err error) *Error {
	return NewError(
		err,
		"Service failed to initialize",
		ErrSpec_ServiceInitFailed,
		serviceID,
		nil,
	)
}

func ServiceShutdownError(serviceID ServiceID, err error) *Error {
	return NewError(
		err,
		"An error occurred while shutting down the service",
		ErrSpec_ServiceInitFailed,
		serviceID,
		nil,
	)
}

func ServiceNotAliveError(serviceID ServiceID) *Error {
	return NewError(
		errors.New("Service is not alive"),
		"",
		ErrSpec_ServiceNotAlive,
		serviceID,
		nil,
	)
}

func ServiceNotRegisteredError(serviceID ServiceID) *Error {
	return NewError(
		errors.New("Service is not registered"),
		"",
		ErrSpec_ServiceNotRegistered,
		serviceID,
		nil,
	)
}

func ServiceNotAvailableError(serviceID ServiceID) *Error {
	return NewError(
		errors.New("Service is not available"),
		"",
		ErrSpec_ServiceNotAvailable,
		serviceID,
		nil,
	)
}

func ServiceAlreadyRegisteredError(serviceID ServiceID) *Error {
	return NewError(
		errors.New("Service is already registered"),
		"",
		ErrSpec_ServiceAlreadyRegistered,
		serviceID,
		nil,
	)
}

func IllegalArgumentError(message string) *Error {
	return NewBug(
		errors.New(message),
		"",
		APP_SERVICE,
		nil,
	)
}

func InvalidLogLevelError(message string) *Error {
	return NewError(
		fmt.Errorf("Invalid log level : %v", message),
		"",
		ErrSpec_InvalidLogLevel,
		APP_SERVICE,
		nil,
	)
}

func HealthCheckAlreadyRegisteredError(healthCheckID HealthCheckID) *Error {
	return NewError(
		fmt.Errorf("HealthCheck(0x%x) is already registered", healthCheckID),
		"",
		ErrSpec_HealthCheckAlreadyRegistered,
		HEALTHCHECK_SERVICE_ID,
		nil,
	)
}

func HealthCheckNotRegisteredError(healthCheckID HealthCheckID) *Error {
	return NewError(
		fmt.Errorf("HealthCheck(0x%x) is not registered", healthCheckID),
		"",
		ErrSpec_HealthCheckNotRegistered,
		HEALTHCHECK_SERVICE_ID,
		nil,
	)
}

func HealthCheckNotAliveError(healthCheckID HealthCheckID) *Error {
	return NewError(
		fmt.Errorf("HealthCheck(0x%x) is not alive", healthCheckID),
		"",
		ErrSpec_HealthCheckNotAlive,
		HEALTHCHECK_SERVICE_ID,
		nil,
	)
}

func HealthCheckKillTimeoutError(healthCheckID HealthCheckID) *Error {
	return NewError(
		fmt.Errorf("Timed out waiting for HealthCheck(0x%x) to die", healthCheckID),
		"",
		ErrSpec_HealthCheckKillTimeout,
		HEALTHCHECK_SERVICE_ID,
		nil,
	)
}

func HealthCheckTimeoutError(healthCheckID HealthCheckID) *Error {
	return NewError(
		fmt.Errorf("HealthCheck(0x%x) run timed out.", healthCheckID),
		"",
		ErrSpec_HealthCheckKillTimeout,
		HEALTHCHECK_SERVICE_ID,
		nil,
	)
}
