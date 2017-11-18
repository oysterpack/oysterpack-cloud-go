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

// this file is used to put shared testing util code

import (
	"context"
	"net"

	"github.com/oysterpack/oysterpack.go/pkg/app/capnprpc"
	"zombiezen.com/go/capnproto2/rpc"
)

func appClientConn(addr net.Addr) (capnprpc.App, net.Conn, error) {
	clientConn, err := net.Dial(addr.Network(), addr.String())
	if err != nil {
		return capnprpc.App{}, nil, err
	}
	rpcClient := rpc.NewConn(rpc.StreamTransport(clientConn))
	return capnprpc.App{Client: rpcClient.Bootstrap(context.Background())}, clientConn, nil
}
