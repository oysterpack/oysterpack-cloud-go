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

import "github.com/oysterpack/oysterpack.go/pkg/app/capnprpc"

func NewAppServer() capnprpc.App_Server {
	return rpcAppServer{}
}

type rpcAppServer struct{}

func (a rpcAppServer) Id(call capnprpc.App_id) error {
	call.Results.SetAppId(uint64(ID()))
	return nil
}

func (a rpcAppServer) Instance(call capnprpc.App_instance) error {
	return call.Results.SetInstanceId(string(InstanceId()))
}

func (a rpcAppServer) StartedOn(call capnprpc.App_startedOn) error {
	call.Results.SetStartedOn(StartedOn().UnixNano())
	return nil
}
