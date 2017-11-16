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
	"github.com/oysterpack/oysterpack.go/pkg/app/capnprpc"
	"github.com/rs/zerolog"
	"zombiezen.com/go/capnproto2"
)

func NewAppServer() capnprpc.App_Server {
	return rpcAppServer{}
}

type rpcAppServer struct{}

func (a rpcAppServer) Id(call capnprpc.App_id) error {
	call.Results.SetAppId(uint64(ID()))
	return nil
}

func (a rpcAppServer) ReleaseId(call capnprpc.App_releaseId) error {
	call.Results.SetReleaseId(uint64(Release()))
	return nil
}

func (a rpcAppServer) Instance(call capnprpc.App_instance) error {
	return call.Results.SetInstanceId(string(Instance()))
}

func (a rpcAppServer) StartedOn(call capnprpc.App_startedOn) error {
	call.Results.SetStartedOn(StartedOn().UnixNano())
	return nil
}

func (a rpcAppServer) LogLevel(call capnprpc.App_logLevel) error {
	level, err := ZerologLevel2capnprpcLogLevel(LogLevel())
	if err != nil {
		return err
	}
	call.Results.SetLevel(level)
	return nil
}

func (a rpcAppServer) RegisteredServices(call capnprpc.App_registeredServices) error {
	ids, err := RegisteredServiceIds()
	if err != nil {
		return err
	}

	list, err := capnp.NewUInt64List(call.Results.Segment(), int32(len(ids)))
	if err != nil {
		return err
	}
	for i := 0; i < list.Len(); i++ {
		list.Set(i, uint64(ids[i]))
	}
	call.Results.SetServiceIds(list)
	return nil
}

func (a rpcAppServer) Kill(call capnprpc.App_kill) error {
	app.Kill(nil)
	return nil
}

func (a rpcAppServer) Service(call capnprpc.App_service) error {
	service, err := GetService(ServiceID(call.Params.Id()))
	if err != nil {
		return err
	}

	return call.Results.SetService(capnprpc.Service_ServerToClient(ServiceServer{service.ID()}))
}

func CapnprpcLogLevel2zerologLevel(logLevel capnprpc.LogLevel) (zerolog.Level, error) {
	switch logLevel {
	case capnprpc.LogLevel_debug:
		return zerolog.DebugLevel, nil
	case capnprpc.LogLevel_info:
		return zerolog.InfoLevel, nil
	case capnprpc.LogLevel_warn:
		return zerolog.WarnLevel, nil
	case capnprpc.LogLevel_error:
		return zerolog.ErrorLevel, nil
	default:
		return zerolog.ErrorLevel, ErrUnknownLogLevel
	}
}

func ZerologLevel2capnprpcLogLevel(logLevel zerolog.Level) (capnprpc.LogLevel, error) {
	switch logLevel {
	case zerolog.DebugLevel:
		return capnprpc.LogLevel_debug, nil
	case zerolog.InfoLevel:
		return capnprpc.LogLevel_info, nil
	case zerolog.WarnLevel:
		return capnprpc.LogLevel_warn, nil
	case zerolog.ErrorLevel:
		return capnprpc.LogLevel_error, nil
	default:
		return capnprpc.LogLevel_error, ErrUnknownLogLevel
	}
}

type ServiceServer struct {
	ServiceID
}

func (a ServiceServer) Id(call capnprpc.Service_id) error {
	call.Results.SetServiceId(uint64(a.ServiceID))
	return nil
}

func (a ServiceServer) LogLevel(call capnprpc.Service_logLevel) error {
	service, err := GetService(a.ServiceID)
	if err != nil {
		return err
	}
	logLevel, err := ZerologLevel2capnprpcLogLevel(service.LogLevel())
	if err != nil {
		return err
	}
	call.Results.SetLevel(logLevel)
	return nil
}

func (a ServiceServer) Alive(call capnprpc.Service_alive) error {
	service, err := GetService(a.ServiceID)
	if err != nil {
		if err == ErrServiceNotRegistered {
			call.Results.SetAlive(false)
			return nil
		}
		return err
	}
	call.Results.SetAlive(service.Alive())
	return nil
}

func (a ServiceServer) Kill(call capnprpc.Service_kill) error {
	service, err := GetService(a.ServiceID)
	if err != nil {
		return err
	}
	service.Kill(nil)
	return nil
}
