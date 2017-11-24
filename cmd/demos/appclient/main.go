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

	"time"

	"github.com/oysterpack/oysterpack.go/pkg/app"
)

func main() {

	go TestRPCAppClient()

	<-app.Dead()
}

func TestRPCAppClient() {
	ctx := context.Background()

	connect := func() (*app.AppRPCClient, error) {
		const APP_RPC_SERVICE_CLIENT_ID = app.ServiceID(0xdb6c5b7c386221bc)
		return app.NewAppClient(APP_RPC_SERVICE_CLIENT_ID)
		//return app.NewAppClientForAddr(APP_RPC_SERVICE_CLIENT_ID, "") // for local testing
	}

	appClient, err := connect()
	if err != nil {
		app.Logger().Error().Err(err).Msg("Failed to create App RPCService client")
	} else {
		app.Logger().Info().Msg("App RPCService client is connected")

		if result, err := appClient.Id(ctx).Struct(); err != nil {
			app.Logger().Error().Err(err).Msg("RPC App.Id() error")
		} else {
			app.Logger().Info().Msgf("app id : %x", result.AppId())
		}
	}

	reconnect := func() {
		app.Logger().Error().Err(err).Msg("RPC App.Id() error")

		appClient, err = connect()
		if err != nil {
			app.Logger().Error().Err(err).Msg("Failed to create App RPCService client")
		} else {
			app.Logger().Info().Msg("App RPCService client is connected")
		}
	}

	ticker := time.NewTicker(time.Second * 5)
	for {
		select {
		case <-app.Dying():
			return
		case <-ticker.C:
			if appClient == nil {
				appClient, err = connect()
				if err != nil {
					app.Logger().Error().Err(err).Msg("Failed to create App RPCService client")
				} else {
					app.Logger().Info().Msg("App RPCService client is connected")
				}
			} else {
				idPromise := appClient.Id(ctx)
				instancePromise := appClient.Instance(ctx)

				if result, err := idPromise.Struct(); err != nil {
					reconnect()
				} else {
					appId := result.AppId()
					if result, err := instancePromise.Struct(); err != nil {
						reconnect()
					} else {
						if instanceID, err := result.InstanceId(); err != nil {
							reconnect()
						} else {
							app.Logger().Info().Msgf("app id : %x, instance id: %s", appId, instanceID)
						}
					}
				}
			}
		}
	}
}
