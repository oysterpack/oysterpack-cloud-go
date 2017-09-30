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

package nats_test

import (
	"testing"

	"github.com/oysterpack/oysterpack.go/pkg/messaging/nats"
	"github.com/oysterpack/oysterpack.go/pkg/messaging/natstest"
	"github.com/oysterpack/oysterpack.go/pkg/service"
)

func TestNewConnManagerClient(t *testing.T) {
	connManager := service.App().ServiceByType(nats.ConnManagerInterface).(nats.ConnManager)
	defer connManager.CloseAll()

	_, err := connManager.Connect()
	if err == nil {
		t.Errorf("Connect should have failed because no server is nats running")
	} else {
		t.Logf("Connect failed because server is not running")
	}

	server := natstest.RunServer()
	defer server.Shutdown()

	conn, err := connManager.Connect()
	if err != nil {
		t.Errorf("Server is running. We should have connected : %v", err)
	}

	if !conn.IsConnected() {
		t.Errorf("we should be connected")
	}

	t.Logf("%v", conn.ConnInfo())
}
