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
	return fmt.Sprint(a.ErrorID, a.Err)
}

var (
	ErrAppNotAlive     = &Error{ErrorID(0xdf76e1927f240401), errors.New("App is not alive")}
	ErrServiceNotAlive = &Error{ErrorID(0x9cb3a496d32894d2), errors.New("Service is not alive")}

	ErrServiceAlreadyRegistered = &Error{ErrorID(0xcfd879a478f9c733), errors.New("Service already registered")}
	ErrServiceNil               = &Error{ErrorID(0x9d95c5fac078b82c), errors.New("Service is nil")}
)
