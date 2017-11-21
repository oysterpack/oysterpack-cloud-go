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

package main

import (
	"context"
	"fmt"

	"crypto/tls"

	"github.com/oysterpack/oysterpack.go/pkg/app"
	"github.com/oysterpack/oysterpack.go/pkg/app/capnprpc"
	"zombiezen.com/go/capnproto2/rpc"
)

func main() {
	defer func() {
		app.Kill()
		<-app.Dead()
	}()

	const (
		//host = "192.168.99.101"
		host = ""
		port = 44228
	)

	tlsConfig, err := app.ClientTLSConfig(
		"/home/alfio/go/src/github.com/oysterpack/oysterpack.go/pkg/app/testdata/.easypki/pki",
		"app.dev.oysterpack.com",
		"client.dev.oysterpack.com",
	)
	if err != nil {
		app.Logger().Fatal().Err(err).Msg("Failed to load TLSConfig")
	}
	conn, err := tls.Dial("tcp", fmt.Sprintf("%s:%d", host, port), tlsConfig)

	//conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", host, port))

	if err != nil {
		app.Logger().Fatal().Err(err).Msg("Failed to connect to RPC server")
	}

	rpcClient := rpc.NewConn(rpc.StreamTransport(conn))

	ctx := context.Background()
	appClient := capnprpc.App{Client: rpcClient.Bootstrap(ctx)}

	if result, err := appClient.Id(ctx, func(params capnprpc.App_id_Params) error {
		return nil
	}).Struct(); err != nil {
		app.Logger().Error().Err(err).Msg("Failed to make RPC Id() call")
	} else {
		app.Logger().Info().Msgf("app id : %x", result.AppId())
	}

}
