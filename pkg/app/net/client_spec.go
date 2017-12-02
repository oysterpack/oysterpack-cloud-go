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

package net

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net"

	"errors"

	"github.com/oysterpack/oysterpack.go/pkg/app"
	"github.com/oysterpack/oysterpack.go/pkg/app/net/config"
)

func NewClientSpec(spec config.ClientSpec) (*ClientSpec, error) {
	if err := CheckClientSpec(spec); err != nil {
		return nil, err
	}
	serviceSpec, err := spec.ServiceSpec()
	if err != nil {
		return nil, err
	}
	serverServiceSpec, err := NewServerServiceSpec(serviceSpec)
	if err != nil {
		return nil, err
	}
	clientSpec := &ClientSpec{
		ServerServiceSpec: serverServiceSpec,
		RootCAs:           x509.NewCertPool(),
	}

	clientCert, err := spec.ClientCert()
	if err != nil {
		return nil, err
	}
	cert, err := clientCert.Cert()
	if err != nil {
		return nil, err
	}
	key, err := clientCert.Key()
	if err != nil {
		return nil, err
	}
	clientSpec.Cert, err = tls.X509KeyPair(cert, key)
	if err != nil {
		return nil, err
	}

	caCert, err := spec.CaCert()
	if err != nil {
		return nil, err
	}
	if !clientSpec.RootCAs.AppendCertsFromPEM(caCert) {
		return nil, ErrPEMParsing
	}

	return clientSpec, nil
}

func CheckClientSpec(spec config.ClientSpec) error {
	if !spec.HasClientCert() {
		return NewClientSpecError(errors.New("Client cert is required"))
	}
	serverCert, err := spec.ClientCert()
	if err != nil {
		return err
	}
	if !serverCert.HasCert() {
		return NewClientSpecError(errors.New("Client cert is required"))
	}
	if !serverCert.HasKey() {
		return NewClientSpecError(errors.New("Client key is required"))
	}

	if !spec.HasCaCert() {
		return NewClientSpecError(errors.New("CA cert is required"))
	}

	return nil
}

// RPCClientSpec is the client spec for the RPCService
type ClientSpec struct {
	*ServerServiceSpec
	RootCAs *x509.CertPool
	Cert    tls.Certificate
}

func (a *ClientSpec) TLSConfig() *tls.Config {
	tlsConfig := &tls.Config{
		MinVersion: tls.VersionTLS12,

		// Ensure that we only use our "CA" to validate certificates
		RootCAs: a.RootCAs,
		// Server cert
		Certificates: []tls.Certificate{a.Cert},
		ServerName:   a.CN(),
	}
	tlsConfig.BuildNameToCertificate()
	return tlsConfig
}

// Conn returns an RPC conn using the service's standard network address - see NetworkAddr()
func (a *ClientSpec) Conn() (net.Conn, error) {
	networkAddr := a.NetworkAddr()
	app.Logger().Debug().
		Uint64("service", uint64(a.ServiceID)).
		Str("NetworkAddr", networkAddr).
		Msg("RPCClientSpec")
	addr := fmt.Sprintf("%s:%d", networkAddr, a.ServerPort)
	return tls.Dial("tcp", addr, a.TLSConfig())
}

// ConnForAddr returns an RPC conn using the specified network address
// This mainly intended for testing purposes to connect locally
func (a *ClientSpec) ConnForAddr(networkAddr string) (net.Conn, error) {
	addr := fmt.Sprintf("%s:%d", networkAddr, a.ServerPort)
	return tls.Dial("tcp", addr, a.TLSConfig())
}
