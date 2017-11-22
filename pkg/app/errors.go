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

type Err struct {
	ErrorID ErrorID
	Err     error
}

func (a *Err) Error() string {
	return fmt.Sprintf("%x : %v", a.ErrorID, a.Err)
}

// UnrecoverableError is a marker interface for errors that cannot be recovered from
type UnrecoverableError interface {
	UnrecoverableError()
}

var (
	ErrAppNotAlive = &Err{ErrorID: ErrorID(0xdf76e1927f240401), Err: errors.New("App is not alive")}

	ErrServiceNotAlive      = &Err{ErrorID: ErrorID(0x9cb3a496d32894d2), Err: errors.New("Service is not alive")}
	ErrServiceNotRegistered = &Err{ErrorID: (0xf34b64bac786f536), Err: errors.New("Service is not registered")}
	ErrServiceNotAvailable  = &Err{ErrorID: ErrorID(0x8aae12f3016b7f50), Err: errors.New("Service is not available")}

	ErrServiceAlreadyRegistered = &Err{ErrorID: ErrorID(0xcfd879a478f9c733), Err: errors.New("Service already registered")}
	ErrServiceNil               = &Err{ErrorID: ErrorID(0x9d95c5fac078b82c), Err: errors.New("Service is nil")}
	ErrServiceIDZero            = &Err{ErrorID: ErrorID(0xd33c54b382368d97), Err: errors.New("ServiceID(0) is not allowed")}

	ErrUnknownLogLevel = &Err{ErrorID(0x814a17666a94fe39), errors.New("Unknown log level")}

	ErrListenerFactoryNil           = &Err{ErrorID: ErrorID(0xbded147157abdee4), Err: errors.New("ListenerFactory is nil")}
	ErrRPCMainInterfaceNil          = &Err{ErrorID: ErrorID(0x9730bacf079fb410), Err: errors.New("RPCMainInterface is nil")}
	ErrRPCServiceMaxConnsZero       = &Err{ErrorID: ErrorID(0xea3df278b2992429), Err: errors.New("The max numbe of connection must be > 0")}
	ErrRPCServiceUnknownMessageType = &Err{ErrorID: ErrorID(0xdff7dbf6f092058e), Err: errors.New("Unknown message type")}
	ErrRPCListenerNotStarted        = &Err{ErrorID: ErrorID(0xdb45f1d3176ecad3), Err: errors.New("RPC server listener is not started")}
)

func NewListenerFactoryError(err error) ListenerFactoryError {
	return ListenerFactoryError{
		&Err{ErrorID: ErrorID(0x828e2024b2a12526), Err: err},
	}
}

type ListenerFactoryError struct {
	*Err
}

func (a ListenerFactoryError) UnrecoverableError() {}

func NewTLSConfigError(err error) ListenerFactoryError {
	return ListenerFactoryError{
		&Err{ErrorID: ErrorID(0xb67cbd821c0ab946), Err: err},
	}
}

type TLSConfigError struct {
	*Err
}

func (a TLSConfigError) UnrecoverableError() {}

func NewNetListenError(err error) NetListenError {
	return NetListenError{
		&Err{ErrorID: ErrorID(0xa1dcc954855732fc), Err: err},
	}
}

type NetListenError struct {
	*Err
}

func (a NetListenError) UnrecoverableError() {}

func NewRPCServerFactoryError(err error) RPCServerFactoryError {
	return RPCServerFactoryError{
		&Err{ErrorID: ErrorID(0x954d1590f06ffee5), Err: err},
	}
}

type RPCServerFactoryError struct {
	*Err
}

func (a RPCServerFactoryError) UnrecoverableError() {}

func NewRPCServicePKIError(err error) RPCServerFactoryError {
	return RPCServerFactoryError{
		&Err{ErrorID: ErrorID(0x9394e42b4cf30b1b), Err: err},
	}
}

type RPCServicePKIError struct {
	*Err
}

func (a RPCServicePKIError) UnrecoverableError() {}

func NewConfigError(err error) ConfigError {
	return ConfigError{
		&Err{ErrorID: ErrorID(0xe75f1a73534f382d), Err: err},
	}
}

type ConfigError struct {
	*Err
}

func (a ConfigError) UnrecoverableError() {}

func NewServiceConfigNotExistError(id ServiceID) ServiceConfigNotExistError {
	return ServiceConfigNotExistError{
		&Err{ErrorID: ErrorID(0x9394e42b4cf30b1b), Err: fmt.Errorf("Service config does not exist : %x", id)},
		id,
	}
}

type ServiceConfigNotExistError struct {
	*Err
	ServiceID
}
