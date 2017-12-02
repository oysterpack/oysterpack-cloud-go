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

	"github.com/oysterpack/oysterpack.go/pkg/app/net/config"
	"zombiezen.com/go/capnproto2"

	"github.com/oysterpack/oysterpack.go/pkg/app"
)

// NewServerSpec converts a config.ServiceSpec to a ServiceSpec
func NewServerServiceSpec(spec config.ServiceSpec) (*ServerServiceSpec, error) {
	serviceSpec := &ServerServiceSpec{
		app.DomainID(spec.DomainID()),
		app.AppID(spec.AppId()),
		app.ServiceID(spec.ServiceId()),
		ServerPort(spec.Port()),
	}
	if err := serviceSpec.Validate(); err != nil {
		return nil, err
	}
	return serviceSpec, nil
}

// ServerServiceSpec is the common Service spec shared by the  server and client
type ServerServiceSpec struct {
	app.DomainID
	app.AppID
	app.ServiceID

	ServerPort
}

// CN returns the x509 CN - this used by client TLS to set the x509.Config.ServerName
func (a *ServerServiceSpec) CN() string {
	return ServerCN(a.DomainID, a.AppID, a.ServiceID)
}

// NetworkAddr returns the service network address, which is used by the client to connect to the service.
// The network address uses the following naming convention :
//
//		fmt.Sprintf("%x_%x", a.DomainID, a.AppID)
//
//		e.g. ed5cf026e8734361-d113a2e016e12f0f
func (a *ServerServiceSpec) NetworkAddr() string {
	return fmt.Sprintf("%x_%x", a.DomainID, a.AppID)
}

func (a *ServerServiceSpec) ToCapnp(s *capnp.Segment) (config.ServiceSpec, error) {
	spec, err := config.NewServiceSpec(s)
	if err != nil {
		return spec, err
	}
	spec.SetDomainID(uint64(a.DomainID))
	spec.SetAppId(uint64(a.AppID))
	spec.SetServiceId(uint64(a.ServiceID))
	spec.SetPort(uint16(a.ServerPort))
	return spec, nil
}

func (a *ServerServiceSpec) Validate() error {
	if a.DomainID == app.DomainID(0) {
		return app.ErrDomainIDZero
	}
	if a.AppID == app.AppID(0) {
		return app.ErrAppIDZero
	}
	if a.ServiceID == app.ServiceID(0) {
		return app.ErrServiceIDZero
	}
	if a.ServerPort == ServerPort(0) {
		return ErrServerPortZero
	}
	return nil
}

// ServerPort represents a server network port
type ServerPort uint16

func CheckServerSpec(spec config.ServerSpec) error {
	if !spec.HasServerCert() {
		return NewServerSpecError(errors.New("ServerCert is required"))
	}
	serverCert, err := spec.ServerCert()
	if err != nil {
		return err
	}
	if !serverCert.HasCert() {
		return NewServerSpecError(errors.New("ServerCert.Cert is required"))
	}
	if !serverCert.HasKey() {
		return NewServerSpecError(errors.New("ServerCert.Key is required"))
	}

	if !spec.HasCaCert() {
		return NewServerSpecError(errors.New("CaCert is required"))
	}

	return nil
}

func NewServerSpec(spec config.ServerSpec) (*ServerSpec, error) {
	if err := CheckServerSpec(spec); err != nil {
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
	serverSpec := &ServerSpec{
		ServerServiceSpec: serverServiceSpec,
		ClientCAs:         x509.NewCertPool(),
		MaxConns:          spec.MaxConns(),
	}

	serverCert, err := spec.ServerCert()
	if err != nil {
		return nil, err
	}
	cert, err := serverCert.Cert()
	if err != nil {
		return nil, err
	}
	key, err := serverCert.Key()
	if err != nil {
		return nil, err
	}
	serverSpec.Cert, err = tls.X509KeyPair(cert, key)
	if err != nil {
		return nil, err
	}

	caCert, err := spec.CaCert()
	if err != nil {
		return nil, err
	}
	if !serverSpec.ClientCAs.AppendCertsFromPEM(caCert) {
		return nil, ErrPEMParsing
	}

	return serverSpec, nil
}

// ServerSpec is the server spec for the Service
type ServerSpec struct {
	*ServerServiceSpec
	ClientCAs *x509.CertPool
	Cert      tls.Certificate
	MaxConns  uint32
}

func (a *ServerSpec) ToCapnp(s *capnp.Segment) (config.ServerSpec, error) {
	serverSpec, err := config.NewRootServerSpec(s)
	if err != nil {
		return serverSpec, err
	}

	serviceSpec, err := a.ServerServiceSpec.ToCapnp(s)
	if err != nil {
		return serverSpec, err
	}
	serverSpec.SetServiceSpec(serviceSpec)

	serverSpec.SetMaxConns(a.MaxConns)
	// TODO: set server Cert and client CA Cert

	return serverSpec, nil
}

func (a *ServerSpec) TLSConfig() *tls.Config {
	tlsConfig := &tls.Config{
		MinVersion: tls.VersionTLS12,

		// Reject any TLS certificate that cannot be validated
		ClientAuth: tls.RequireAndVerifyClientCert,

		// Ensure that we only use our "CA" to validate certificates
		ClientCAs: a.ClientCAs,
		// Server cert
		Certificates: []tls.Certificate{a.Cert},

		PreferServerCipherSuites: true,

		// TODO: PFS because we can but this will reject client with RSA certificates
		//CipherSuites:             []uint16{tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384},
	}
	tlsConfig.BuildNameToCertificate()
	return tlsConfig
}

func (a *ServerSpec) TLSConfigProvider() func() (*tls.Config, error) {
	return func() (*tls.Config, error) {
		return a.TLSConfig(), nil
	}
}

func (a *ServerSpec) ListenerProvider() func() (net.Listener, error) {
	return func() (net.Listener, error) {
		return net.Listen("tcp", fmt.Sprintf(":%d", a.ServerPort))
	}
}
