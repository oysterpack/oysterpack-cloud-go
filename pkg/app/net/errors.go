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
	ErrListenerProviderNil  = &app.Err{ErrorID: app.ErrorID(0xbded147157abdee4), Err: errors.New("net.Listener provider is nil")}
	ErrTLSConfigProviderNil = &app.Err{ErrorID: app.ErrorID(0xfde67413e94bf34c), Err: errors.New("tls.Config provider is nil")}
	ErrConnHandlerNil       = &app.Err{ErrorID: app.ErrorID(0xff3042c0c093461b), Err: errors.New("Conn handler is nil")}
	ErrListenerDown         = &app.Err{ErrorID: app.ErrorID(0xed6320923d7ac51e), Err: errors.New("Listener is down")}
)

// NewListenerProviderError wraps the specified error as a ListenerProviderError
func NewListenerProviderError(err error) ListenerProviderError {
	return ListenerProviderError{
		&app.Err{ErrorID: app.ErrorID(0x828e2024b2a12526), Err: err},
	}
}

// ListenerProviderError for ListenerFactory related errors
type ListenerProviderError struct {
	*app.Err
}

// UnrecoverableError - if we can't start a listener, then that is something we cannot recover from automatically.
func (a ListenerProviderError) UnrecoverableError() {}

// NewTLSConfigError wraps the specified error as a ListenerProviderError
func NewTLSConfigError(err error) ListenerProviderError {
	return ListenerProviderError{
		&app.Err{ErrorID: app.ErrorID(0xb67cbd821c0ab946), Err: err},
	}
}

// TLSConfigError for TLS configuration related issues
type TLSConfigError struct {
	*app.Err
}

// UnrecoverableError required manual intervention to resolve the misconfiguration
func (a TLSConfigError) UnrecoverableError() {}

// NewNetListenError wraps the specified error as a NetListenError
func NewNetListenError(err error) NetListenError {
	return NetListenError{
		&app.Err{ErrorID: app.ErrorID(0xa1dcc954855732fc), Err: err},
	}
}

// NetListenError indicates there was an error when trying to start a network listener
type NetListenError struct {
	*app.Err
}

func (a NetListenError) UnrecoverableError() {}
