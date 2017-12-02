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

import "github.com/oysterpack/oysterpack.go/pkg/app"

const (
	RPC_SERVICE_LISTENER_STARTED = app.LogEventID(0xc1398919f7426edb)
	RPC_SERVICE_LISTENER_RESTART = app.LogEventID(0xa611d10b1dfc880d)
	RPC_SERVICE_NEW_CONN         = app.LogEventID(0xeeb8cd1422232a22)
	RPC_SERVICE_CONN_CLOSED      = app.LogEventID(0x8b5dd1b82559601b)
	RPC_SERVICE_CONN_REMOVED     = app.LogEventID(0x9156bdee6b48f2b3)
	RPC_CONN_CLOSE_ERR           = app.LogEventID(0xe4ce88e6d408a26c)
)
