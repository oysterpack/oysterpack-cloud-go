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

	"fmt"

	natsio "github.com/nats-io/go-nats"
	"github.com/oysterpack/oysterpack.go/pkg/messaging/nats"
	"github.com/oysterpack/oysterpack.go/pkg/messaging/natstest"
)

func TestNewConnectionManager(t *testing.T) {
	server := natstest.RunServer()
	defer server.Shutdown()

	connMgr := nats.NewConnManager()
	conn := mustConnect(t, connMgr)

	conn2 := connMgr.ManagedConn(conn.ID())
	if conn2 == nil {
		t.Error("No connection was returned")
	}
	if conn2.ID() != conn.ID() {
		t.Errorf("The wrong conn was returned : %v != %v", conn.ID(), conn2.ID())
	}

	if connMgr.ConnCount() != 1 {
		t.Errorf("Expected ConnCount() == 1, but was %d", connMgr.ConnCount())
	}

	if connMgr.ConnInfo(conn.ConnInfo().Id) == nil {
		t.Errorf("should have been found")
	}
}

func TestConnManager_ConnInfo(t *testing.T) {
	server := natstest.RunServer()
	defer server.Shutdown()

	connMgr := nats.NewConnManager()
	conn := mustConnect(t, connMgr, "a", "b", "c")
	conn2 := connMgr.ManagedConn(conn.ID())

	const (
		topic = "TestConnManager_ConnInfo"
		count = 10
	)
	ch := make(chan *natsio.Msg)
	conn.ChanSubscribe(topic, ch)

	for i := 0; i < count; i++ {
		conn.Publish(topic, []byte(fmt.Sprintf("%v", i)))
		msg := <-ch
		t.Logf("received msg %v", string(msg.Data))
	}

	connInfo := conn.ConnInfo()
	t.Logf("connInfo : %v", connInfo)
	if len(conn.Tags()) != 3 || connInfo.Tags[0] != "a" || connInfo.Tags[1] != "b" || connInfo.Tags[2] != "c" {
		t.Error("Tags are not matching")
	}
	if connInfo.InMsgs != count || connInfo.OutMsgs != count || connInfo.InBytes != count || connInfo.OutBytes != count {
		t.Error("stats are not lining up")
	}

	connInfo2 := conn2.ConnInfo()
	t.Logf("connInfo2 : %v", connInfo2)
	if connInfo.InBytes != connInfo2.InBytes {
		t.Error("cnnInfo2 did not match connInfo")
	}

}

func TestConnManager_CloseAll(t *testing.T) {
	server := natstest.RunServer()
	defer server.Shutdown()

	connMgr := nats.NewConnManager()
	conn := mustConnect(t, connMgr)

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

func TestManagedConn_ClosingConn(t *testing.T) {
	server := natstest.RunServer()
	defer server.Shutdown()

	connMgr := nats.NewConnManager()
	conns := []*nats.ManagedConn{}
	const COUNT = 5
	for i := 0; i < COUNT; i++ {
		conns = append(conns, mustConnect(t, connMgr))
	}

	if connMgr.ConnCount() != COUNT {
		t.Errorf("There should be %d conns, but the ConnManager reported : %d", COUNT, connMgr.ConnCount())
	}
	if len(connMgr.ConnInfos()) != COUNT {
		t.Errorf("The number of ConnInfo(s) returned did not match the expected count : %d != %d", len(connMgr.ConnInfos()), COUNT)
	}

	conn := conns[0]
	conn.Close()
	// the connection is removed async via the connection closed handler - so let's give it some time
	for i := 0; connMgr.ConnCount() != COUNT-1 && i < 3; i++ {
		time.Sleep(10 * time.Millisecond)
	}
	if connMgr.ConnCount() != COUNT-1 {
		t.Errorf("There should be %d conns, but the ConnManager reported : %d", COUNT-1, connMgr.ConnCount())
	}

	conns = conns[1:]
	server.Shutdown()

	for i := 0; i < 3; i++ {
		if count, _ := connMgr.ConnectedCount(); count == 0 {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
	if count, _ := connMgr.ConnectedCount(); count != 0 {
		t.Errorf("All connections should be disconnected because the server is down")
	}

}

func TestManagedConn_DisconnectReconnect(t *testing.T) {
	backup := nats.DefaultReConnectTimeout
	const ReConnectTimeout = 10 * time.Millisecond
	nats.DefaultReConnectTimeout = natsio.ReconnectWait(ReConnectTimeout)
	defer func() { nats.DefaultReConnectTimeout = backup }()

	server := natstest.RunServer()
	defer server.Shutdown()

	now := time.Now()

	connMgr := nats.NewConnManager()
	conns := []*nats.ManagedConn{}
	const COUNT = 5
	for i := 0; i < COUNT; i++ {
		conns = append(conns, mustConnect(t, connMgr))
	}

	if count, total := connMgr.ConnectedCount(); count != 5 {
		t.Fatalf("Connected count is less than expected : connected = %d, total = %d", count, total)
	}

	for _, c := range conns {
		t.Logf("after reconnect : %v", c)
		if !c.Created().After(now) {
			t.Errorf("Created (%v) should be after (%v)", c.Created(), now)
		}
	}

	server.Shutdown()

	for i := 0; i < 3; i++ {
		if count, _ := connMgr.ConnectedCount(); count == 0 {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
	if count, total := connMgr.ConnectedCount(); count != 0 {
		t.Errorf("All connections should be disconnected because the server is down : connected = %d, total = %d", count, total)
	}
	if count, total := connMgr.DisconnectedCount(); count != len(conns) {
		t.Errorf("All connections should be disconnected because the server is down : connected = %d, total = %d", count, total)
	}

	server = natstest.RunServer()
	conn := mustConnect(t, connMgr)
	t.Logf("new connection after server restarted: %v", conn)

	time.Sleep(ReConnectTimeout)
	if count, _ := connMgr.DisconnectedCount(); count != 0 {
		t.Errorf("All connections should be reconnected")
	}
	for _, c := range conns {
		t.Logf("after reconnect : %v", c)
		if !c.LastDisconnectTime().After(c.Created()) {
			t.Errorf("LastDisconnectTime (%v) should be after Created (%v)", c.LastDisconnectTime(), c.Created())
		}
		if !c.LastReconnectTime().After(c.LastDisconnectTime()) {
			t.Errorf("LastReconnectTime (%v) should be after LastDisconnectTime (%v)", c.LastReconnectTime(), c.LastDisconnectTime())
		}
	}
}

func mustConnect(t *testing.T, connMgr nats.ConnManager, tags ...string) *nats.ManagedConn {
	t.Helper()
	conn, err := connMgr.Connect(tags...)
	if err != nil {
		t.Fatalf("Connect() failed : %v", err)
	}
	t.Logf("conn : %v", conn)
	if !conn.IsConnected() {
		t.Fatalf("should be connected")
	}
	return conn
}
