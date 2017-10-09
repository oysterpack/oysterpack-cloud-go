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

	natsserver "github.com/nats-io/gnatsd/server"
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
		ClientCAs: pool,
		ClientAuth:   tls.RequireAndVerifyClientCert,
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
		go server.Start()
	}

	for _, server := range servers {
	SERVER:
		for {
			if !server.ReadyForConnections(time.Second * 2) {
				log.Logger.Info().Msgf("Waiting for server to startup ...")
				continue
			}
			break SERVER
		}
	}
}

func startMonitoring(servers []server.NATSServer) error {
	for _, server := range servers {
		if err := server.StartMonitoring(); err != nil {
			return err
		}
	}

	return nil
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
			Str("NumRemotes", fmt.Sprintf("%v", server.NumRemotes())).
			Str("NumClients", fmt.Sprintf("%v", server.NumClients())).
			Str("NumSubscriptions", fmt.Sprintf("%v", server.NumSubscriptions())).
			Msg(msg)
	}
}

func logNATServerInfo(servers []server.NATSServer, msg string) {
	for _, server := range servers {
		log.Info().
			Str("Addr", fmt.Sprintf("%v", server.Addr())).
			Str("ClusterAddr", fmt.Sprintf("%v", server.ClusterAddr())).
			Str("NumRemotes", fmt.Sprintf("%v", server.NumRemotes())).
			Str("NumClients", fmt.Sprintf("%v", server.NumClients())).
			Str("NumSubscriptions", fmt.Sprintf("%v", server.NumSubscriptions())).
			Msg(msg)
	}
}
