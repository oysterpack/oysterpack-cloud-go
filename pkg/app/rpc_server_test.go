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

package app_test

import (
	"fmt"
	"net"
	"testing"

	"context"
	"time"

	"github.com/oysterpack/oysterpack.go/pkg/app"
	"github.com/oysterpack/oysterpack.go/pkg/app/capnprpc"
	"github.com/rs/zerolog/log"
	"gopkg.in/tomb.v2"
	"zombiezen.com/go/capnproto2/rpc"
)

func startRPCAppServer() (listener net.Listener, client capnprpc.App) {
	app.Reset()

	sockAddress := fmt.Sprintf("/tmp/app-%s.sock", app.InstanceId())
	log.Logger.Info().Msgf("sockAddress = %s", sockAddress)

	l, err := net.Listen("unix", sockAddress)
	if err != nil {
		panic(err)
	}

	var listenerTomb tomb.Tomb
	var rpcTomb tomb.Tomb
	go func() {
		err := rpcTomb.Wait()
		if err != nil {
			log.Logger.Error().Err(err).Msgf("rpcTomb - DEAD: %T", err)
			return
		}
		log.Logger.Info().Msg("rpcTomb - DEAD")
	}()
	go func() {
		err := listenerTomb.Wait()
		if err != nil {
			log.Logger.Error().Err(err).Msgf("listenerTomb - DEAD : %T", err)
			return
		}
		log.Logger.Info().Msg("listenerTomb - DEAD")
	}()

	listenerTomb.Go(func() error {
		server := capnprpc.App_ServerToClient(app.NewAppServer())
		for {
			conn, err := l.Accept()
			if err != nil {
				return err
			}
			rpcTomb.Go(func() error {
				rpcConn := rpc.NewConn(rpc.StreamTransport(conn), rpc.MainInterface(server.Client))
				return rpcConn.Wait()
			})
		}
	})

	clientConn, err := net.Dial("unix", sockAddress)
	if err != nil {
		panic(err)
	}
	rpcClient := rpc.NewConn(rpc.StreamTransport(clientConn))

	ctx := context.Background()
	return l, capnprpc.App{Client: rpcClient.Bootstrap(ctx)}
}

func TestRpcAppServer_LogLevel(t *testing.T) {
	l, appClient := startRPCAppServer()
	defer l.Close()
	defer appClient.Client.Close()

	ctx := context.Background()

	if result, err := appClient.LogLevel(ctx, func(params capnprpc.App_logLevel_Params) error {
		return nil
	}).Struct(); err != nil {
		t.Error(err)
	} else {
		 if level, err := app.CapnprpcLogLevel2zerologLevel(result.Level()); err != nil {
		 	t.Error(err)
		 } else {
		 	t.Logf("level : %v", level)
		 }
	}

	appClient.SetLogLevel(ctx, func(params capnprpc.App_setLogLevel_Params) error {
		params.SetLevel(capnprpc.LogLevel_warn)
		return  nil
	})

	if result, err := appClient.LogLevel(ctx, func(params capnprpc.App_logLevel_Params) error {
		return nil
	}).Struct(); err != nil {
		t.Error(err)
	} else {
		if result.Level() != capnprpc.LogLevel_warn {
			t.Errorf("Level does not match : %v",result.Level())
		}
	}
}

