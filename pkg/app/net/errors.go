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

package net

import (
	"errors"

	"github.com/oysterpack/oysterpack.go/pkg/app"
)

var (
	ErrSpec_ListenerDown          = app.ErrSpec{ErrorID: app.ErrorID(0xed6320923d7ac51e), ErrorType: app.ErrorType_KNOWN_EDGE_CASE, ErrorSeverity: app.ErrorSeverity_HIGH}
	ErrSpec_ListenerProviderError = app.ErrSpec{ErrorID: app.ErrorID(0x828e2024b2a12526), ErrorType: app.ErrorType_KNOWN_EDGE_CASE, ErrorSeverity: app.ErrorSeverity_FATAL}
	ErrSpec_TLSConfigError        = app.ErrSpec{ErrorID: app.ErrorID(0xb67cbd821c0ab946), ErrorType: app.ErrorType_Config, ErrorSeverity: app.ErrorSeverity_FATAL}
	ErrSpec_NetListenError        = app.ErrSpec{ErrorID: app.ErrorID(0xa1dcc954855732fc), ErrorType: app.ErrorType_KNOWN_EDGE_CASE, ErrorSeverity: app.ErrorSeverity_FATAL}
	ErrSpec_ServerFactoryError    = app.ErrSpec{ErrorID: app.ErrorID(0x954d1590f06ffee5), ErrorType: app.ErrorType_KNOWN_EDGE_CASE, ErrorSeverity: app.ErrorSeverity_FATAL}
	ErrSpec_ServerSpecError       = app.ErrSpec{ErrorID: app.ErrorID(0x9394e42b4cf30b1b), ErrorType: app.ErrorType_Config, ErrorSeverity: app.ErrorSeverity_FATAL}
	ErrSpec_ClientSpecError       = app.ErrSpec{ErrorID: app.ErrorID(0xebcb20d1b8ffd569), ErrorType: app.ErrorType_Config, ErrorSeverity: app.ErrorSeverity_FATAL}

	//ErrServerNameBlank               = &app.Err{ErrorID: app.ErrorID(0x82ba8744c43fe673), Err: errors.New("Server name is blank")}
	//ErrServerMaxConnsZero            = &app.Err{ErrorID: app.ErrorID(0x999e5626a881b99b), Err: errors.New("Server max conns must be > 0")}
	//ErrServerConnKeepAlivePeriodZero = &app.Err{ErrorID: app.ErrorID(0xb25783843b427f53), Err: errors.New("Server conn keep alive period must be > 0")}
	//ErrConnBuffersNotConfigurable    = &app.Err{ErrorID: app.ErrorID(0xf6cd72498eb50225), Err: errors.New("Server conn buffers are not configurable")}

	//ErrPEMParsing = &app.Err{ErrorID: app.ErrorID(0xa7b59b95250c2789), Err: errors.New("Failed to parse PEM encoded cert(s)")}
	//
	//ErrServerPortZero = &app.Err{ErrorID: app.ErrorID(0x9580be146625218d), Err: errors.New("Server port cannot be 0")}
)

// "Listener is down"
func ListenerDownError(serviceID app.ServiceID) *app.Error {
	return app.NewError(
		errors.New("Listener is down"),
		"",
		ErrSpec_ListenerDown,
		serviceID,
		nil,
	)
}

func ListenerProviderError(serviceID app.ServiceID, err error) *app.Error {
	return app.NewError(
		err,
		"ListenerProvider failed",
		ErrSpec_ListenerProviderError,
		serviceID,
		nil,
	)
}

func TLSConfigError(serviceID app.ServiceID, err error) *app.Error {
	return app.NewError(
		err,
		"TLS config is invalid",
		ErrSpec_TLSConfigError,
		serviceID,
		nil,
	)
}

// NewServerFactoryError wraps an error as an ServerFactoryError
func ServerFactoryError(serviceID app.ServiceID, err error) *app.Error {
	return app.NewError(
		err,
		"Listener is down",
		ErrSpec_ServerFactoryError,
		serviceID,
		nil,
	)
}

// NewServerSpecError wraps the error as an ServerSpecError
func ServerSpecError(serviceID app.ServiceID, err error) *app.Error {
	return app.NewError(
		err,
		"ServerSpec is invalid",
		ErrSpec_ServerSpecError,
		serviceID,
		nil,
	)
}

// NewClientSpecError wraps the error as an ClientSpecError
func ClientSpecError(serviceID app.ServiceID, err error) *app.Error {
	return app.NewError(
		errors.New("Listener is down"),
		"",
		ErrSpec_ClientSpecError,
		serviceID,
		nil,
	)
}
