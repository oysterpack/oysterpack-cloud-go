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
	"net"
	"testing"

	"time"

	"bytes"

	"io"

	"github.com/oysterpack/oysterpack.go/pkg/app"
	"github.com/oysterpack/oysterpack.go/pkg/app/message"
	opnet "github.com/oysterpack/oysterpack.go/pkg/app/net"
	"github.com/oysterpack/oysterpack.go/pkg/app/uid"
	"zombiezen.com/go/capnproto2"
)

// The goal is to benchmark the capnp messaging via the Server. The server will simply read the message.Message and echo
// it back to the client.
//
// BenchmarkServer/REQUEST-RESPONSE_-_empty_data_payload-8                    50000             42313 ns/op               0 B/op          0 allocs/op
// BenchmarkServer/REQUEST-RESPONSE_-_with_data-8                             30000             43117 ns/op               0 B/op          0 allocs/op
// BenchmarkServer/ENCODING-DECODING-8                                      2000000               752 ns/op               0 B/op          0 allocs/op
//
// 	- that translates to ~26K/sec RPC calls - and that's with TLS !!!
// 	- the message marshalling overhead was ~1.76%
func BenchmarkServer(b *testing.B) {
	app.Reset()
	defer app.Reset()

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

	// Given an ServerSpec for the app RPCService
	_, seg, err := capnp.NewMessage(capnp.SingleSegment(nil))
	if err != nil {
		b.Fatal(err)
	}
	serviceSpec, err := PKI.ServiceSpec(seg, DOMAIN_ID, APP_ID, SERVICE_ID, PORT)
	if err != nil {
		b.Fatal(err)
	}
	serverSpec, err := PKI.ServerSpec(seg, serviceSpec, MAX_CONNS)
	if err != nil {
		b.Fatal(err)
	}

	// Then the server can be started
	service := app.NewService(SERVICE_ID)
	app.Services.Register(service)
	serverSettings, err := opnet.NewServerSettings(service, serverSpec, func(conn net.Conn) {
		defer conn.Close()
		service.Logger().Info().Msg("New Connection")

		encoder := capnp.NewPackedEncoder(conn)
		decoder := capnp.NewPackedDecoder(conn)
		decoder.ReuseBuffer()

		service.Logger().Info().Msg("Connection handler is initialized")
		for {
			msg, err := decoder.Decode()
			if err != nil {
				if err == io.EOF {
					return
				}
				if service.Alive() || app.Alive() {
					service.Logger().Error().Err(err).Msgf("decoder.Decode() failed : %T", err)
				}
				return
			}
			// we want to make sure we can read the message
			_, err = message.ReadRootMessage(msg)
			if err != nil {
				service.Logger().Error().Err(err).Msg("message.ReadRootMessage(msg) failed")
				return
			}
			// echo back the message
			if err := encoder.Encode(msg); err != nil {
				if service.Alive() || app.Alive() {
					service.Logger().Error().Err(err).Msgf("decoder.Decode() failed : %T", err)
				}
				return
			}
		}
	})
	if err != nil {
		b.Fatal(err)
	}

	// Start the server using the settings
	server, err := opnet.StartServer(serverSettings)
	if err != nil {
		b.Fatal(err)
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
		b.Fatal("Server is not running")
	}

	// Given a ClientSpec for the Server
	_, seg, err = capnp.NewMessage(capnp.SingleSegment(nil))
	if err != nil {
		b.Fatal(err)
	}

	clientSpecConfig, err := PKI.ClientSpec(seg, DOMAIN_ID, APP_ID, SERVICE_ID, PORT, CLIENT_CN)
	if err != nil {
		b.Fatal(err)
	}
	clientSpec, err := opnet.NewClientSpec(clientSpecConfig)
	if err != nil {
		b.Fatal(err)
	}

	// Then we are able to connect to the RPC server
	addr, err := server.Address()
	if err != nil {
		b.Fatal(err)
	}
	b.Logf("server addresss : %v:%v", addr.Network(), addr.String())

	//addr := fmt.Sprintf(":%d", clientSpec.ServerPort)
	clientConn, err := clientSpec.ConnForAddr("") //tls.Dial("tcp", addr, clientSpec.TLSConfig())
	if err != nil {
		b.Fatal(err)
	}

	encoder := capnp.NewPackedEncoder(clientConn)
	decoder := capnp.NewPackedDecoder(clientConn)
	decoder.ReuseBuffer()

	msg, seg, err := capnp.NewMessage(capnp.SingleSegment(nil))
	if err != nil {
		b.Fatal(err)
	}
	req, err := message.NewRootMessage(seg)
	if err != nil {
		b.Fatal(err)
	}
	req.SetId(uint64(uid.NextUIDHash()))
	req.SetType(uint64(uid.NextUIDHash()))
	req.SetCorrelationID(uint64(uid.NextUIDHash()))
	req.SetTimestamp(uint64(time.Now().UnixNano()))

	b.Run("REQUEST-RESPONSE - empty data payload", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			// write request
			if err := encoder.Encode(msg); err != nil {
				b.Fatal(err)
			}

			// read response
			_, err := decoder.Decode()
			if err != nil {
				b.Fatal(err)
			}
		}
	})

	if err := req.SetData([]byte("DATA")); err != nil {
		b.Fatal(err)
	}
	b.Run("REQUEST-RESPONSE - with data", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			// write request
			if err := encoder.Encode(msg); err != nil {
				b.Fatal(err)
			}

			// read response
			_, err := decoder.Decode()
			if err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("ENCODING-DECODING", func(b *testing.B) {
		buf := new(bytes.Buffer)
		encoder := capnp.NewPackedEncoder(buf)
		decoder := capnp.NewPackedDecoder(buf)
		decoder.ReuseBuffer()

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			buf.Reset()
			// write request
			if err := encoder.Encode(msg); err != nil {
				b.Fatal(err)
			}

			// read response
			respMsg, err := decoder.Decode()
			if err != nil {
				b.Fatal(err)
			}

			// reading the message to ensure it was mrshalled / unmarshalled properly
			// also we want to include this in the benchmark because the server will always read the message
			_, err = message.ReadRootMessage(respMsg)
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}
