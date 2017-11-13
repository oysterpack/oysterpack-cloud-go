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

type Error struct {
	ErrorID
	Err error
}

func (a *Error) Error() string {
	return fmt.Sprintf("%x : %v", a.ErrorID, a.Err)
}

var (
	ErrAppNotAlive = &Error{ErrorID(0xdf76e1927f240401), errors.New("App is not alive")}

	ErrServiceNotAlive      = &Error{ErrorID(0x9cb3a496d32894d2), errors.New("Service is not alive")}
	ErrServiceNotRegistered = &Error{ErrorID(0xf34b64bac786f536), errors.New("Service is not registered")}
	ErrServiceNotAvailable  = &Error{ErrorID(0x8aae12f3016b7f50), errors.New("Service is not available")}

	ErrServiceAlreadyRegistered = &Error{ErrorID(0xcfd879a478f9c733), errors.New("Service already registered")}
	ErrServiceNil               = &Error{ErrorID(0x9d95c5fac078b82c), errors.New("Service is nil")}
	ErrServiceIDZero            = &Error{ErrorID(0xd33c54b382368d97), errors.New("ServiceID(0) is not allowed")}

	ErrUnknownLogLevel = &Error{ErrorID(0x814a17666a94fe39), errors.New("Unknown log level")}

	ErrListenerFactoryNil     = &Error{ErrorID(0xbded147157abdee4), errors.New("ListenerFactory is nil")}
	ErrRPCServerFactoryNil    = &Error{ErrorID(0x9730bacf079fb410), errors.New("RPCServerFactory is nil")}
	ErrRPCServiceMaxConnsZero = &Error{ErrorID(0xea3df278b2992429), errors.New("The max numbe of connection must be > 0")}
)

func NewListenerFactoryError(err error) ListenerFactoryError {
	return ListenerFactoryError{
		&Error{ErrorID(0x828e2024b2a12526), err},
	}
}

type ListenerFactoryError struct {
	*Error
}

func NewRPCServerFactoryError(err error) RPCServerFactoryError {
	return RPCServerFactoryError{
		&Error{ErrorID(0x954d1590f06ffee5), err},
	}
}

type RPCServerFactoryError struct {
	*Error
}


