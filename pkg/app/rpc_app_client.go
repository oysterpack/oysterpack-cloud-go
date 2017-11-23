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
	"context"

	"github.com/oysterpack/oysterpack.go/pkg/app/capnprpc"
	"github.com/oysterpack/oysterpack.go/pkg/app/config"
)

func NewAppClient(serviceId ServiceID) (*capnprpc.App, error) {
	cfg, err := Config(serviceId)
	if err != nil {
		return nil, err
	}

	spec, err := config.ReadRootRPCClientSpec(cfg)
	if err != nil {
		return nil, err
	}

	rpcClientSpec, err := NewRPCClientSpec(spec)
	if err != nil {
		return nil, err
	}

	conn, err := rpcClientSpec.Conn()
	if err != nil {
		return nil, err
	}
	return &capnprpc.App{Client: conn.Bootstrap(context.Background())}, nil
}

func NewAppClientForAddr(serviceId ServiceID, networkAddr string) (*capnprpc.App, error) {
	cfg, err := Config(serviceId)
	if err != nil {
		return nil, err
	}

	spec, err := config.ReadRootRPCClientSpec(cfg)
	if err != nil {
		return nil, err
	}

	rpcClientSpec, err := NewRPCClientSpec(spec)
	if err != nil {
		return nil, err
	}

	conn, err := rpcClientSpec.ConnForAddr(networkAddr)
	if err != nil {
		return nil, err
	}
	return &capnprpc.App{Client: conn.Bootstrap(context.Background())}, nil
}
