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
	ErrorID
	Err error
}

func (a *Err) Error() string {
	return fmt.Sprintf("%x : %v", a.ErrorID, a.Err)
}

var (
	ErrAppNotAlive = &Err{ErrorID(0xdf76e1927f240401), errors.New("App is not alive")}

	ErrServiceNotAlive      = &Err{ErrorID(0x9cb3a496d32894d2), errors.New("Service is not alive")}
	ErrServiceNotRegistered = &Err{ErrorID(0xf34b64bac786f536), errors.New("Service is not registered")}
	ErrServiceNotAvailable  = &Err{ErrorID(0x8aae12f3016b7f50), errors.New("Service is not available")}

	ErrServiceAlreadyRegistered = &Err{ErrorID(0xcfd879a478f9c733), errors.New("Service already registered")}
	ErrServiceNil               = &Err{ErrorID(0x9d95c5fac078b82c), errors.New("Service is nil")}
	ErrServiceIDZero            = &Err{ErrorID(0xd33c54b382368d97), errors.New("ServiceID(0) is not allowed")}

	ErrUnknownLogLevel = &Err{ErrorID(0x814a17666a94fe39), errors.New("Unknown log level")}

	ErrListenerFactoryNil           = &Err{ErrorID(0xbded147157abdee4), errors.New("ListenerFactory is nil")}
	ErrRPCMainInterfaceNil          = &Err{ErrorID(0x9730bacf079fb410), errors.New("RPCMainInterface is nil")}
	ErrRPCServiceMaxConnsZero       = &Err{ErrorID(0xea3df278b2992429), errors.New("The max numbe of connection must be > 0")}
	ErrRPCServiceUnknownMessageType = &Err{ErrorID(0xdff7dbf6f092058e), errors.New("Unknown message type")}
	ErrRPCListenerNotStarted        = &Err{ErrorID(0xdb45f1d3176ecad3), errors.New("RPC server listener is not started")}
)

func NewListenerFactoryError(err error) ListenerFactoryError {
	return ListenerFactoryError{
		&Err{ErrorID(0x828e2024b2a12526), err},
	}
}

type ListenerFactoryError struct {
	*Err
}

func NewNetListenError(err error) NetListenError {
	return NetListenError{
		&Err{ErrorID(0xa1dcc954855732fc), err},
	}
}

type NetListenError struct {
	*Err
}

func NewRPCServerFactoryError(err error) RPCServerFactoryError {
	return RPCServerFactoryError{
		&Err{ErrorID(0x954d1590f06ffee5), err},
	}
}

type RPCServerFactoryError struct {
	*Err
}
