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
	"github.com/oysterpack/oysterpack.go/pkg/metrics"
)

func TestNewConnectionManager(t *testing.T) {
	metrics.ResetRegistry()
	server := natstest.RunServer()
	defer server.Shutdown()

	connMgr := nats.NewConnManager()
	defer connMgr.CloseAll()
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
	metrics.ResetRegistry()
	server := natstest.RunServer()
	defer server.Shutdown()

	connMgr := nats.NewConnManager()
	defer connMgr.CloseAll()
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
	metrics.ResetRegistry()
	server := natstest.RunServer()
	defer server.Shutdown()

	connMgr := nats.NewConnManager()
	defer connMgr.CloseAll()
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

func TestManagedConn_CloseConn(t *testing.T) {
	metrics.ResetRegistry()
	server := natstest.RunServer()
	defer server.Shutdown()

	connMgr := nats.NewConnManager()
	defer connMgr.CloseAll()
	conns := []*nats.ManagedConn{}

	// create some connections
	const COUNT = 5
	for i := 0; i < COUNT; i++ {
		conns = append(conns, mustConnect(t, connMgr))
	}

	// make sure they are registered with the ConnManager
	if connMgr.ConnCount() != COUNT {
		t.Errorf("There should be %d conns, but the ConnManager reported : %d", COUNT, connMgr.ConnCount())
	}
	if len(connMgr.ConnInfos()) != COUNT {
		t.Errorf("The number of ConnInfo(s) returned did not match the expected count : %d != %d", len(connMgr.ConnInfos()), COUNT)
	}

	// Close a connectio
	conn := conns[0]
	conn.Close()
	// the connection is removed async via the connection closed handler - so let's give it some time
	for i := 0; connMgr.ConnCount() != COUNT-1 && i < 3; i++ {
		time.Sleep(10 * time.Millisecond)
	}
	// verify that ConnManager has removed the closed conn
	if connMgr.ConnCount() != COUNT-1 {
		t.Errorf("There should be %d conns, but the ConnManager reported : %d", COUNT-1, connMgr.ConnCount())
	}
}

func TestNewConnManager_CreatedTimestamp(t *testing.T) {
	metrics.ResetRegistry()
	server := natstest.RunServer()
	defer server.Shutdown()

	connMgr := nats.NewConnManager()
	defer connMgr.CloseAll()

	now := time.Now()
	conn := mustConnect(t, connMgr)
	if !conn.Created().After(now) {
		t.Errorf("Created (%v) should be after (%v)", conn.Created(), now)
	}
}

func TestManagedConn_DisconnectReconnect(t *testing.T) {
	metrics.ResetRegistry()
	backup := nats.DefaultReConnectTimeout
	const ReConnectTimeout = 10 * time.Millisecond
	nats.DefaultReConnectTimeout = natsio.ReconnectWait(ReConnectTimeout)
	defer func() { nats.DefaultReConnectTimeout = backup }()

	server := natstest.RunServer()
	defer server.Shutdown()

	connMgr := nats.NewConnManager()
	defer connMgr.CloseAll()

	// create some connection
	conns := []*nats.ManagedConn{}
	const COUNT = 5
	for i := 0; i < COUNT; i++ {
		conns = append(conns, mustConnect(t, connMgr))
	}
	// make sure they all connected
	if count, total := connMgr.ConnectedCount(); count != 5 {
		t.Fatalf("Connected count is less than expected : connected = %d, total = %d", count, total)
	}

	server.Shutdown()

	// wait for all of the connections to report as disconnected
	for i := 0; i < 3; i++ {
		if count, _ := connMgr.ConnectedCount(); count == 0 {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
	// check that none are connected
	if count, total := connMgr.ConnectedCount(); count != 0 {
		t.Errorf("All connections should be disconnected because the server is down : connected = %d, total = %d", count, total)
	}
	// check that all are disconnected
	if count, total := connMgr.DisconnectedCount(); count != len(conns) {
		t.Errorf("All connections should be disconnected because the server is down : connected = %d, total = %d", count, total)
	}

	server = natstest.RunServer()
	defer server.Shutdown()

	// create a new connection the server - verifying that we can connect to the new server
	conn := mustConnect(t, connMgr)
	t.Logf("new connection after server restarted: %v", conn)

	// wait until the reconnect wait period expires
	time.Sleep(ReConnectTimeout)
	// check that all connections are connected
	if count, _ := connMgr.DisconnectedCount(); count != 0 {
		t.Errorf("All connections should be reconnected")
	}
	if count, _ := connMgr.ConnectedCount(); count != len(connMgr.ConnInfos()) {
		t.Errorf("All connections should be reconnected")
	}
	// check that LastDisconnectTime and LastReconnectTime were updated
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

func TestManagedConn_SubscribingWhileDisconnected(t *testing.T) {
	metrics.ResetRegistry()
	backup := nats.DefaultReConnectTimeout
	const ReConnectTimeout = 5 * time.Millisecond
	nats.DefaultReConnectTimeout = natsio.ReconnectWait(ReConnectTimeout)
	defer func() { nats.DefaultReConnectTimeout = backup }()

	server := natstest.RunServer()
	defer server.Shutdown()

	connMgr := nats.NewConnManager()
	defer connMgr.CloseAll()
	subConn := mustConnect(t, connMgr)
	pubConn := mustConnect(t, connMgr)

	const SUBJECT = "TestManagedConn_Errors"
	sub, err := subConn.SubscribeSync(SUBJECT)
	if err != nil {
		t.Fatalf("Failed to create subscription")
	}

	server.Shutdown()
	time.Sleep(ReConnectTimeout)
	if pubConn.IsConnected() || subConn.IsConnected() {
		t.Fatal("conns should not be connected because the server is shutdown")
	}
	if _, err := sub.NextMsg(1 * time.Millisecond); err == nil {
		t.Error("Expected error because the server was shutdown")
	} else {
		t.Logf("server was shutdown : sub.NextMsg() err : %v", err)
		t.Logf("after failing to receive next msg on disconnected pubConn :\n%v\n%v", pubConn, subConn)
	}

	// expecting messages to be dropped on unbuffered channel subscription after reconnecting
	unbufferedChan := make(chan *natsio.Msg)
	unbufferedChanSub, err := subConn.ChanSubscribe(SUBJECT, unbufferedChan)
	if err == nil {
		t.Logf("Subscription was created on disconnected conn")
	} else {
		t.Errorf("Failed to subscribe on disconnected conn : %v", err)
	}

	// expecting messages to be delivered on buffered channel subscription after reconnecting
	bufferedChan := make(chan *natsio.Msg, 10)
	bufferedChanSub, err := subConn.ChanSubscribe(SUBJECT, bufferedChan)
	if !bufferedChanSub.IsValid() {
		t.Errorf("subcription should be valid as long as the connection is not closed")
	}
	if err != nil {
		t.Errorf("There should not have been an err : %v", err)
	}

	logSubcriptionInfo(t, "unbufferedChanSub", unbufferedChanSub)
	logSubcriptionInfo(t, "bufferedChanSub", bufferedChanSub)

	select {
	case <-unbufferedChan:
		t.Error("no msg should have been received")
	default:
		t.Log("as expected, no message was received while disconnected")
	}

	if err := pubConn.Publish(SUBJECT, []byte("TEST MSG")); err != nil {
		t.Errorf("Did not expect an error because the conn will buffer messages whil disconnected : %v", err)
	}

	server = natstest.RunServer()
	defer server.Shutdown()
	time.Sleep(ReConnectTimeout)

	if !pubConn.IsConnected() || !subConn.IsConnected() {
		t.Fatalf("Expected conns to be reconnected")
	} else {
		t.Logf("after restarting the server :\npub: %v\nsub: %v", pubConn, subConn)
	}

	if msg, err := sub.NextMsg(5 * time.Millisecond); err == nil {
		t.Logf("msg was sent and received after reconnecting: %v", string(msg.Data))
	} else {
		t.Errorf("no msg was received after reconnected: %v", err)
	}

	select {
	case msg := <-unbufferedChan:
		t.Logf("msg was received on channel after reconnected : %v", string(msg.Data))
		t.Errorf("Expected msg to be dropped because the message was not received on the channel in a timely manner")
	default:
		t.Logf("As expected msg was dropped because the message was not received on the channel in a timely manner")
	}

	select {
	case msg := <-bufferedChan:
		t.Logf("msg was received on buffered channel after reconnected : %v", string(msg.Data))
	default:
		t.Errorf("No msg was received on buffered channel")
	}

	t.Logf("pub :%v\nsub : %v", pubConn, subConn)

	if pubConn.LastError() != nil {
		t.Errorf("Expected no errors to have recorded")
	}

	if subConn.LastError() != nil {
		t.Logf("subscriber conn error : %v", subConn.LastError())
	}

}

// Observed behavior :
// When using a channel based subscription, messages will be dropped if they cannot be sent on the channel, i.e., if the
// subscriber is not ready to receive the message on the channel, then the message is dropped.
func TestManagedConn_UnbufferedChanSubscribingWhileDisconnected(t *testing.T) {
	metrics.ResetRegistry()
	backup := nats.DefaultReConnectTimeout
	const ReConnectTimeout = 5 * time.Millisecond
	nats.DefaultReConnectTimeout = natsio.ReconnectWait(ReConnectTimeout)
	defer func() { nats.DefaultReConnectTimeout = backup }()

	server := natstest.RunServer()
	defer server.Shutdown()

	connMgr := nats.NewConnManager()
	defer connMgr.CloseAll()

	// create separate publisher and subscriber connections
	subConn := mustConnect(t, connMgr)
	pubConn := mustConnect(t, connMgr)

	const SUBJECT = "TestManagedConn_UnbufferedChanSubscribingWhileDisconnected"

	server.Shutdown()

	time.Sleep(ReConnectTimeout)
	if pubConn.IsConnected() || subConn.IsConnected() {
		t.Fatal("conns should not be connected because the server is shutdown")
	}

	// subscribe using an unbuffered channel
	msgChan := make(chan *natsio.Msg)
	chanSub, err := subConn.ChanSubscribe(SUBJECT, msgChan)
	if err == nil {
		t.Logf("Subscription was created on disconnected conn")
	} else {
		t.Fatalf("Failed to subscribe on disconnected conn : %v", err)
	}

	logSubcriptionInfo(t, "chanSub", chanSub)

	// publish message while disconnected
	if err := pubConn.Publish(SUBJECT, []byte("TEST MSG")); err != nil {
		t.Errorf("Did not expect an error because the conn will buffer messages while disconnected : %v", err)
	}

	t.Logf("publisher conn stats after publishing the message while disconnected : %v", pubConn.ConnInfo())

	// try receiving a message while disconnected should not cause any errors - simply no messages to receive
	select {
	case <-msgChan:
		t.Error("no msg should have been received")
	default:
		t.Log("as expected, no message was received while disconnected")
	}

	server = natstest.RunServer()
	time.Sleep(ReConnectTimeout)
	if !pubConn.IsConnected() || !subConn.IsConnected() {
		t.Fatalf("Expected conns to be reconnected")
	} else {
		t.Logf("after restarting the server :\npub: %v\nsub: %v", pubConn, subConn)
	}

	// flush messages that may be sitting on the publisher connection
	pubConn.Flush()

	select {
	case msg := <-msgChan:
		t.Logf("msg was received on channel after reconnected : %v", string(msg.Data))
		t.Error("Expected msg to be dropped because the message was not received on the channel in a timely manner")
	default:
		t.Log("As expected msg was dropped because the subscriber was not ready to receive the message when it was delivered - simulating slow consumer")
		t.Log("The message was dropped because the channel is not buffered. There was no one ready to receive the message when it was delivered.")
	}

	t.Logf("pubConn : %v", pubConn)
	t.Logf("subConn : %v", subConn)

	if pubConn.LastError() != nil {
		t.Errorf("Unexpected error : %v", pubConn.LastError())
	}

	if subConn.LastError() != nil {
		t.Logf("subscriber conn error : %v", subConn.LastError())
	} else {
		t.Errorf("Expected error due to message slow consumer")
	}

	connMgr.CloseAll()
	server.Shutdown()
}

func TestManagedConn_BufferedChanSubscribingWhileDisconnected(t *testing.T) {
	metrics.ResetRegistry()
	backup := nats.DefaultReConnectTimeout
	const ReConnectTimeout = 5 * time.Millisecond
	nats.DefaultReConnectTimeout = natsio.ReconnectWait(ReConnectTimeout)
	defer func() { nats.DefaultReConnectTimeout = backup }()

	server := natstest.RunServer()
	defer server.Shutdown()

	connMgr := nats.NewConnManager()
	defer connMgr.CloseAll()

	// create separate publisher and subscriber connections
	subConn := mustConnect(t, connMgr)
	pubConn := mustConnect(t, connMgr)

	const SUBJECT = "TestManagedConn_BufferedChanSubscribingWhileDisconnected"

	server.Shutdown()

	time.Sleep(ReConnectTimeout)
	if pubConn.IsConnected() || subConn.IsConnected() {
		t.Fatal("conns should not be connected because the server is shutdown")
	}

	// subscribe using a buffered channel
	msgChan := make(chan *natsio.Msg, 10)
	chanSub, err := subConn.ChanSubscribe(SUBJECT, msgChan)
	if !chanSub.IsValid() {
		t.Errorf("subcription should be valid as long as the connection is not closed")
	}
	if err != nil {
		t.Errorf("There should not have been an err : %v", err)
	}

	// publish message while disconnected
	if err := pubConn.Publish(SUBJECT, []byte("TEST MSG")); err != nil {
		t.Errorf("Did not expect an error because the conn will buffer messages whil disconnected : %v", err)
	}

	select {
	case <-msgChan:
		t.Error("no msg should have been received")
	default:
		t.Log("as expected, no message was received while disconnected")
	}

	server = natstest.RunServer()
	time.Sleep(ReConnectTimeout)
	// ensure the connetions are reconnected
	if !pubConn.IsConnected() || !subConn.IsConnected() {
		t.Fatalf("Expected conns to be reconnected")
	} else {
		t.Logf("after restarting the server :\npub: %v\nsub: %v", pubConn, subConn)
	}

	pubConn.Flush()
	subConn.Flush()

	select {
	case msg := <-msgChan:
		t.Logf("msg was received on buffered channel after reconnected : %v", string(msg.Data))
	default:
		t.Errorf("No msg was received on buffered channel")
	}

	t.Logf("pubConn : %v", pubConn)
	t.Logf("subConn : %v", subConn)

	if pubConn.LastError() != nil {
		t.Errorf("Unexpected error : %v", pubConn.LastError())
	}

	if subConn.LastError() != nil {
		t.Errorf("Unexpected error : %v", subConn.LastError())
		t.Logf("subscriber conn error : %v", subConn.LastError())
	}

	connMgr.CloseAll()
	server.Shutdown()
}

// When using an async subscriber, NATS will buffer pending messages based on the subscription's pending limits.
// Thus messages won't be dropped until the pending limits have been exceeded
func TestManagedConn_AsyncSubscribingWhileDisconnected(t *testing.T) {
	metrics.ResetRegistry()
	backup := nats.DefaultReConnectTimeout
	const ReConnectTimeout = 5 * time.Millisecond
	nats.DefaultReConnectTimeout = natsio.ReconnectWait(ReConnectTimeout)
	defer func() { nats.DefaultReConnectTimeout = backup }()

	server := natstest.RunServer()
	defer server.Shutdown()

	connMgr := nats.NewConnManager()
	defer connMgr.CloseAll()
	subConn := mustConnect(t, connMgr)
	pubConn := mustConnect(t, connMgr)

	const SUBJECT = "TestManagedConn_Errors"
	// the channel is used to pull messages from the message handler
	ch := make(chan *natsio.Msg)
	sub, err := subConn.Subscribe(SUBJECT, func(msg *natsio.Msg) {
		ch <- msg
	})
	if err != nil {
		t.Fatalf("Failed to create subscription")
	}
	logSubcriptionInfo(t, "async sub", sub)

	server.Shutdown()
	time.Sleep(ReConnectTimeout)
	if pubConn.IsConnected() || subConn.IsConnected() {
		t.Fatal("conns should not be connected because the server is shutdown")
	}

	// publish messages while disconnected
	const SEND_COUNT = 20
	for i := 1; i <= SEND_COUNT; i++ {
		if err := pubConn.Publish(SUBJECT, []byte(fmt.Sprintf("TEST MSG #%d", i))); err != nil {
			t.Fatalf("%d : Did not expect an error because the conn will buffer messages whil disconnected : %v", i, err)
		} else {
			t.Logf("published msg #%d", i)
		}
	}

	server = natstest.RunServer()
	defer server.Shutdown()
	time.Sleep(ReConnectTimeout)

	if !pubConn.IsConnected() || !subConn.IsConnected() {
		t.Fatalf("Expected conns to be reconnected")
	} else {
		t.Logf("after restarting the server :\npub: %v\nsub: %v", pubConn, subConn)
	}

	pubConn.Flush()

	t.Logf("pubConn : %v", pubConn)
	t.Logf("subConn : %v", subConn)
	logSubcriptionInfo(t, "async sub", sub)
	msgs, bytes, err := sub.Pending()
	if err != nil {
		t.Errorf("Error while retrieving pending info : %v", err)
	}
	if msgs != SEND_COUNT-1 {
		t.Logf("The message handler should be blocked on 1 message and the reset should be pending : %v, %v", msgs, bytes)
	}

	if pubConn.LastError() != nil {
		t.Errorf("Expected no errors")
	}
	if subConn.LastError() != nil {
		t.Errorf("Expected no errors")
	}

	receivedCount := 0
	for i := 0; i < SEND_COUNT; i++ {
		msg := <-ch
		t.Logf("%v", string(msg.Data))
		receivedCount++
	}

	t.Logf("receivedCount = %d", receivedCount)
	logSubcriptionInfo(t, "async sub", sub)
	msgs, bytes, err = sub.Pending()
	if err != nil {
		t.Errorf("Error while retrieving pending info : %v", err)
	}
	if msgs != 0 && bytes != 0 {
		t.Logf("There should be none pending : %v, %v", msgs, bytes)
	}

	connMgr.CloseAll()
	server.Shutdown()
}

// When using an async subscriber, NATS will buffer pending messages based on the subscription's pending limits.
// Thus messages won't be dropped until the pending limits have been exceeded
func TestManagedConn_AsyncSubscribingWhileDisconnected_WithPendingLimitsExceeded(t *testing.T) {
	metrics.ResetRegistry()
	backup := nats.DefaultReConnectTimeout
	const ReConnectTimeout = 5 * time.Millisecond
	nats.DefaultReConnectTimeout = natsio.ReconnectWait(ReConnectTimeout)
	defer func() { nats.DefaultReConnectTimeout = backup }()

	server := natstest.RunServer()
	defer server.Shutdown()

	connMgr := nats.NewConnManager()
	defer connMgr.CloseAll()
	subConn := mustConnect(t, connMgr)
	pubConn := mustConnect(t, connMgr)

	const SUBJECT = "TestManagedConn_Errors"
	// the channel is used to pull messages from the message handler
	ch := make(chan *natsio.Msg)
	sub, err := subConn.Subscribe(SUBJECT, func(msg *natsio.Msg) {
		ch <- msg
	})
	if err != nil {
		t.Fatalf("Failed to create subscription")
	}

	const SEND_COUNT = 20
	sub.SetPendingLimits(SEND_COUNT/2, 1024)
	logSubcriptionInfo(t, "async sub", sub)

	server.Shutdown()
	time.Sleep(ReConnectTimeout)
	if pubConn.IsConnected() || subConn.IsConnected() {
		t.Fatal("conns should not be connected because the server is shutdown")
	}

	// publish messages while disconnected
	for i := 1; i <= SEND_COUNT; i++ {
		if err := pubConn.Publish(SUBJECT, []byte(fmt.Sprintf("TEST MSG #%d", i))); err != nil {
			t.Fatalf("%d : Did not expect an error because the conn will buffer messages whil disconnected : %v", i, err)
		} else {
			t.Logf("published msg #%d", i)
		}
	}

	server = natstest.RunServer()
	defer server.Shutdown()
	time.Sleep(ReConnectTimeout)

	if !pubConn.IsConnected() || !subConn.IsConnected() {
		t.Fatalf("Expected conns to be reconnected")
	} else {
		t.Logf("after restarting the server :\npub: %v\nsub: %v", pubConn, subConn)
	}

	pubConn.Flush()

	t.Logf("pubConn : %v", pubConn)
	t.Logf("subConn : %v", subConn)
	logSubcriptionInfo(t, "async sub", sub)
	msgs, bytes, err := sub.Pending()
	if err != nil {
		t.Errorf("Error while retrieving pending info : %v", err)
	}
	if msgs != SEND_COUNT-1 {
		t.Logf("The message handler should be blocked on 1 message and the reset should be pending : %v, %v", msgs, bytes)
	}

	if pubConn.LastError() != nil {
		t.Errorf("Expected no errors")
	}
	if subConn.LastError() == nil {
		t.Errorf("Expected errors because subscriber should have been flagged as slow consumer")
	} else {
		t.Logf("%v : %v", subConn.LastError(), subConn.ConnInfo())
	}

	receivedCount := 0
	for i := 0; i < SEND_COUNT/2; i++ {
		msg := <-ch
		t.Logf("%v", string(msg.Data))
		receivedCount++
	}

	t.Logf("receivedCount = %d", receivedCount)
	logSubcriptionInfo(t, "async sub", sub)
	msgs, bytes, err = sub.Pending()
	if err != nil {
		t.Errorf("Error while retrieving pending info : %v", err)
	}
	if msgs != 0 && bytes != 0 {
		t.Logf("There should be none pending : %v, %v", msgs, bytes)
	}

	connMgr.CloseAll()
	server.Shutdown()
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

func logSubcriptionInfo(t *testing.T, name string, s *natsio.Subscription) {
	t.Helper()
	switch s.Type() {
	case natsio.ChanSubscription:
		// channel subscription limits are limited by the channel buffer size
		t.Logf("%s: %s : valid = %v", name, "ChanSubscription", s.IsValid())
	default:
		pendingMsgs, pendingBytes, err := s.Pending()
		if err != nil {
			t.Errorf("There should not have been an err : %v", err)
		}
		pendingMsgLimit, pendingByteLimit, err := s.PendingLimits()
		if err != nil {
			t.Errorf("There should not have been an err : %v", err)
		}
		t.Logf("%s : %s : valid = %v, pending : (%d, %d), limits: (%d, %d)", name, "SyncSubscription", s.IsValid(), pendingMsgs, pendingBytes, pendingMsgLimit, pendingByteLimit)
	}

}
