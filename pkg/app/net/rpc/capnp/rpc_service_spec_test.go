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
	"testing"

	"io/ioutil"

	"time"

	"crypto/tls"
	"fmt"

	"context"

	"os"

	"github.com/oysterpack/oysterpack.go/pkg/app"
	"github.com/oysterpack/oysterpack.go/pkg/app/capnprpc"
	"github.com/oysterpack/oysterpack.go/pkg/app/config"
	"zombiezen.com/go/capnproto2"
	"zombiezen.com/go/capnproto2/rpc"
)

func TestNewRPCServerSpec(t *testing.T) {
	app.Reset()
	defer app.Reset()

	const (
		APP_ID    = app.AppID(0xd113a2e016e12f0f)
		DOMAIN_ID = app.DomainID(0xed5cf026e8734361)

		PORT = app.RPCPort(44222)

		CLIENT_CN = "client.dev.oysterpack.com"
	)

	// Given an RPCServerSpec for the app RPCService
	_, seg, err := capnp.NewMessage(capnp.SingleSegment(nil))
	if err != nil {
		t.Fatal(err)
	}
	serverSpecConfig, err := app.NewRPCServerSpec(rpcServerSpec(t, seg, rpcServiceSpec(t, seg, DOMAIN_ID, APP_ID, app.APP_RPC_SERVICE_ID, PORT)))
	if err != nil {
		t.Fatal(err)
	}

	// Then the RPC server can be started
	appRPCService := app.NewService(app.APP_RPC_SERVICE_ID)
	app.Services.Register(appRPCService)

	server := capnprpc.App_ServerToClient(app.NewAppServer())
	rpcMainInterface := func() (capnp.Client, error) {
		return server.Client, nil
	}
	rpcService, err := serverSpecConfig.StartRPCService(appRPCService, rpcMainInterface)
	if err != nil {
		t.Fatal(err)
	}
	if rpcService.MaxConns() != int(serverSpecConfig.MaxConns) {
		t.Errorf("RPCService max conns did not match : %d != %d", rpcService.MaxConns(), int(serverSpecConfig.MaxConns))
	}

	ticker := time.NewTicker(time.Millisecond * 100)
	select {
	case <-ticker.C:
		t.Error("RPC server did not start up")
	case <-rpcService.Started():
	}

	// Given an RPCClientSpec for the app RPCService
	_, seg, err = capnp.NewMessage(capnp.SingleSegment(nil))
	if err != nil {
		t.Fatal(err)
	}
	clientSpecConfig, err := app.NewRPCClientSpec(rpcClientSpec(t, seg, rpcServiceSpec(t, seg, DOMAIN_ID, APP_ID, app.APP_RPC_SERVICE_ID, PORT), CLIENT_CN))
	if err != nil {
		t.Fatal(err)
	}

	// Then we are able to connect to the RPC server
	addr := fmt.Sprintf(":%d", clientSpecConfig.RPCPort)
	clientConn, err := tls.Dial("tcp", addr, clientSpecConfig.TLSConfig())
	if err != nil {
		t.Fatal(err)
	}
	rpcConn := rpc.NewConn(rpc.StreamTransport(clientConn))
	appClient := capnprpc.App{Client: rpcConn.Bootstrap(context.Background())}
	if result, err := appClient.Id(context.Background(), func(params capnprpc.App_id_Params) error {
		return nil
	}).Struct(); err != nil {
		t.Fatal(err)
	} else {
		t.Logf("app id : %x", result.AppId())
	}

}

