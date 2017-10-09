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

package server_test

import (
	"fmt"

	"net/url"
	"testing"
	"time"

	natsserver "github.com/nats-io/gnatsd/server"
	"github.com/nats-io/go-nats"
	"github.com/oysterpack/oysterpack.go/pkg/messaging/nats/server"
	"github.com/rs/zerolog/log"
)

func TestTLS(t *testing.T) {
	servers := map[string]*natsserver.Server{
		"server1": RunServer(Options(server.DEFAULT_SERVER_PORT, server.DEFAULT_CLUSTER_PORT, NATS_SEED_SERVER_URL)),
		"server2": RunServer(Options(server.DEFAULT_SERVER_PORT+1, server.DEFAULT_CLUSTER_PORT+1, NATS_SEED_SERVER_URL)),
	}

	defer func() {
		for _, server := range servers {
			server.Shutdown()
		}
	}()

	logServerInfo(servers, "Servers are running")

	const TOPIC = "TestTLS"

	pubConn, err := nats.Connect(fmt.Sprintf("tls://localhost:%d", server.DEFAULT_SERVER_PORT), nats.Secure(clientTLSConfig()))
	if err != nil {
		t.Fatalf("Failed to connect : %v", err)
	}

	subConn, err := nats.Connect(fmt.Sprintf("tls://localhost:%d", server.DEFAULT_SERVER_PORT+1), nats.Secure(clientTLSConfig()))
	if err != nil {
		t.Fatalf("Failed to connect : %v", err)
	}

	logServerInfo(servers, "Created 2 connections - one to each natsserver")

	ch := make(chan *nats.Msg, 20)
	subOnServer2, err := subConn.Subscribe(TOPIC, func(msg *nats.Msg) {
		t.Logf("Recievied message in handler on %s - forwarding to channel : %v", "subOnServer2", string(msg.Data))
		ch <- msg
	})
	if err != nil {
		t.Fatalf("Failed to subscribe : %v", err)
	}
	subOnServer1, err := pubConn.Subscribe(TOPIC, func(msg *nats.Msg) {
		t.Logf("Received message in handler on %s - forwarding to channel : %v", "subOnServer1", string(msg.Data))
		ch <- msg
	})
	if err != nil {
		t.Fatalf("Failed to subscribe : %v", err)
	}

	logServerInfo(servers, "Created a subscriber on server2")
	// if the subscriber connection is not flushed, then the publishing server1 is not immediately aware of the subscription
	// depending on the timing, the subscriber on server2 will miss messages until server1 becomes aware of the subscription on server2
	subConn.Flush()
	logServerInfo(servers, "Flushed the subscriber connection to natsserver 2")

	for i := 0; i < 10; i++ {
		pubConn.Publish(TOPIC, []byte(fmt.Sprintf("MSG #%d", i)))
	}
	logServerInfo(servers, "Published 10 messages on natsserver 1 connection")
	if count := receiveMessages(t, ch); count != 20 {
		t.Errorf("*** ERROR *** Received less than expected : %d", count)
	}

	logServerInfo(servers, "After receiving all messages")
	subOnServer2.Unsubscribe()
	logServerInfo(servers, "Unsubscribed subOnServer2")
	subConn.Flush()
	logServerInfo(servers, "Flushed publisher connection")

	for i := 10; i < 20; i++ {
		pubConn.Publish(TOPIC, []byte(fmt.Sprintf("MSG #%d", i)))
	}
	logServerInfo(servers, "Published 10 more messages")
	if count := receiveMessages(t, ch); count != 10 {
		t.Errorf("*** ERROR *** Should have received 10 messages on 1 subscriber : %d", count)
	}
	logServerInfo(servers, "After trying to receive for more messages")

	subOnServer1.Unsubscribe()
	logServerInfo(servers, "Unsubscribed subOnServer1")

	for i := 20; i < 30; i++ {
		pubConn.Publish(TOPIC, []byte(fmt.Sprintf("MSG #%d", i)))
	}
	logServerInfo(servers, "Published 10 more messages")
	if count := receiveMessages(t, ch); count != 0 {
		t.Errorf("*** ERROR *** No messages should have been received because the subscriber was unsuscribed : %d", count)
	}
	logServerInfo(servers, "After trying to receive for more messages")
}

func receiveMessages(t *testing.T, ch chan *nats.Msg) (count int) {
	t.Helper()

	retries := 0
LOOP:
	for {
		select {
		case msg := <-ch:
			count++
			t.Logf("Received msg : %v", string(msg.Data))
		default:
			if retries < 3 {
				retries++
				log.Info().Msgf("No messages on try #%d", retries)
				time.Sleep(10 * time.Millisecond)
			} else {
				break LOOP
			}
		}
	}
	return
}

func Options(port, clusterPort int, seedURL string) *natsserver.Options {
	opts := &natsserver.Options{
		Port:           port,
		NoLog:          true,
		NoSigs:         true,
		MaxControlLine: 256,
		TLSConfig:      serverTLSConfig(),
	}

	// configure clustering
	route, err := url.Parse(seedURL)
	if err != nil {
		panic(err)
	}
	opts.Routes = []*url.URL{route}
	opts.Cluster = natsserver.ClusterOpts{
		Port:      clusterPort,
		TLSConfig: opts.TLSConfig,
	}

	return opts
}

func DefaultTestOptions() *natsserver.Options {
	return Options(server.DEFAULT_SERVER_PORT, server.DEFAULT_CLUSTER_PORT, NATS_SEED_SERVER_URL)
}

func RunServer(opts *natsserver.Options) *natsserver.Server {
	if opts == nil {
		opts = DefaultTestOptions()
	}
	s := natsserver.New(opts)
	if s == nil {
		panic("No NATS Server object returned.")
	}

	go s.Start()

	if !s.ReadyForConnections(10 * time.Second) {
		panic("Unable to start NATS Server in Go Routine")
	}
	return s
}
