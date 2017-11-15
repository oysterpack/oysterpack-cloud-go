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
	app.RegisterService(appRPCService)

	var listenerFactory app.ListenerFactory = func() (net.Listener, error) {
		return net.Listen("tcp", ":0")
	}

	const maxConns = 5
	rpcService, err := app.StartRPCService(appRPCService, listenerFactory, appRPCMainInterface(), maxConns)
	if err != nil {
		t.Fatal(err)
	}

	if rpcService.MaxConns() != maxConns {
		t.Errorf("MaxConns did not match : %d", rpcService.MaxConns())
	}
	if rpcService.ActiveConns() != 0 {
		t.Errorf("There should be no connections : %d", rpcService.ActiveConns())
	}

	for {
		if _, err := rpcService.ListenerAddress(); err == nil {
			break
		}
		log.Logger.Warn().Msg("Waiting for listener to start up ...")
		time.Sleep(time.Millisecond * 50)
	}
	addr, err := rpcService.ListenerAddress()
	if err != nil {
		t.Fatal(err)
	}

	client, conn := appClientConn(addr)

	ctx := context.Background()
	if result, err := client.Id(ctx, func(params capnprpc.App_id_Params) error { return nil }).Struct(); err != nil {
		t.Error(err)
	} else {
		t.Log("app id : %v", result.AppId())
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
		client, conn := appClientConn(addr)
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

func appRPCMainInterface() func() (capnp.Client, error) {
	server := capnprpc.App_ServerToClient(app.NewAppServer())
	return func() (capnp.Client, error) {
		return server.Client, nil
	}
}
