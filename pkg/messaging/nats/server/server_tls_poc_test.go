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
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"net/url"
	"testing"
	"time"

	"github.com/nats-io/gnatsd/server"
	"github.com/nats-io/go-nats"
	"github.com/rs/zerolog/log"
)

const (
	PKI_ROOT             = "./testdata/.easypki/pki"
	NATS_SEED_SERVER_URL = "tls://localhost:6443"
)

func TestTLS(t *testing.T) {
	servers := map[string]*server.Server{
		"server1": RunServer(Options(4443, 6443, NATS_SEED_SERVER_URL)),
		"server2": RunServer(Options(5443, 6444, NATS_SEED_SERVER_URL)),
	}

	defer func() {
		for _, server := range servers {
			server.Shutdown()
		}
	}()

	logServerInfo(servers, "Servers are running")

	const TOPIC = "TestTLS"

	pubConn, err := nats.Connect("tls://localhost:4443", nats.Secure(clientTLSConfig()))
	if err != nil {
		t.Fatalf("Failed to connect : %v", err)
	}
	//subConn, err := nats.Connect("tls://localhost:5443, tls://localhost:4443", nats.Secure(clientTLSConfig()))
	subConn, err := nats.Connect("tls://localhost:5443", nats.Secure(clientTLSConfig()))
	if err != nil {
		t.Fatalf("Failed to connect : %v", err)
	}

	logServerInfo(servers, "Created 2 connections - one to each server")

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
	logServerInfo(servers, "Flushed the subscriber connection to server 2")

	for i := 0; i < 10; i++ {
		pubConn.Publish(TOPIC, []byte(fmt.Sprintf("MSG #%d", i)))
	}
	logServerInfo(servers, "Published 10 messages on server 1 connection")
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

func logServerInfo(servers map[string]*server.Server, msg string) {
	for name, server := range servers {
		log.Info().
			Str("name", name).
			Str("Addr", fmt.Sprintf("%v", server.Addr())).
			Str("ClusterAddr", fmt.Sprintf("%v", server.ClusterAddr())).
			Str("NumRemotes", fmt.Sprintf("%v", server.NumRemotes())).
			Str("NumClients", fmt.Sprintf("%v", server.NumClients())).
			Str("NumSubscriptions", fmt.Sprintf("%v", server.NumSubscriptions())).
			Msg(msg)
	}
}

func clientTLSConfig() *tls.Config {
	certKeyPair, err := tls.LoadX509KeyPair(
		PKI_ROOT+"/oysterpack/certs/client.nats.dev.oysterpack.com.crt",
		PKI_ROOT+"/oysterpack/keys/client.nats.dev.oysterpack.com.key",
	)
	if err != nil {
		panic(err)
	}

	pool := x509.NewCertPool()
	caCert := PKI_ROOT + "/oysterpack/certs/oysterpack.crt"
	rootPEM, err := ioutil.ReadFile(caCert)
	if err != nil || rootPEM == nil {
		panic(err)
	}
	ok := pool.AppendCertsFromPEM([]byte(rootPEM))
	if !ok {
		panic("nats: failed to parse root certificate")
	}

	return &tls.Config{
		MinVersion:   tls.VersionTLS12,
		RootCAs:      pool,
		Certificates: []tls.Certificate{certKeyPair},
	}
}

func serverTLSConfig() *tls.Config {
	certKeyPair, err := tls.LoadX509KeyPair(
		PKI_ROOT+"/oysterpack/certs/nats.dev.oysterpack.com.crt",
		PKI_ROOT+"/oysterpack/keys/nats.dev.oysterpack.com.key",
	)
	if err != nil {
		panic(err)
	}

	pool := x509.NewCertPool()
	caCert := PKI_ROOT + "/oysterpack/certs/oysterpack.crt"
	rootPEM, err := ioutil.ReadFile(caCert)
	if err != nil || rootPEM == nil {
		panic(err)
	}
	ok := pool.AppendCertsFromPEM([]byte(rootPEM))
	if !ok {
		panic("nats: failed to parse root certificate")
	}

	return &tls.Config{
		MinVersion:   tls.VersionTLS12,
		RootCAs:      pool,
		ClientAuth:   tls.RequireAndVerifyClientCert,
		ClientCAs:    pool,
		Certificates: []tls.Certificate{certKeyPair},
	}
}

func Options(port, clusterPort int, seedURL string) *server.Options {
	opts := &server.Options{
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
	opts.Cluster = server.ClusterOpts{
		Port:      clusterPort,
		TLSConfig: opts.TLSConfig,
	}

	return opts
}

func DefaultTestOptions() *server.Options {
	return Options(4443, 6443, NATS_SEED_SERVER_URL)
}

func RunServer(opts *server.Options) *server.Server {
	if opts == nil {
		opts = DefaultTestOptions()
	}
	s := server.New(opts)
	if s == nil {
		panic("No NATS Server object returned.")
	}

	// Run server in Go routine.
	go s.Start()

	// Wait for accept loop(s) to be started
	if !s.ReadyForConnections(10 * time.Second) {
		panic("Unable to start NATS Server in Go Routine")
	}
	return s
}

func RunDefaultServer() *server.Server {
	return RunServer(DefaultTestOptions())
}
