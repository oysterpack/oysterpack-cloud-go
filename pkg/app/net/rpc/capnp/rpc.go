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
	"sync"

	"github.com/oysterpack/oysterpack.go/pkg/app"
)

var (
	rpcServices = make(map[app.ServiceID]*RPCService)

	RPC AppRPCServices
)

type AppRPCServices struct {
	sync.Mutex
}

// ServiceIDs returns ServiceID(s) for registered RPCService(s)
//
// errors:
//	- ErrAppNotAlive
func (a AppRPCServices) ServiceIDs() []app.ServiceID {
	RPC.Lock()
	defer RPC.Unlock()

	ids := make([]app.ServiceID, len(rpcServices))
	i := 0
	for id := range rpcServices {
		ids[i] = id
		i++
	}
	return ids
}

// Get returns the registered *RPCService
//
// errors :
// - ErrAppNotAlive
// - ErrServiceNotRegistered
func (a AppRPCServices) Service(id app.ServiceID) *RPCService {
	RPC.Lock()
	defer RPC.Unlock()
	return rpcServices[id]
}