func TestRpcAppServer(t *testing.T) {
	l, appClient := startRPCAppServer()
	defer l.Close()
	defer appClient.Client.Close()

	ctx := context.Background()

	if result, err := appClient.Id(ctx, func(params capnprpc.App_id_Params) error {
		return nil
	}).Struct(); err != nil {
		t.Error(err)
	} else {
		t.Logf("app id : %v", result.AppId())
		if app.AppID(result.AppId()) != app.ID() {
			t.Errorf("Returned app id did not match : %v != %v", result.AppId(), app.ID())
		}
	}

	if result, err := appClient.StartedOn(ctx, func(params capnprpc.App_startedOn_Params) error {
		return nil
	}).Struct(); err != nil {
		t.Error(err)
	} else {
		startedOn := time.Unix(0, result.StartedOn())
		t.Logf("startedOn: %v", startedOn)
		if !startedOn.Equal(app.StartedOn()) {
			t.Errorf("Returned app startedOn did not match : %v != %v", startedOn, app.StartedOn())
		}
	}

	if result, err := appClient.Instance(ctx, func(params capnprpc.App_instance_Params) error {
		return nil
	}).Struct(); err != nil {
		t.Error(err)
	} else {
		instanceId, err := result.InstanceId()
		if err != nil {
			t.Error(err)
		} else {
			t.Logf("instance id : %v", instanceId)
			if app.InstanceID(instanceId) != app.InstanceId() {
				t.Errorf("Returned app instance id did not match : %v != %v", instanceId, app.InstanceId())
			}
		}
	}

	service:= app.NewService(app.ServiceID(999))
	app.RegisterService(service)
	remoteService := appClient.Service(ctx, func(params capnprpc.App_service_Params) error {
		params.SetId(uint64(service.ID()))
		return nil
	}).Service()

	if results, err :=remoteService.Id(ctx, func(params capnprpc.Service_id_Params) error {
		return nil
	}).Struct(); err != nil {
		t.Error(err)
	} else {
		t.Logf("remote service id = %v",results.ServiceId())
		if results.ServiceId() != uint64(service.ID()){
			t.Errorf("service id did not match : %v", results.ServiceId())
		}
	}

	// When an invalid ServiceID is used to lookup a Service
	// Then an error should be returned.
	if _, err := appClient.Service(ctx, func(params capnprpc.App_service_Params) error {
		params.SetId(uint64(service.ID()+1))
		return nil
	}).Struct(); err == nil {
		t.Error("No service should have been found")
	} else {
		t.Logf("rpc error : %T : %v", err,err)
	}

	remoteService = appClient.Service(ctx, func(params capnprpc.App_service_Params) error {
		params.SetId(uint64(service.ID()+1))
		return nil
	}).Service()

	if _, err := remoteService.Id(ctx, func(params capnprpc.Service_id_Params) error {
		return nil
	}).Struct(); err == nil {
		t.Error("The service references should be invalid")
	} else {
		t.Logf("rpc error : %T : %v", err,err)
	}

	// When the service is registered
	service1000 := app.NewService(app.ServiceID(service.ID()+1))
	if err := app.RegisterService(service1000); err != nil {
		t.Fatal(err)
	}

	// Then the remote service reference is still invalid, because the error is cached in the pipeline
	if _, err := remoteService.Id(ctx, func(params capnprpc.Service_id_Params) error {
		return nil
	}).Struct(); err == nil {
		t.Errorf("the remote sevice reference is invalid : %T : %v", err,err)
	} else {
		t.Logf("remote error : %T : %v",err, err)
	}

	// When a new remote service reference is retrieved
	remoteService = appClient.Service(ctx, func(params capnprpc.App_service_Params) error {
		params.SetId(uint64(service.ID()+1))
		return nil
	}).Service()

	// Then it works fine
	if result, err := remoteService.Id(ctx, func(params capnprpc.Service_id_Params) error {
		return nil
	}).Struct(); err != nil {
		t.Errorf("service should now be registered : %T : %v", err,err)
	} else {
		if app.ServiceID(result.ServiceId()) != service1000.ID() {
			t.Errorf("Service id did not match : %d", result.ServiceId())
		}
	}

	// When the service is killed
	service1000.Kill(nil)
	service1000.Wait()

	// Then the remote service reference becomes invalid
	if result, err := remoteService.LogLevel(ctx, func(params capnprpc.Service_logLevel_Params) error {
		return nil
	}).Struct(); err == nil {
		t.Errorf("The service reference should be invalid : %v", result.Level())
	} else {
		t.Logf("rpc error : %T : %v", err,err)
	}


	// Given a registered service
	service1000 = app.NewService(service1000.ID())
	if err := app.RegisterService(service1000); err != nil {
		t.Fatal(err)
	}
	// When the service is killed remotely
	remoteService = appClient.Service(ctx, func(params capnprpc.App_service_Params) error {
		params.SetId(uint64(service1000.ID()))
		return nil
	}).Service()

	if result, err := remoteService.Alive(ctx, func(params capnprpc.Service_alive_Params) error {
		return nil
	}).Struct(); err != nil {
		t.Error(err)
	} else if !result.Alive() {
		t.Error("service should be alive")
	}

	remoteService.Kill(ctx, func(params capnprpc.Service_kill_Params) error {
		return nil
	})

	if result, err := remoteService.Alive(ctx, func(params capnprpc.Service_alive_Params) error {
		return nil
	}).Struct(); err != nil {
		t.Error(err)
	} else if result.Alive() {
		t.Error("service should not be alive")
	}

}

