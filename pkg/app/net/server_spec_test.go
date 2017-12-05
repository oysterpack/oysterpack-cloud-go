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

package net_test

import (
	"testing"

	"net"

	"time"

	"fmt"

	"github.com/oysterpack/oysterpack.go/pkg/app"
	"github.com/oysterpack/oysterpack.go/pkg/app/message"
	opnet "github.com/oysterpack/oysterpack.go/pkg/app/net"
	"github.com/oysterpack/oysterpack.go/pkg/app/uid"
	"zombiezen.com/go/capnproto2"
)

func TestNewServerSpec(t *testing.T) {
	const (
		// *** THE EASYPKI CA AND CERTS NEED TO PREEXIST ***
		// For now, use the testdata/.easypki/generate-certs.sh script to bootstrap
		// Ideally, we will want to to generate these on demand, i.e., programatically.
		DOMAIN_ID  = app.DomainID(0xed5cf026e8734361)
		APP_ID     = app.AppID(0xd113a2e016e12f0f)
		SERVICE_ID = app.ServiceID(0xe49214fa20b35ba8)

		PORT = opnet.ServerPort(44222)

		CLIENT_CN = "client.dev.oysterpack.com"

		MAX_CONNS = 16
	)

	configDir := "./testdata/server_spec_test/TestNewServerSpec"
	initConfigDir(configDir)
	initServerMetricsConfig(SERVICE_ID)
	app.ResetWithConfigDir(configDir)
	defer app.Reset()

	// Given an ServerSpec for the app RPCService
	_, seg, err := capnp.NewMessage(capnp.SingleSegment(nil))
	if err != nil {
		t.Fatal(err)
	}
	serviceSpec, err := PKI.ServiceSpec(seg, DOMAIN_ID, APP_ID, SERVICE_ID, PORT)
	if err != nil {
		t.Fatal(err)
	}
	serverSpec, err := PKI.ServerSpec(seg, serviceSpec, MAX_CONNS)
	if err != nil {
		t.Fatal(err)
	}

	// Then the server can be started
	service := app.NewService(SERVICE_ID)
	app.Services.Register(service)
	serverSettings, err := opnet.NewServerSettings(service, serverSpec, func(conn net.Conn) {
		defer conn.Close()
		service.Logger().Info().
			Str("remote-addr", conn.RemoteAddr().String()).
			Str("local-addr", conn.LocalAddr().String()).
			Msg("New Connection")

		encoder := capnp.NewPackedEncoder(conn)
		decoder := capnp.NewPackedDecoder(conn)
		decoder.ReuseBuffer()

		msgCounter := 0

		service.Logger().Info().Msg("Connection handler is initialized")
		for {

			msg, err := decoder.Decode()
			if err != nil {
				if service.Alive() {
					service.Logger().Error().Err(err).Msgf("decoder.Decode() failed : %T", err)
				}
				return
			}
			msgCounter++
			service.Logger().Info().Msgf("Received message #%d", msgCounter)

			m, err := message.ReadRootMessage(msg)
			if err != nil {
				service.Logger().Error().Err(err).Msg("message.ReadRootMessage(msg) failed")
				return
			}

			service.Logger().Info().Msgf("MSG#%d : parsed", msgCounter)

			data, err := m.Data()
			if err != nil {
				service.Logger().Error().Err(err).Msg("Message.Data() failed")
				return
			}
			service.Logger().Info().Msgf("request message : id = 0x%x : type = 0x%x, timestamp = %v, correlationId = 0x%x, len(data) = %d",
				m.Id(),
				m.Type(),
				time.Unix(0, int64(m.Timestamp())),
				m.CorrelationID(),
				len(data),
			)

			// echo back the message
			encoder.Encode(msg)
			service.Logger().Info().Msgf("MSG#%d : response sent", msgCounter)
		}
	})
	if err != nil {
		t.Fatal(err)
	}

	// Start the server using the settings
	server, err := opnet.StartServer(serverSettings)
	if err != nil {
		t.Fatal(err)
	}

WAIT_FOR_SERVER_RUNNING:
	for i := 0; i < 5; i++ {
		select {
		case <-server.Running():
			app.Logger().Info().Msg("Server is running")
			break WAIT_FOR_SERVER_RUNNING
		case <-time.After(time.Second):
			app.Logger().Warn().Msg("Waiting for Server to run")
			continue WAIT_FOR_SERVER_RUNNING
		}
	}
	select {
	case <-server.Running():
	default:
		t.Fatal("Server is not running")
	}

	// Given a ClientSpec for the Server
	_, seg, err = capnp.NewMessage(capnp.SingleSegment(nil))
	if err != nil {
		t.Fatal(err)
	}

	clientSpecConfig, err := PKI.ClientSpec(seg, DOMAIN_ID, APP_ID, SERVICE_ID, PORT, CLIENT_CN)
	if err != nil {
		t.Fatal(err)
	}
	clientSpec, err := opnet.NewClientSpec(clientSpecConfig)
	if err != nil {
		t.Fatal(err)
	}

	// Then we are able to connect to the RPC server
	addr, err := server.Address()
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("server addresss : %v:%v", addr.Network(), addr.String())

	//addr := fmt.Sprintf(":%d", clientSpec.ServerPort)
	clientConn, err := clientSpec.ConnForAddr("") //tls.Dial("tcp", addr, clientSpec.TLSConfig())
	if err != nil {
		t.Fatal(err)
	}

	encoder := capnp.NewPackedEncoder(clientConn)
	decoder := capnp.NewPackedDecoder(clientConn)
	decoder.ReuseBuffer()

	const REQ_COUNT = 10
	for i := 0; i < REQ_COUNT; i++ {
		msg, seg, err := capnp.NewMessage(capnp.SingleSegment(nil))
		if err != nil {
			t.Fatal(err)
		}

		req, err := message.NewRootMessage(seg)
		if err != nil {
			t.Fatal(err)
		}

		req.SetId(uint64(uid.NextUIDHash()))
		req.SetType(uint64(uid.NextUIDHash()))
		req.SetTimestamp(time.Now().UnixNano())
		req.SetCorrelationID(uint64(uid.NextUIDHash()))
		if err := req.SetData([]byte(fmt.Sprintf("REQ #%d", i))); err != nil {
			t.Fatal(err)
		}

		// write request
		encoder.Encode(msg)

		// read response
		responseMsg, err := decoder.Decode()

		if err != nil {
			t.Fatal(err)
		}
		m, err := message.ReadRootMessage(responseMsg)
		if err != nil {
			t.Fatal(err)
		}
		data, err := m.Data()
		if err != nil {
			t.Fatal(err)
		}
		service.Logger().Info().Msgf("response message : id = 0x%x : type = 0x%x, timestamp = %v, correlationId = 0x%x, len(data) = %d, data = %s",
			m.Id(),
			m.Type(),
			time.Unix(0, int64(m.Timestamp())),
			m.CorrelationID(),
			len(data),
			string(data),
		)
	}

}
