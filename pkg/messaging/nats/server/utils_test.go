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

	"time"

	"errors"
	"testing"

	natsserver "github.com/nats-io/gnatsd/server"
	"github.com/oysterpack/oysterpack.go/pkg/messaging"
	"github.com/oysterpack/oysterpack.go/pkg/messaging/nats/server"
	"github.com/rs/zerolog/log"
)

const (
	PKI_ROOT = "./testdata/.easypki/pki"
)

var (
	NATS_SEED_SERVER_URL = fmt.Sprintf("nats://localhost:%d", server.DEFAULT_CLUSTER_PORT)
)

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

func clusterTLSConfig() *tls.Config {
	certKeyPair, err := tls.LoadX509KeyPair(
		PKI_ROOT+"/oysterpack/certs/cluster.nats.dev.oysterpack.com.crt",
		PKI_ROOT+"/oysterpack/keys/cluster.nats.dev.oysterpack.com.key",
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
		//InsecureSkipVerify:true,
	}
}

func waitforClusterMeshToForm(servers []server.NATSServer) error {
	timeout := time.After(time.Second * 10)
	expectedRemoteCount := len(servers) - 1
	for {
		select {
		case <-timeout:
			totalRemoteCount := 0
			totalRouteCount := 0
			for _, server := range servers {
				totalRemoteCount += server.NumRemotes()
				totalRouteCount += server.NumRoutes()
			}
			if totalRemoteCount != expectedRemoteCount*len(servers) {
				return fmt.Errorf("Remotes are missing : %d / %d", totalRemoteCount, expectedRemoteCount*len(servers))
			}
			if totalRouteCount != expectedRemoteCount*len(servers) {
				return fmt.Errorf("Routes are missing : %d / %d", totalRouteCount, expectedRemoteCount*len(servers))
			}

			return nil
		default:
			for i, server := range servers {
				if count := server.NumRemotes(); count != expectedRemoteCount {
					log.Logger.Info().Int("remote count", count).Int("i", i).Msg("")
					time.Sleep(10 * time.Millisecond)
					continue
				}
				if count := server.NumRoutes(); count != expectedRemoteCount {
					log.Logger.Info().Int("route count", count).Int("i", i).Msg("")
					time.Sleep(10 * time.Millisecond)
					continue
				}
			}
			return nil
		}
	}
}

func defaultRouteURLsWithSeed(ports ...int) []*url.URL {
	seedRoute, err := url.Parse(NATS_SEED_SERVER_URL)
	if err != nil {
		panic(err)
	}
	routes := []*url.URL{seedRoute}
	for _, port := range ports {
		route, err := url.Parse(fmt.Sprintf("tls://localhost:%d", port))
		if err != nil {
			panic(err)
		}
		routes = append(routes, route)
	}
	return routes
}

func defaultRoutesWithSeed(ports ...int) []string {
	routes := []string{NATS_SEED_SERVER_URL}
	for _, port := range ports {
		routes = append(routes, fmt.Sprintf("tls://localhost:%d", port))
	}
	return routes
}

func startServers(servers []server.NATSServer) {
	for _, server := range servers {
		server.Start()
	}
}

func shutdownServers(servers []server.NATSServer) {
	for _, server := range servers {
		server.Shutdown()
	}
}

func logServerInfo(servers map[string]*natsserver.Server, msg string) {
	for name, server := range servers {
		log.Info().
			Str("name", name).
			Str("Addr", fmt.Sprintf("%v", server.Addr())).
			Str("ClusterAddr", fmt.Sprintf("%v", server.ClusterAddr())).
			Int("NumRoutes", server.NumRoutes()).
			Int("NumRemotes", server.NumRemotes()).
			Int("NumClients", server.NumClients()).
			Uint32("NumSubscriptions", server.NumSubscriptions()).
			Msg(msg)
	}
}

