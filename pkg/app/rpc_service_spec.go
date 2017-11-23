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

package app

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"

	"net"

	"errors"

	"github.com/oysterpack/oysterpack.go/pkg/app/config"
	"zombiezen.com/go/capnproto2"
	"zombiezen.com/go/capnproto2/rpc"
)

// NewRPCServiceSpec converts a config.RPCServiceSpec to a RPCServiceSpec
func NewRPCServiceSpec(spec config.RPCServiceSpec) (*RPCServiceSpec, error) {
	serviceSpec := &RPCServiceSpec{
		DomainID(spec.DomainID()),
		AppID(spec.AppId()),
		ServiceID(spec.ServiceId()),
		RPCPort(spec.Port()),
	}
	if err := serviceSpec.Validate(); err != nil {
		return nil, err
	}
	return serviceSpec, nil
}

// RPCServiceSpec is the common RPCService spec shared by the RPC server and client
type RPCServiceSpec struct {
	DomainID
	AppID
	ServiceID

	RPCPort
}

// ServerCN returns the CN for the service server
func ServerCN(domain DomainID, app AppID, service ServiceID) string {
	return fmt.Sprintf("%x.%x.%x", service, app, domain)
}

// CN returns the x509 CN - this used by client TLS to set the x509.Config.ServerName
func (a *RPCServiceSpec) CN() string {
	return ServerCN(a.DomainID, a.AppID, a.ServiceID)
}

// NetworkAddr returns the service network address, which is used by the client to connect to the service.
// The network address uses the following naming convention :
//
//		fmt.Sprintf("%x_%x", a.DomainID, a.AppID)
//
//		e.g. ed5cf026e8734361-d113a2e016e12f0f
func (a *RPCServiceSpec) NetworkAddr() string {
	return fmt.Sprintf("%x_%x", a.DomainID, a.AppID)
}

func (a *RPCServiceSpec) ToCapnp(s *capnp.Segment) (config.RPCServiceSpec, error) {
	spec, err := config.NewRPCServiceSpec(s)
	if err != nil {
		return spec, err
	}
	spec.SetDomainID(uint64(a.DomainID))
	spec.SetAppId(uint64(a.AppID))
	spec.SetServiceId(uint64(a.ServiceID))
	spec.SetPort(uint16(a.RPCPort))
	return spec, nil
}

func (a *RPCServiceSpec) Validate() error {
	if a.DomainID == DomainID(0) {
		return ErrDomainIDZero
	}
	if a.AppID == AppID(0) {
		return ErrAppIDZero
	}
	if a.ServiceID == ServiceID(0) {
		return ErrServiceIDZero
	}
	if a.RPCPort == RPCPort(0) {
		return ErrRPCPortZero
	}
	return nil
}

// RPCPort represents an RPC port
type RPCPort uint16

func CheckRPCServerSpec(spec config.RPCServerSpec) error {
	if !spec.HasServerCert() {
		return NewRPCServerSpecError(errors.New("ServerCert is required"))
	}
	serverCert, err := spec.ServerCert()
	if err != nil {
		return err
	}
	if !serverCert.HasCert() {
		return NewRPCServerSpecError(errors.New("ServerCert.Cert is required"))
	}
	if !serverCert.HasKey() {
		return NewRPCServerSpecError(errors.New("ServerCert.Key is required"))
	}

	if !spec.HasCaCert() {
		return NewRPCServerSpecError(errors.New("CaCert is required"))
	}

	return nil
}

