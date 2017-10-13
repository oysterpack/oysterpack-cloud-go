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

package nats

import (
	"time"

	"github.com/nats-io/go-nats"
)

// ConnInfo snapshot
// TODO: expose this information via an HTTP API
type ConnInfo struct {
	Id      string
	Created time.Time
	Tags    []string

	nats.Statistics
	LastReconnectTime time.Time

	Disconnects        int
	LastDisconnectTime time.Time

	Errors        int
	LastErrorTime time.Time

	nats.Status
}

func (a *ConnInfo) String() string {
	// ignoring error, because marshalling should never fail
	bytes, _ := json.Marshal(a)
	return string(bytes)
}
