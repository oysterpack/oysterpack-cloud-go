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

// NewAppClient creates a new App capnp RPC client.
// An RPCClientSpec config must exist for the specified ServiceID.
//
// NOTE: when the app is clustered, the RPC client will connect to any instance in the cluster
// It will connect using the following network address : {DomainID}_{AppID}
func NewAppClient(serviceId ServiceID) (*AppRPCClient, error) {
	cfg, err := Configs.Config(serviceId)
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
	return &AppRPCClient{&capnprpc.App{Client: conn.Bootstrap(context.Background())}}, nil
}

// NewAppClientForAddr works the same as NewAppClient, except that it enables connecting to a specific app instance by network address
func NewAppClientForAddr(serviceId ServiceID, networkAddr string) (*AppRPCClient, error) {
	cfg, err := Configs.Config(serviceId)
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
	return &AppRPCClient{&capnprpc.App{Client: conn.Bootstrap(context.Background())}}, nil
}

// capnprpc.App function params - for RPC functions that take no params
var (
	_App_id_Params        = func(_ capnprpc.App_id_Params) error { return nil }
	_App_releaseId_Params = func(_ capnprpc.App_releaseId_Params) error { return nil }
	_App_instance         = func(_ capnprpc.App_instance_Params) error { return nil }
	_App_startedOn        = func(_ capnprpc.App_startedOn_Params) error { return nil }
	_App_logLevel         = func(_ capnprpc.App_logLevel_Params) error { return nil }
	_App_serviceIds       = func(_ capnprpc.App_serviceIds_Params) error { return nil }
	_App_rpcServiceIds    = func(_ capnprpc.App_rpcServiceIds_Params) error { return nil }
	_App_kill             = func(_ capnprpc.App_kill_Params) error { return nil }
	_App_runtime          = func(_ capnprpc.App_runtime_Params) error { return nil }
	_App_configs          = func(_ capnprpc.App_configs_Params) error { return nil }
)

// AppRPCClient wraps the capnprpc.App in order to provide a more user friendly interface
type AppRPCClient struct {
	*capnprpc.App
}

func (a *AppRPCClient) Id(ctx context.Context) capnprpc.App_id_Results_Promise {
	return a.App.Id(ctx, _App_id_Params)
}

func (a *AppRPCClient) ReleaseId(ctx context.Context) capnprpc.App_releaseId_Results_Promise {
	return a.App.ReleaseId(ctx, _App_releaseId_Params)
}

func (a *AppRPCClient) Instance(ctx context.Context) capnprpc.App_instance_Results_Promise {
	return a.App.Instance(ctx, _App_instance)
}

func (a *AppRPCClient) StartedOn(ctx context.Context) capnprpc.App_startedOn_Results_Promise {
	return a.App.StartedOn(ctx, _App_startedOn)
}

func (a *AppRPCClient) LogLevel(ctx context.Context) capnprpc.App_logLevel_Results_Promise {
	return a.App.LogLevel(ctx, _App_logLevel)
}

func (a *AppRPCClient) ServiceIds(ctx context.Context) capnprpc.App_serviceIds_Results_Promise {
	return a.App.ServiceIds(ctx, _App_serviceIds)
}

func (a *AppRPCClient) Service(ctx context.Context, id ServiceID) capnprpc.App_service_Results_Promise {
	return a.App.Service(ctx, func(params capnprpc.App_service_Params) error {
		params.SetId(uint64(id))
		return nil
	})
}

func (a *AppRPCClient) RpcServiceIds(ctx context.Context) capnprpc.App_rpcServiceIds_Results_Promise {
	return a.App.RpcServiceIds(ctx, _App_rpcServiceIds)
}

func (a *AppRPCClient) RpcService(ctx context.Context, id ServiceID) capnprpc.App_rpcService_Results_Promise {
	return a.App.RpcService(ctx, func(params capnprpc.App_rpcService_Params) error {
		params.SetId(uint64(id))
		return nil
	})
}

func (a *AppRPCClient) Kill(ctx context.Context) error {
	_, err := a.App.Kill(ctx, _App_kill).Struct()
	return err
}

func (a *AppRPCClient) Runtime(ctx context.Context) capnprpc.App_runtime_Results_Promise {
	return a.App.Runtime(ctx, _App_runtime)
}

func (a *AppRPCClient) Configs(ctx context.Context) capnprpc.App_configs_Results_Promise {
	return a.App.Configs(ctx, _App_configs)
}

// Close releases any resources associated with this client.
// No further calls to the client should be made after calling Close.
func (a *AppRPCClient) Close() {
	a.App.Client.Close()
}
