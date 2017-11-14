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

package messages

import (
	"errors"

	"github.com/oysterpack/oysterpack.go/pkg/app"
)

var (
	ErrMessageIDBlank              = &app.Err{app.ErrorID(0x932a01de3b2b5d79), errors.New("MessageID cannot be blank")}
	ErrEnvelopeCreatedOnZero       = &app.Err{app.ErrorID(0xa09b77a0c83578b8), errors.New("CreatedOn cannot be zero.")}
	ErrAddressBlank                = &app.Err{app.ErrorID(0xa10bd3bd0162cb89), errors.New("Address cannot be blank")}
	ErrEnvelopeMessageRequired     = &app.Err{app.ErrorID(0xbd2b8b2bcccc8415), errors.New("Message is required")}
	ErrMessageTypeZero             = &app.Err{app.ErrorID(0xcdee88fa5b948509), errors.New("MessageType cannot be zero")}
	ErrMessageTypeDoesNotMatch     = &app.Err{app.ErrorID(0x84d32a119aa62240), errors.New("MessageType does not match")}
	ErrCorrelationIdBlank          = &app.Err{app.ErrorID(0xbd4a8fa096fd7e6d), errors.New("CorrelationID cannot be blank")}
	ErrEnvelopeMessagePrototypeNil = &app.Err{app.ErrorID(0xc959e8ecd08e9442), errors.New("A Message prototype instance in needed for unmarshalling")}
)