func logNATServerInfo(servers []server.NATSServer, msg string) {
	for _, server := range servers {
		log.Info().
			Str("Addr", fmt.Sprintf("%v", server.Addr())).
			Str("ClusterAddr", fmt.Sprintf("%v", server.ClusterAddr())).
			Int("NumRoutes", server.NumRoutes()).
			Int("NumRemotes", server.NumRemotes()).
			Int("NumClients", server.NumClients()).
			Uint32("NumSubscriptions", server.NumSubscriptions()).
			Msg(msg)
	}
}

// - returns the number of messages received
// - stops receiving when the number of messages received reaches messageCount
// - performs 10 tries, and sleeps for 1 msec between tries
func receiveMessagesOnQueueSubscriptions(qsubscriptions []messaging.QueueSubscription, messageCount int) int {
	msgReceivedCount := 0
	for i := 0; i < 10; i++ {
		for i, subscription := range qsubscriptions {
			select {
			case msg := <-subscription.Channel():
				msgReceivedCount++
				log.Logger.Info().Msgf("qsubscriptions[%d] #%d : %v", i, msgReceivedCount, string(msg.Data))
			default:
			}
		}

		if msgReceivedCount >= messageCount {
			return msgReceivedCount
		}
		log.Logger.Info().Msgf("receiveMessagesOnQueueSubscriptions: %d / %d", msgReceivedCount, messageCount)
		time.Sleep(time.Millisecond)
	}
	return msgReceivedCount
}

// timeout after 5 seconds
func waitForClusterToBecomeAwareOfAllSubscriptions(servers []server.NATSServer, subscriptionCount int) error {
	timeout := time.After(time.Second * 5)
	for {
		select {
		case <-timeout:
			for _, server := range servers {
				if int(server.NumSubscriptions()) != subscriptionCount {
					return errors.New("Timed out : waitForClusterToBecomeAwareOfAllSubscriptions()")
				}
			}
			log.Logger.Info().Msg("Entire cluster is aware of all subscriptions")
			return nil
		default:
			for _, server := range servers {
				if int(server.NumSubscriptions()) != subscriptionCount {
					log.Logger.Info().Msgf("Subscription count = %d", server.NumSubscriptions())
					time.Sleep(time.Millisecond)
					continue
				}
			}
			log.Logger.Info().Msg("Entire cluster is aware of all subscriptions")
			return nil
		}

	}
}

func createNATSServers(t *testing.T, configs []*server.NATSServerConfig) []server.NATSServer {
	var servers []server.NATSServer
	for _, config := range configs {
		server, err := server.NewNATSServer(config)
		if err != nil {
			t.Fatalf("server.NewNATSServer failed : %v", err)
		}
		servers = append(servers, server)
	}
	return servers
}

func checkClientConnectionCounts(servers []server.NATSServer, countPerServer int) error {
	for _, server := range servers {
		if server.NumClients() != countPerServer {
			return fmt.Errorf("The number of clients did not match : %d != %d", server.NumClients(), countPerServer)
		}
	}
	return nil
}

// - returns the number of messages received
// - stops receiving when the number of messages received reaches messageCount
// - performs 10 tries, and sleeps for 1 msec between tries
func receiveMessagesOnSubscriptions(subscriptions []messaging.Subscription, messageCount int) int {
	msgReceivedCount := 0
	for i, subscription := range subscriptions {
		msg := <-subscription.Channel()
		msgReceivedCount++
		log.Logger.Info().Msgf("subscriptions[%d] #%d : %v", i, msgReceivedCount, string(msg.Data))
	}

	for i := 0; i < 10; i++ {
		for i, subscription := range subscriptions {
			select {
			case msg := <-subscription.Channel():
				msgReceivedCount++
				log.Logger.Info().Msgf("subscriptions[%d] #%d : %v", i, msgReceivedCount, string(msg.Data))
			default:
			}
		}

		if msgReceivedCount >= messageCount {
			return msgReceivedCount
		}
		log.Logger.Info().Msgf("receiveMessagesOnQueueSubscriptions: %d / %d", msgReceivedCount, messageCount)
		time.Sleep(time.Millisecond)
	}
	return msgReceivedCount
}