// BenchmarkRpcAppServer_UnixSocket/ID()-8                    20000             79748 ns/op           18266 B/op         78 allocs/op
// BenchmarkRpcAppServer_UnixSocket/Instance()-8              20000             81747 ns/op           18361 B/op         79 allocs/op
// BenchmarkRpcAppServer_UnixSocket/StartedOn()-8             20000             79975 ns/op           18265 B/op         78 allocs/op
// BenchmarkRpcAppServer_UnixSocket/connect_-_close-8         10000            123790 ns/op           32753 B/op        174 allocs/op
func BenchmarkRpcAppServer_UnixSocket(b *testing.B) {
	app.Reset()
	sockAddress := fmt.Sprintf("/tmp/app-%s.sock", app.InstanceId())
	b.Logf("sockAddress = %s", sockAddress)

	// start the server
	l, err := net.Listen("unix", sockAddress)
	if err != nil {
		b.Fatal(err)
	}
	defer l.Close()
	var listenerTomb tomb.Tomb
	var rpcTomb tomb.Tomb
	listenerTomb.Go(func() error {
		server := capnprpc.App_ServerToClient(app.NewAppServer())
		for {
			conn, err := l.Accept()
			if err != nil {
				return err
			}
			rpcTomb.Go(func() error {
				rpcConn := rpc.NewConn(rpc.StreamTransport(conn), rpc.MainInterface(server.Client))
				return rpcConn.Wait()
			})
		}
	})

	// create client
	clientConn, err := net.Dial("unix", sockAddress)
	if err != nil {
		b.Fatal(err)
	}
	rpcClient := rpc.NewConn(rpc.StreamTransport(clientConn))
	defer rpcClient.Close()
	ctx := context.Background()
	appClient := capnprpc.App{Client: rpcClient.Bootstrap(ctx)}

	appIdParams := func(params capnprpc.App_id_Params) error { return nil }
	b.Run("ID()", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			if result, err := appClient.Id(ctx, appIdParams).Struct(); err != nil {
				b.Error(err)
			} else {
				result.AppId()
			}
		}
	})

	instanceParams := func(params capnprpc.App_instance_Params) error { return nil }
	b.Run("Instance()", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			if result, err := appClient.Instance(ctx, instanceParams).Struct(); err != nil {
				b.Error(err)
			} else {
				result.InstanceId()
			}
		}
	})

	startedOnParams := func(params capnprpc.App_startedOn_Params) error { return nil }
	b.Run("StartedOn()", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			if result, err := appClient.StartedOn(ctx, startedOnParams).Struct(); err != nil {
				b.Error(err)
			} else {
				result.StartedOn()
			}
		}
	})

	b.Run("connect - close", func(b *testing.B) {
		ctx := context.Background()
		for i := 0; i < b.N; i++ {
			clientConn, err := net.Dial("unix", sockAddress)
			if err != nil {
				b.Fatal(err)
			}
			rpcClient := rpc.NewConn(rpc.StreamTransport(clientConn))
			appClient := capnprpc.App{Client: rpcClient.Bootstrap(ctx)}
			appClient.Client.Close()
		}
	})
}