func NewRPCServerSpec(spec config.RPCServerSpec) (*RPCServerSpec, error) {
	if err := CheckRPCServerSpec(spec); err != nil {
		return nil, err
	}

	serviceSpec, err := spec.RpcServiceSpec()
	if err != nil {
		return nil, err
	}
	rpcServiceSpec, err := NewRPCServiceSpec(serviceSpec)
	if err != nil {
		return nil, err
	}
	serverSpec := &RPCServerSpec{
		RPCServiceSpec: rpcServiceSpec,
		ClientCAs:      x509.NewCertPool(),
		MaxConns:       spec.MaxConns(),
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

// RPCServerSpec is the server spec for the RPCService
type RPCServerSpec struct {
	*RPCServiceSpec
	ClientCAs *x509.CertPool
	Cert      tls.Certificate
	MaxConns  uint32
}

func (a *RPCServerSpec) ToCapnp(s *capnp.Segment) (config.RPCServerSpec, error) {
	serverSpec, err := config.NewRootRPCServerSpec(s)
	if err != nil {
		return serverSpec, err
	}

	serviceSpec, err := a.RPCServiceSpec.ToCapnp(s)
	if err != nil {
		return serverSpec, err
	}
	serverSpec.SetRpcServiceSpec(serviceSpec)

	return serverSpec, nil
}

func (a *RPCServerSpec) TLSConfig() *tls.Config {
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

func (a *RPCServerSpec) TLSConfigProvider() TLSConfigProvider {
	return func() (*tls.Config, error) {
		return a.TLSConfig(), nil
	}
}

func (a *RPCServerSpec) ListenerFactory() ListenerFactory {
	return func() (net.Listener, error) {
		return net.Listen("tcp", fmt.Sprintf(":%d", a.RPCPort))
	}
}

func (a *RPCServerSpec) StartRPCService(service *Service, mainInterface RPCMainInterface) (*RPCService, error) {
	return StartRPCService(service, a.ListenerFactory(), a.TLSConfigProvider(), mainInterface, uint(a.MaxConns))
}

func CheckRPCClientSpec(spec config.RPCClientSpec) error {
	if !spec.HasClientCert() {
		return NewRPCClientSpecError(errors.New("Client cert is required"))
	}
	serverCert, err := spec.ClientCert()
	if err != nil {
		return err
	}
	if !serverCert.HasCert() {
		return NewRPCClientSpecError(errors.New("Client cert is required"))
	}
	if !serverCert.HasKey() {
		return NewRPCClientSpecError(errors.New("Client key is required"))
	}

	if !spec.HasCaCert() {
		return NewRPCClientSpecError(errors.New("CA cert is required"))
	}

	return nil
}

func NewRPCClientSpec(spec config.RPCClientSpec) (*RPCClientSpec, error) {
	if err := CheckRPCClientSpec(spec); err != nil {
		return nil, err
	}
	serviceSpec, err := spec.RpcServiceSpec()
	if err != nil {
		return nil, err
	}
	rpcServiceSpec, err := NewRPCServiceSpec(serviceSpec)
	if err != nil {
		return nil, err
	}
	clientSpec := &RPCClientSpec{
		RPCServiceSpec: rpcServiceSpec,
		RootCAs:        x509.NewCertPool(),
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

// RPCClientSpec is the client spec for the RPCService
type RPCClientSpec struct {
	*RPCServiceSpec
	RootCAs *x509.CertPool
	Cert    tls.Certificate
}

func (a *RPCClientSpec) TLSConfig() *tls.Config {
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
func (a *RPCClientSpec) Conn() (*rpc.Conn, error) {
	networkAddr := a.NetworkAddr()
	Logger().Debug().
		Uint64("service", uint64(a.ServiceID)).
		Str("NetworkAddr", networkAddr).
		Msg("RPCClientSpec")
	addr := fmt.Sprintf("%s:%d", networkAddr, a.RPCPort)
	clientConn, err := tls.Dial("tcp", addr, a.TLSConfig())
	if err != nil {
		return nil, err
	}
	return rpc.NewConn(rpc.StreamTransport(clientConn)), nil
}

// ConnForAddr returns an RPC conn using the specified network address
// This mainly intended for testing purposes to connect locally
func (a *RPCClientSpec) ConnForAddr(networkAddr string) (*rpc.Conn, error) {
	addr := fmt.Sprintf("%s:%d", networkAddr, a.RPCPort)
	clientConn, err := tls.Dial("tcp", addr, a.TLSConfig())
	if err != nil {
		return nil, err
	}
	return rpc.NewConn(rpc.StreamTransport(clientConn)), nil
}
