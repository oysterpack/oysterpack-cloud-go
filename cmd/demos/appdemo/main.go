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

import "github.com/oysterpack/oysterpack.go/pkg/app"

// ./appdemo -app-id 0xd8e06a3d73a1d426 -release-id 2 -tags A,B,C -app-rpc-port 44228 -log-level DEBUG \
// -pki-root /pki -tls-cacert app.dev.oysterpack.com -tls-cert server.dev.oysterpack.com
//
// # run the following docker command from the pki root directory to create a new appdemo that will start the app RPCService with TLS
// docker run --name appdemo -p 44228:44228 -d -v `pwd`:/pki oysterpack/appdemo:0.1 -log-level DEBUG -app-id 0xd8e06a3d73a1d426
// -release-id 2 -tags A,B,C -app-rpc-port 44228 -pki-root /pki -tls-cacert app.dev.oysterpack.com -tls-cert server.dev.oysterpack.com
//
// an even better way is to leverage docker volumes :
//
//	docker volume create pki
//
//  # then get volume location using
//	docker volume inspect pki
//
// 	docker run --name appdemo -p 44228:44228 -d --mount source=pki,target=/pki oysterpack/appdemo:0.1 -log-level DEBUG \
// 		   -app-id 0xd8e06a3d73a1d426 -release-id 2 -tags A,B,C -app-rpc-port 44228 \
// 		   -pki-root /pki -tls-cacert app.dev.oysterpack.com -tls-cert server.dev.oysterpack.com
func main() {
	<-app.Dead()
}
