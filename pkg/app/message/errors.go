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

package message

import (
	"errors"

	"fmt"

	"github.com/oysterpack/oysterpack.go/pkg/app"
)

var (
	ERR_NODATA      = &app.Err{ErrorID: app.ErrorID(0xdd05f420d972bcab), Err: errors.New("Message has no data")}
	ERR_MESSAGE_NIL = &app.Err{ErrorID: app.ErrorID(0xdb9c6bc7777e1eb5), Err: errors.New("Message is nil")}
)

func NewUnsupportedCompressionError(compression Message_Compression) *UnsupportedCompressionError {
	return &UnsupportedCompressionError{
		Err:         &app.Err{ErrorID: app.ErrorID(0x992d6f6806a5374f), Err: fmt.Errorf("Unsupported compression : %v", compression)},
		Compression: compression,
	}
}

type UnsupportedCompressionError struct {
	*app.Err
	Compression Message_Compression
}
