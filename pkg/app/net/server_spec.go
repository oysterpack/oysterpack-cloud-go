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

	"time"

	"errors"

	"github.com/oysterpack/oysterpack.go/pkg/app"
	"github.com/oysterpack/oysterpack.go/pkg/app/net/config"
	"zombiezen.com/go/capnproto2"
)

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
		ServerServiceSpec:   serverServiceSpec,
		clientCAs:           x509.NewCertPool(),
		maxConns:            spec.MaxConns(),
		keepAlivePeriodSecs: spec.KeepAlivePeriodSecs(),
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
	serverSpec.cert, err = tls.X509KeyPair(cert, key)
	if err != nil {
		return nil, err
	}

	caCert, err := spec.CaCert()
	if err != nil {
		return nil, err
	}
	if !serverSpec.clientCAs.AppendCertsFromPEM(caCert) {
		return nil, app.ConfigError(serverSpec.ServiceID(), errors.New("Failed to parse PEM encoded cert(s)"), "")
	}

	return serverSpec, nil
}

func CheckServerSpec(spec config.ServerSpec) error {
	if !spec.HasServerCert() {
		return app.IllegalArgumentError("ServerCert is required")
	}
	serverCert, err := spec.ServerCert()
	if err != nil {
		return err
	}
	if !serverCert.HasCert() {
		return app.IllegalArgumentError("ServerCert.Cert is required")
	}
	if !serverCert.HasKey() {
		return app.IllegalArgumentError("ServerCert.Key is required")
	}

	if !spec.HasCaCert() {
		return app.IllegalArgumentError("CaCert is required")
	}

	if spec.MaxConns() == 0 {
		return app.IllegalArgumentError("Server max conn must be > 0")
	}

	if spec.KeepAlivePeriodSecs() == 0 {
		return app.IllegalArgumentError("Server conn keep alive period must be > 0")
	}

	return nil
}

// ServerSpec is the server spec for the Service
type ServerSpec struct {
	*ServerServiceSpec
	clientCAs           *x509.CertPool
	cert                tls.Certificate
	maxConns            uint32
	keepAlivePeriodSecs uint8

	// If > 0, sets the size of the operating system's receive buffer associated with the connection.
	readBufferSize uint
	// If > 0, sets the size of the operating system's transmit buffer associated with the connection.
	writeBufferSize uint

	// The ReadTimeout is to protect the server from connections that are idle for an extended period
	// of time. It also protects the server from very slow clients, which could potentially be intentionally slow as a
	// for of Ddos attack.
	//
	// if > 0, then used to set the connection read deadline
	readTimeout time.Duration

	// The WriteTimeout is to protect the server from very slow clients, which could potentially be intentionally slow as a
	// for of DDos attack. It may also be a sign og an unhealthy client. We want to protect server connection resources,
	// and close slow connections.
	//
	// if > 0, then used to set the connection write deadline
	writeTimeout time.Duration
}

func (a *ServerSpec) ClientCAs() *x509.CertPool {
	return a.clientCAs
}

func (a *ServerSpec) Cert() tls.Certificate {
	return a.cert
}

func (a *ServerSpec) MaxConns() uint32 {
	return a.maxConns
}

func (a *ServerSpec) KeepAlivePeriodSecs() uint8 {
	return a.keepAlivePeriodSecs
}

func (a *ServerSpec) ReadBufferSize() uint {
	return a.readBufferSize
}

func (a *ServerSpec) WriteBufferSize() uint {
	return a.writeBufferSize
}

func (a *ServerSpec) ReadTimeout() time.Duration {
	return a.readTimeout
}

func (a *ServerSpec) WriteTimeout() time.Duration {
	return a.writeTimeout
}

// ConfigureConnBuffers configures the conn read and write buffer sizes.
func (a *ServerSpec) ConfigureConnBuffers(conn net.Conn) error {
	if a.readBufferSize == 0 && a.writeBufferSize == 0 {
		return nil
	}
	type BufferedConn interface {
		SetReadBuffer(bytes int) error
		SetWriteBuffer(bytes int) error
	}

	c, ok := conn.(BufferedConn)
	if !ok {
		return app.ConfigError(a.ServiceID(), errors.New("Server conn buffers are not configurable"), "")
	}
	if a.readBufferSize > 0 {
		c.SetReadBuffer(int(a.readBufferSize))
	}
	if a.writeBufferSize > 0 {
		c.SetWriteBuffer(int(a.writeBufferSize))
	}
	return nil
}

// ConfigureConnReadDeadline is meant to be used within the connection handler.
// The deadline needs to be reset after each successful read.
func (a *ServerSpec) ConfigureConnReadDeadline(conn net.Conn) {
	if a.readTimeout > 0 {
		conn.SetReadDeadline(time.Now().Add(a.readTimeout))
	}
}

// ConfigureConnWriteDeadline is meant to be used within the connection handler.
// The deadline needs to be reset after each successful write.
func (a *ServerSpec) ConfigureConnWriteDeadline(conn net.Conn) {
	if a.writeTimeout > 0 {
		conn.SetWriteDeadline(time.Now().Add(a.writeTimeout))
	}
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

	serverSpec.SetMaxConns(a.maxConns)
	// TODO: set server Cert and client CA Cert

	return serverSpec, nil
}

func (a *ServerSpec) TLSConfig() *tls.Config {
	// Caveat, all these recommended settings apply only to the amd64 architecture, for which fast, constant time
	// implementations of the crypto primitives (AES-GCM, ChaCha20-Poly1305, P256) are available.

	tlsConfig := &tls.Config{
		MinVersion: tls.VersionTLS12,

		// Reject any TLS certificate that cannot be validated
		ClientAuth: tls.RequireAndVerifyClientCert,

		// Ensure that we only use our "CA" to validate certificates
		ClientCAs: a.clientCAs,
		// Server cert
		Certificates: []tls.Certificate{a.cert},

		PreferServerCipherSuites: true,

		// TODO: PFS because we can but this will reject client with RSA certificates
		//CipherSuites:             []uint16{tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384},

		CurvePreferences: []tls.CurveID{
			tls.CurveP256,
			tls.X25519, // Go 1.8 only
		},

		CipherSuites: []uint16{
			tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305, // Go 1.8 only
			tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305,   // Go 1.8 only
			tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
			tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,

			// Best disabled, as they don't provide Forward Secrecy,
			// but might be necessary for some clients
			// tls.TLS_RSA_WITH_AES_256_GCM_SHA384,
			// tls.TLS_RSA_WITH_AES_128_GCM_SHA256,
		},
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
		return net.Listen("tcp", fmt.Sprintf(":%d", a.serverPort))
	}
}
