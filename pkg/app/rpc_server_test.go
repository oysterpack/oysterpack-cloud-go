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
	"sync"
	"time"

	"github.com/oysterpack/oysterpack.go/pkg/app"
	"github.com/oysterpack/oysterpack.go/pkg/app/capnprpc"
	"github.com/rs/zerolog/log"
	"gopkg.in/tomb.v2"
	"zombiezen.com/go/capnproto2/rpc"
)

func TestRpcAppServer(t *testing.T) {
	app.Reset()
	sockAddress := fmt.Sprintf("/tmp/app-%s.sock", app.InstanceId())
	t.Logf("sockAddress = %s", sockAddress)

	l, err := net.Listen("unix", sockAddress)
	if err != nil {
		t.Fatal(err)
	}

	tombWait := sync.WaitGroup{}
	tombWait.Add(2)

	defer func() {
		l.Close()
		tombWait.Wait()
	}()

	var listenerTomb tomb.Tomb
	var rpcTomb tomb.Tomb
	go func() {
		defer tombWait.Done()
		err := rpcTomb.Wait()
		if err != nil {
			log.Logger.Error().Err(err).Msgf("rpcTomb - DEAD: %T", err)
			return
		}
		log.Logger.Info().Msg("rpcTomb - DEAD")
	}()
	go func() {
		defer tombWait.Done()
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
		t.Fatal(err)
	}
	rpcClient := rpc.NewConn(rpc.StreamTransport(clientConn))
	defer rpcClient.Close()

	ctx := context.Background()
	appClient := capnprpc.App{Client: rpcClient.Bootstrap(ctx)}

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