func TestMarshallingUnmarshallingConfig(t *testing.T) {
	app.Reset()
	defer app.Reset()

	const (
		APP_ID    = app.AppID(0xd113a2e016e12f0f)
		DOMAIN_ID = app.DomainID(0xed5cf026e8734361)

		PORT = app.RPCPort(44222)

		CLIENT_CN = "client.dev.oysterpack.com"

		APP_RPC_SERVICE_CLIENT_ID = app.ServiceID(0xdb6c5b7c386221bc)
	)

	configDir := "testdata/TestMarshallingUnmarshallingConfig"
	os.RemoveAll(configDir)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatal(err)
	}

	t.Run("Marshall app RPCService Server Spec", func(t *testing.T) {
		msg, seg, err := capnp.NewMessage(capnp.MultiSegment(nil))
		if err != nil {
			t.Fatal(err)
		}
		serverSpec := rpcServerSpec(t, seg, rpcServiceSpec(t, seg, DOMAIN_ID, APP_ID, app.APP_RPC_SERVICE_ID, PORT))
		if err := app.CheckRPCServerSpec(serverSpec); err != nil {
			t.Fatal(err)
		}
		if _, err = app.NewRPCServerSpec(serverSpec); err != nil {
			t.Fatal(err)
		}
		serverSpecConfigFile, err := os.Create(fmt.Sprintf("%s/0x%x", configDir, app.APP_RPC_SERVICE_ID))
		if err != nil {
			t.Fatal(err)
		}
		if err := app.MarshalCapnpMessage(msg, serverSpecConfigFile); err != nil {
			t.Fatal(err)
		}
		serverSpecConfigFile.Close()
	})

	t.Run("Unmarshall app RPCService Server Spec", func(t *testing.T) {
		serverSpecConfigFile, err := os.Open(fmt.Sprintf("%s/0x%x", configDir, app.APP_RPC_SERVICE_ID))
		msg, err := app.UnmarshalCapnpMessage(serverSpecConfigFile)
		serverSpecConfigFile.Close()
		if err != nil {
			t.Fatal(err)
		}
		serverSpec, err := config.ReadRootRPCServerSpec(msg)
		if err != nil {
			t.Fatal(err)
		}
		if _, err = app.NewRPCServerSpec(serverSpec); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("Marshal app RPCService Client Spec", func(t *testing.T) {
		msg, seg, err := capnp.NewMessage(capnp.SingleSegment(nil))
		if err != nil {
			t.Fatal(err)
		}
		if _, err = app.NewRPCClientSpec(rpcClientSpec(t, seg, rpcServiceSpec(t, seg, DOMAIN_ID, APP_ID, app.APP_RPC_SERVICE_ID, PORT), CLIENT_CN)); err != nil {
			t.Fatal(err)
		}
		clientSpecConfigFile, err := os.Create(fmt.Sprintf("%s/0x%x", configDir, APP_RPC_SERVICE_CLIENT_ID))
		if err != nil {
			t.Fatal(err)
		}
		if err := app.MarshalCapnpMessage(msg, clientSpecConfigFile); err != nil {
			t.Fatal(err)
		}
		clientSpecConfigFile.Close()
	})

	t.Run("Unmarshall app RPCService Client Spec", func(t *testing.T) {
		clientSpecConfigFile, err := os.Open(fmt.Sprintf("%s/0x%x", configDir, APP_RPC_SERVICE_CLIENT_ID))
		msg, err := app.UnmarshalCapnpMessage(clientSpecConfigFile)
		clientSpecConfigFile.Close()
		if err != nil {
			t.Fatal(err)
		}
		clientSpec, err := config.ReadRootRPCClientSpec(msg)
		if err != nil {
			t.Fatal(err)
		}

		if _, err := app.NewRPCClientSpec(clientSpec); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("Start App RPCService from Config", func(t *testing.T) {
		app.Reset()
		defer app.Reset()

		serverSpecConfigFile, err := os.Open(fmt.Sprintf("%s/0x%x", configDir, app.APP_RPC_SERVICE_ID))
		msg, err := app.UnmarshalCapnpMessage(serverSpecConfigFile)
		serverSpecConfigFile.Close()
		if err != nil {
			t.Fatal(err)
		}
		serverSpec, err := config.ReadRootRPCServerSpec(msg)
		if err != nil {
			t.Fatal(err)
		}
		serverSpecConfig, err := app.NewRPCServerSpec(serverSpec)
		if err != nil {
			t.Fatal(err)
		}

		// Then the RPC server can be started
		appRPCService := app.NewService(app.APP_RPC_SERVICE_ID)
		app.Services.Register(appRPCService)

		server := capnprpc.App_ServerToClient(app.NewAppServer())
		rpcMainInterface := func() (capnp.Client, error) {
			return server.Client, nil
		}
		rpcService, err := serverSpecConfig.StartRPCService(appRPCService, rpcMainInterface)
		if err != nil {
			t.Fatal(err)
		}
		if rpcService.MaxConns() != int(serverSpecConfig.MaxConns) {
			t.Errorf("RPCService max conns did not match : %d != %d", rpcService.MaxConns(), int(serverSpecConfig.MaxConns))
		}

		ticker := time.NewTicker(time.Millisecond * 100)
		select {
		case <-ticker.C:
			t.Error("RPC server did not start up")
		case <-rpcService.Started():
		}

		// Given an RPCClientSpec for the app RPCService
		clientSpecConfigFile, err := os.Open(fmt.Sprintf("%s/0x%x", configDir, APP_RPC_SERVICE_CLIENT_ID))
		msg, err = app.UnmarshalCapnpMessage(clientSpecConfigFile)
		clientSpecConfigFile.Close()
		if err != nil {
			t.Fatal(err)
		}
		clientSpec, err := config.ReadRootRPCClientSpec(msg)
		if err != nil {
			t.Fatal(err)
		}

		clientSpecConfig, err := app.NewRPCClientSpec(clientSpec)
		//if err != nil {
		//	t.Fatal(err)
		//}
		//
		//// Then we are able to connect to the RPC server
		//addr := fmt.Sprintf(":%d", clientSpecConfig.RPCPort)
		//clientConn, err := tls.Dial("tcp", addr, clientSpecConfig.TLSConfig())
		//if err != nil {
		//	t.Fatal(err)
		//}
		rpcConn, err := clientSpecConfig.ConnForAddr("") //rpc.NewConn(rpc.StreamTransport(clientConn))
		if err != nil {
			t.Fatal(err)
		}
		appClient := capnprpc.App{Client: rpcConn.Bootstrap(context.Background())}
		if result, err := appClient.Id(context.Background(), func(params capnprpc.App_id_Params) error {
			return nil
		}).Struct(); err != nil {
			t.Fatal(err)
		} else {
			t.Logf("app id : %x", result.AppId())
		}
	})
}

func x509KeyPair(t *testing.T, seg *capnp.Segment, ca, cn string) config.X509KeyPair {
	t.Helper()
	serverCert, err := ioutil.ReadFile(EasyPKICertFilePath(ca, cn))
	if err != nil {
		t.Fatal(err)
	}
	serverKey, err := ioutil.ReadFile(EasyPKIKeyFilePath(ca, cn))
	if err != nil {
		t.Fatal(err)
	}
	x509KeyPair, err := config.NewX509KeyPair(seg)
	if err != nil {
		t.Fatal(err)
	}
	if err := x509KeyPair.SetCert(serverCert); err != nil {
		t.Fatal(err)
	}
	if err := x509KeyPair.SetKey(serverKey); err != nil {
		t.Fatal(err)
	}
	return x509KeyPair
}

func rpcServiceSpec(t *testing.T, seg *capnp.Segment, domainID app.DomainID, appID app.AppID, serviceID app.ServiceID, port app.RPCPort) config.RPCServiceSpec {
	serviceSpec, err := config.NewRPCServiceSpec(seg)
	if err != nil {
		t.Fatal(err)
	}
	serviceSpec.SetDomainID(uint64(domainID))
	serviceSpec.SetAppId(uint64(appID))
	serviceSpec.SetServiceId(uint64(app.APP_RPC_SERVICE_ID))
	serviceSpec.SetPort(uint16(port))
	return serviceSpec
}

func rpcServerSpec(t *testing.T, seg *capnp.Segment, serviceSpec config.RPCServiceSpec) config.RPCServerSpec {
	serverSpec, err := config.NewRootRPCServerSpec(seg)
	if err != nil {
		t.Fatal(err)
	}
	serverSpec.SetMaxConns(20)

	serverSpec.SetRpcServiceSpec(serviceSpec)

	ca := fmt.Sprintf("%x", serviceSpec.DomainID())
	cacert, err := ioutil.ReadFile(EasyPKICertFilePath(ca, ca))
	if err != nil {
		t.Fatal(err)
	}
	if err := serverSpec.SetCaCert(cacert); err != nil {
		t.Fatal(err)
	}

	serverCN := app.ServerCN(app.DomainID(serviceSpec.DomainID()), app.AppID(serviceSpec.AppId()), app.ServiceID(serviceSpec.ServiceId()))
	if err := serverSpec.SetServerCert(x509KeyPair(t, seg, ca, serverCN)); err != nil {
		t.Fatal(err)
	}

	return serverSpec
}

func rpcClientSpec(t *testing.T, seg *capnp.Segment, serviceSpec config.RPCServiceSpec, cn string) config.RPCClientSpec {
	clientSpec, err := config.NewRootRPCClientSpec(seg)
	clientSpec.SetRpcServiceSpec(rpcServiceSpec(t, seg, app.DomainID(serviceSpec.DomainID()), app.AppID(serviceSpec.AppId()), app.ServiceID(serviceSpec.ServiceId()), app.RPCPort(serviceSpec.Port())))

	ca := fmt.Sprintf("%x", serviceSpec.DomainID())
	cacert, err := ioutil.ReadFile(EasyPKICertFilePath(ca, ca))
	if err != nil {
		t.Fatal(err)
	}
	if err := clientSpec.SetCaCert(cacert); err != nil {
		t.Fatal(err)
	}
	clientSpec.SetClientCert(x509KeyPair(t, seg, ca, cn))
	return clientSpec
}
