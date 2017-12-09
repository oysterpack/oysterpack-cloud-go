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
	"fmt"

	"errors"

	"github.com/oysterpack/oysterpack.go/pkg/app"
)

var (
	ErrSpec_NoData                 = app.ErrSpec{ErrorID: app.ErrorID(0xdd05f420d972bcab), ErrorType: app.ErrorType_KNOWN_EDGE_CASE, ErrorSeverity: app.ErrorSeverity_HIGH}
	ErrSpec_UnsupportedCompression = app.ErrSpec{ErrorID: app.ErrorID(0x992d6f6806a5374f), ErrorType: app.ErrorType_KNOWN_EDGE_CASE, ErrorSeverity: app.ErrorSeverity_HIGH}
)

func NoDataError(serviceID app.ServiceID) *app.Error {
	return app.NewError(
		errors.New("Message has no data"),
		"",
		ErrSpec_NoData,
		serviceID,
		nil,
	)
}

func UnsupportedCompressionError(serviceID app.ServiceID, compression Message_Compression) *app.Error {
	return app.NewError(
		fmt.Errorf("Unsupported compression : %v", compression),
		"",
		ErrSpec_UnsupportedCompression,
		serviceID,
		nil,
	)
}
