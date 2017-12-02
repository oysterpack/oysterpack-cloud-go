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

import "github.com/oysterpack/oysterpack.go/pkg/app"

const (
	SERVER_LISTENER_STARTED = app.LogEventID(0x982731754b2ce950)

	SERVER_LISTENER_RESTART  = app.LogEventID(0xbf8353ac06579256)
	SERVER_NEW_CONN          = app.LogEventID(0xf4b3ea46a3a5f988)
	SERVER_CONN_CLOSED       = app.LogEventID(0xf5610a189674584b)
	SERVER_MAX_CONNS_REACHED = app.LogEventID(0xa982a966f9be952b)
)
