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

	"time"

	"github.com/oysterpack/oysterpack.go/pkg/messaging/nats"
	"github.com/oysterpack/oysterpack.go/pkg/messaging/natstest"
)

func TestNewConnectionManager(t *testing.T) {
	server := natstest.RunServer()
	defer server.Shutdown()

	connMgr := nats.NewConnectionManager()
	conn, err := connMgr.Connect()
	if err != nil {
		t.Errorf("Connect() failed : %v", err)
	} else {
		t.Logf("conn : %v", conn)
	}

	if !conn.IsConnected() {
		t.Errorf("should be connected")
	}

	if connMgr.ConnCount() != 1 {
		t.Errorf("Expected ConnCount() == 1, but was %d", connMgr.ConnCount())
	}

	if connMgr.ConnInfo(conn.ConnInfo().Id) == nil {
		t.Errorf("should have been found")
	}

}

func TestConnManager_CloseAll(t *testing.T) {
	server := natstest.RunServer()
	defer server.Shutdown()

	connMgr := nats.NewConnectionManager()
	conn, err := connMgr.Connect()
	if err != nil {
		t.Fatal("should have connected")
	}

	connMgr.CloseAll()

	if conn.IsConnected() {
		t.Errorf("should be closed")
	}

	for i := 0; conn.Disconnects() != 1 && i < 3; i++ {
		time.Sleep(10 * time.Millisecond)
	}
	t.Logf("after the connection is closed : %v", conn)
	if conn.Disconnects() != 1 {
		t.Errorf("The disonnect handler should have run by now")
	}

	if connMgr.ConnCount() != 0 {
		t.Errorf("There should be no conns")
	}

	if connMgr.ConnInfo(conn.ConnInfo().Id) != nil {
		t.Errorf("should have been removed")
	}
}
