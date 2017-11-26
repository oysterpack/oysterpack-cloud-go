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
	"context"
	"net"
	"testing"
	"time"

	"github.com/oysterpack/oysterpack.go/pkg/app"
	"github.com/oysterpack/oysterpack.go/pkg/app/capnprpc"
	"github.com/rs/zerolog/log"
	"zombiezen.com/go/capnproto2"
)

func TestStartRPCService(t *testing.T) {
	app.Reset()

	// Given the app RPC service is registered with maxConns = 5
	appRPCService := app.NewService(app.ServiceID(0xfef711bb74ee4e13))
	app.Services.Register(appRPCService)

	var listenerFactory app.ListenerFactory = func() (net.Listener, error) {
		return net.Listen("tcp", ":0")
	}

	const maxConns = 5
	rpcService, err := app.StartRPCService(appRPCService, listenerFactory, nil, appRPCMainInterface(), maxConns)
	if err != nil {
		t.Fatal(err)
	}

	if rpcService.MaxConns() != maxConns {
		t.Errorf("MaxConns did not match : %d", rpcService.MaxConns())
	}
	if rpcService.ActiveConns() != 0 {
		t.Errorf("There should be no connections : %d", rpcService.ActiveConns())
	}

	<-rpcService.Started()
	addr, err := rpcService.ListenerAddress()
	if err != nil {
		t.Fatal(err)
	}

	client, conn, err := appClientConn(addr)
	if err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()
	if result, err := client.Id(ctx, func(params capnprpc.App_id_Params) error { return nil }).Struct(); err != nil {
		t.Error(err)
	} else {
		t.Logf("app id : %v", result.AppId())
	}

	if rpcService.ActiveConns() != 1 {
		t.Errorf("conn count does not match : %v", rpcService.ActiveConns())
	}
	// When the client is closed
	client.Client.Close()
	time.Sleep(time.Millisecond * 50)
	// Then the network connection will remain connected
	if rpcService.ActiveConns() != 1 {
		t.Errorf("conn count does not match : %v", rpcService.ActiveConns())
	}
	// In order to close the network connection, the connection needs to be explicitly closed.
	conn.Close()
	for rpcService.ActiveConns() != 0 {
		log.Logger.Info().Msg("Waiting for connection to close ...")
		time.Sleep(time.Millisecond * 50)
	}

	clients := []capnprpc.App{}
	conns := []net.Conn{}
	for i := 0; i < maxConns; i++ {
		client, conn, err := appClientConn(addr)
		if err != nil {
			t.Fatal(err)
		}
		clients = append(clients, client)
		conns = append(conns, conn)
	}

	t.Logf("Total conns created : %d", rpcService.TotalConnsCreated())

	for rpcService.ActiveConns() != maxConns {
		t.Logf("Waiting for server side connections to establish : %v != %v", rpcService.ActiveConns(), maxConns)
		time.Sleep(time.Millisecond * 10)
	}

	appRPCService.Kill(nil)
	appRPCService.Wait()
}

func TestRPCService_TLS(t *testing.T) {
	app.Reset()
	defer app.Reset()

	var tlsProvider TLSProvider = EasyPKITLS{}

	// Given the app RPC service is registered with maxConns = 5
	appRPCService := app.NewService(app.ServiceID(0xfef711bb74ee4e13))
	app.Services.Register(appRPCService)

	var listenerFactory app.ListenerFactory = func() (net.Listener, error) {
		return net.Listen("tcp", "")
	}

	var tlsConfigProvider app.TLSConfigProvider = tlsProvider.ServerTLSConfig

	const maxConns = 5
	rpcService, err := app.StartRPCService(appRPCService, listenerFactory, tlsConfigProvider, appRPCMainInterface(), maxConns)
	if err != nil {
		t.Fatal(err)
	}

	<-rpcService.Started()
	addr, err := rpcService.ListenerAddress()
	if err != nil {
		t.Fatal(err)
	}

	appClient, conn, err := appTLSClientConn(addr)
	if err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()

	if result, err := appClient.Id(ctx, func(params capnprpc.App_id_Params) error {
		return nil
	}).Struct(); err != nil {
		t.Error(err)
	} else {
		t.Logf("app id : %x", result.AppId())
	}

	if result, err := appClient.Instance(ctx, func(params capnprpc.App_instance_Params) error {
		return nil
	}).Struct(); err != nil {
		t.Error(err)
	} else {
		if instanceId, err := result.InstanceId(); err != nil {
			t.Error(err)
		} else {
			t.Logf("app instance id : %s", instanceId)
		}

	}

	// When the conn is closed
	conn.Close()
	if _, err := appClient.Id(ctx, func(params capnprpc.App_id_Params) error {
		return nil
	}).Struct(); err == nil {
		t.Error("Client RPC call should have failed because the underlying connection has been closed.")
	} else {
		t.Logf("after the network connection is closed, the RPC client fails with the following error : %[1]T : %[1]v", err)
	}
}

func appRPCMainInterface() app.RPCMainInterface {
	server := capnprpc.App_ServerToClient(app.NewAppServer())
	return func() (capnp.Client, error) {
		return server.Client, nil
	}
}

func BenchmarkRPCService(b *testing.B) {
	app.Reset()
	defer app.Reset()

	ctx := context.Background()
	idParams := func(params capnprpc.App_id_Params) error { return nil }

	// Given the app RPC service is registered with maxConns = 5
	appRPCService := app.NewService(app.ServiceID(1))
	app.Services.Register(appRPCService)

	var listenerFactory app.ListenerFactory = func() (net.Listener, error) {
		return net.Listen("tcp", ":0")
	}

	const maxConns = 5
	rpcService, err := app.StartRPCService(appRPCService, listenerFactory, nil, appRPCMainInterface(), maxConns)
	if err != nil {
		b.Fatal(err)
	}
	<-rpcService.Started()
	addr, err := rpcService.ListenerAddress()
	if err != nil {
		b.Fatal(err)
	}
	appClient, _, err := appClientConn(addr)
	if err != nil {
		b.Fatal(err)
	}

	// BenchmarkRPCService/ID()_-_no_TLS-8                20000             91676 ns/op           18267 B/op         78 allocs/op
	b.Run("ID() - no TLS", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			if result, err := appClient.Id(ctx, idParams).Struct(); err != nil {
				b.Fatal(err)
			} else {
				result.AppId()
			}
		}
	})

	app.Reset()

	appRPCService = app.NewService(app.ServiceID(0xfef711bb74ee4e13))
	app.Services.Register(appRPCService)

	var tlsConfigProvider app.TLSConfigProvider = tlsProvider.ServerTLSConfig
	rpcService, err = app.StartRPCService(appRPCService, listenerFactory, tlsConfigProvider, appRPCMainInterface(), maxConns)
	if err != nil {
		b.Fatal(err)
	}
	<-rpcService.Started()
	addr, err = rpcService.ListenerAddress()
	if err != nil {
		b.Fatal(err)
	}
	appClient, _, err = appTLSClientConn(addr)
	if err != nil {
		b.Fatal(err)
	}

	// BenchmarkRPCService/ID()_-_TLS-8                   10000            115054 ns/op           18266 B/op         78 allocs/op
	b.Run("ID() - TLS", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			if result, err := appClient.Id(ctx, idParams).Struct(); err != nil {
				b.Fatal(err)
			} else {
				result.AppId()
			}
		}
	})
}
