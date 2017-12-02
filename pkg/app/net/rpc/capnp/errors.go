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

package capnp

import (
	"errors"

	"github.com/oysterpack/oysterpack.go/pkg/app"
)

var (
	ErrRPCMainInterfaceNil          = &app.Err{ErrorID: app.ErrorID(0x9730bacf079fb410), Err: errors.New("RPCMainInterface is nil")}
	ErrRPCPortZero                  = &app.Err{ErrorID: app.ErrorID(0x9580be146625218d), Err: errors.New("RPC port cannot be 0")}
	ErrRPCServiceMaxConnsZero       = &app.Err{ErrorID: app.ErrorID(0xea3df278b2992429), Err: errors.New("The max number of connection must be > 0")}
	ErrRPCServiceUnknownMessageType = &app.Err{ErrorID: app.ErrorID(0xdff7dbf6f092058e), Err: errors.New("Unknown message type")}
	ErrRPCListenerNotStarted        = &app.Err{ErrorID: app.ErrorID(0xdb45f1d3176ecad3), Err: errors.New("RPC server listener is not started")}
)

// NewRPCServerFactoryError wraps an error as an RPCServerFactoryError
func NewRPCServerFactoryError(err error) RPCServerFactoryError {
	return RPCServerFactoryError{
		&app.Err{ErrorID: app.ErrorID(0x954d1590f06ffee5), Err: err},
	}
}

// RPCServerFactoryError indicates an error trying to create an RPC server
type RPCServerFactoryError struct {
	*app.Err
}

func (a RPCServerFactoryError) UnrecoverableError() {}

// NewRPCServerSpecError wraps the error as an RPCServerSpecError
func NewRPCServerSpecError(err error) RPCServerSpecError {
	return RPCServerSpecError{
		&app.Err{ErrorID: app.ErrorID(0x9394e42b4cf30b1b), Err: err},
	}
}

// RPCServerSpecError indicates the ServerSpec is invalid
type RPCServerSpecError struct {
	*app.Err
}

func (a RPCServerSpecError) UnrecoverableError() {}

// NewRPCClientSpecError wraps the error as an RPCClientSpecError
func NewRPCClientSpecError(err error) RPCClientSpecError {
	return RPCClientSpecError{
		&app.Err{ErrorID: app.ErrorID(0xebcb20d1b8ffd569), Err: err},
	}
}

// RPCClientSpecError indicates the RPCClientSpec is invalid
type RPCClientSpecError struct {
	*app.Err
}

func (a RPCClientSpecError) UnrecoverableError() {}
