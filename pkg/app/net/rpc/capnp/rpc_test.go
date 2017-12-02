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

package capnp

// this file is used to put shared testing util code

import (
	"context"
	"net"

	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"

	"errors"

	"github.com/oysterpack/oysterpack.go/pkg/app"
	"github.com/oysterpack/oysterpack.go/pkg/app/capnprpc"
	"zombiezen.com/go/capnproto2/rpc"
)

const (
	EASY_PKI_ROOT = "./testdata/.easypki/pki"
	EASY_PKI_CA   = "app.dev.oysterpack.com"

	TINY_CERT_PKI_ROOT = "./testdata/tinycert"
)

var (
	tlsProvider TLSProvider = EasyPKITLS{}
	//tlsProvider TLSProvider = TinyCertTLS{}
)

func appClientConn(addr net.Addr) (capnprpc.App, net.Conn, error) {
	clientConn, err := net.Dial(addr.Network(), addr.String())
	if err != nil {
		return capnprpc.App{}, nil, err
	}
	rpcClient := rpc.NewConn(rpc.StreamTransport(clientConn))
	return capnprpc.App{Client: rpcClient.Bootstrap(context.Background())}, clientConn, nil
}

func appTLSClientConn(addr net.Addr) (capnprpc.App, net.Conn, error) {
	return tlsProvider.AppTLSClientConn(addr)
}

func EasyPKICertFilePath(ca, cn string) string {
	return fmt.Sprintf("%s/%s/certs/%s.crt", EASY_PKI_ROOT, ca, cn)
}

func EasyPKIKeyFilePath(ca, cn string) string {
	return fmt.Sprintf("%s/%s/keys/%s.key", EASY_PKI_ROOT, ca, cn)
}

type TLSProvider interface {
	CACertPool() (*x509.CertPool, error)
	ClientTLSConfig() (*tls.Config, error)
	ServerTLSConfig() (*tls.Config, error)

	AppTLSClientConn(addr net.Addr) (capnprpc.App, net.Conn, error)
}

type EasyPKITLS_DomainAppService struct {
	app.DomainID
	app.AppID
	app.ServiceID

	CACerts []string
}

func (a EasyPKITLS_DomainAppService) ServiceCN() string {
	return fmt.Sprintf("%x.%x.%x", a.ServiceID, a.AppID, a.DomainID)
}

func (a EasyPKITLS_DomainAppService) AppTLSClientConn(addr net.Addr) (capnprpc.App, net.Conn, error) {
	tlsConfig, err := a.ClientTLSConfig()
	if err != nil {
		return capnprpc.App{}, nil, err
	}
	clientConn, err := tls.Dial(addr.Network(), addr.String(), tlsConfig)
	if err != nil {
		return capnprpc.App{}, nil, err
	}
	rpcClient := rpc.NewConn(rpc.StreamTransport(clientConn))
	return capnprpc.App{Client: rpcClient.Bootstrap(context.Background())}, clientConn, nil
}

func (a EasyPKITLS_DomainAppService) CACertPool() (*x509.CertPool, error) {
	pool := x509.NewCertPool()

	for _, ca := range a.CACerts {
		caCert := a.CertFilePath(ca)
		rootPEM, err := ioutil.ReadFile(caCert)
		if err != nil || rootPEM == nil {
			return nil, err
		}
		ok := pool.AppendCertsFromPEM([]byte(rootPEM))
		if !ok {
			return nil, errors.New("rootCA() : failed to parse root certificate")
		}
	}
	return pool, nil
}

func (a EasyPKITLS_DomainAppService) ClientTLSConfig() (*tls.Config, error) {
	const CERT_NAME = "client.dev.oysterpack.com"
	certKeyPair, err := tls.LoadX509KeyPair(a.CertFilePath(CERT_NAME), a.KeyFilePath(CERT_NAME))
	if err != nil {
		panic(err)
	}

	rootCAs, err := a.CACertPool()
	if err != nil {
		return nil, err
	}

	tlsConfig := &tls.Config{
		MinVersion:   tls.VersionTLS12,
		RootCAs:      rootCAs,
		Certificates: []tls.Certificate{certKeyPair},

		InsecureSkipVerify: false,
		ServerName:         a.ServiceCN(),
	}
	tlsConfig.BuildNameToCertificate()
	return tlsConfig, nil
}

func (a EasyPKITLS_DomainAppService) CertFilePath(cn string) string {
	return fmt.Sprintf("%s/%s/certs/%s.crt", EASY_PKI_ROOT, EASY_PKI_CA, cn)
}

func (a EasyPKITLS_DomainAppService) KeyFilePath(cn string) string {
	return fmt.Sprintf("%s/%s/keys/%s.key", EASY_PKI_ROOT, EASY_PKI_CA, cn)
}

func (a EasyPKITLS_DomainAppService) ServerTLSConfig() (*tls.Config, error) {
	var certName = a.ServiceCN()
	cert, err := ioutil.ReadFile(a.CertFilePath(certName))
	if err != nil {
		return nil, err
	}
	key, err := ioutil.ReadFile(a.KeyFilePath(certName))
	if err != nil {
		return nil, err
	}
	certKeyPair, err := tls.X509KeyPair(cert, key)
	if err != nil {
		return nil, err
	}

	rootCAs, err := a.CACertPool()
	if err != nil {
		return nil, err
	}

	tlsConfig := &tls.Config{
		//RootCAs: rootCAs,

		MinVersion: tls.VersionTLS12,

		// Reject any TLS certificate that cannot be validated
		ClientAuth: tls.RequireAndVerifyClientCert,

		// Ensure that we only use our "CA" to validate certificates
		ClientCAs: rootCAs,
		// Server cert
		Certificates: []tls.Certificate{certKeyPair},

		PreferServerCipherSuites: true,
		//CipherSuites:             []uint16{tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384},
	}
	tlsConfig.BuildNameToCertificate()
	return tlsConfig, nil
}

type EasyPKITLS struct {
}

func (a EasyPKITLS) AppTLSClientConn(addr net.Addr) (capnprpc.App, net.Conn, error) {
	tlsConfig, err := a.ClientTLSConfig()
	if err != nil {
		return capnprpc.App{}, nil, err
	}
	clientConn, err := tls.Dial(addr.Network(), addr.String(), tlsConfig)
	if err != nil {
		return capnprpc.App{}, nil, err
	}
	rpcClient := rpc.NewConn(rpc.StreamTransport(clientConn))
	return capnprpc.App{Client: rpcClient.Bootstrap(context.Background())}, clientConn, nil
}

func (a EasyPKITLS) CACertPool() (*x509.CertPool, error) {
	pool := x509.NewCertPool()
	caCert := fmt.Sprintf("%s/%[2]s/certs/%[2]s.crt", EASY_PKI_ROOT, EASY_PKI_CA)
	rootPEM, err := ioutil.ReadFile(caCert)
	if err != nil || rootPEM == nil {
		return nil, err
	}
	ok := pool.AppendCertsFromPEM([]byte(rootPEM))
	if !ok {
		return nil, errors.New("rootCA() : failed to parse root certificate")
	}
	return pool, nil
}

func (a EasyPKITLS) ClientTLSConfig() (*tls.Config, error) {
	const CERT_NAME = "client.dev.oysterpack.com"
	certKeyPair, err := tls.LoadX509KeyPair(
		fmt.Sprintf("%s/%s/certs/%s.crt", EASY_PKI_ROOT, EASY_PKI_CA, CERT_NAME),
		fmt.Sprintf("%s/%s/keys/%s.key", EASY_PKI_ROOT, EASY_PKI_CA, CERT_NAME),
	)
	if err != nil {
		panic(err)
	}

	rootCAs, err := a.CACertPool()
	if err != nil {
		return nil, err
	}

	tlsConfig := &tls.Config{
		MinVersion:   tls.VersionTLS12,
		RootCAs:      rootCAs,
		Certificates: []tls.Certificate{certKeyPair},

		InsecureSkipVerify: false,
		ServerName:         "server.dev.oysterpack.com",
	}
	tlsConfig.BuildNameToCertificate()
	return tlsConfig, nil
}

func (a EasyPKITLS) ServerTLSConfig() (*tls.Config, error) {
	const certName = "server.dev.oysterpack.com"
	cert, err := ioutil.ReadFile(fmt.Sprintf("%s/%s/certs/%s.crt", EASY_PKI_ROOT, EASY_PKI_CA, certName))
	if err != nil {
		return nil, err
	}
	key, err := ioutil.ReadFile(fmt.Sprintf("%s/%s/keys/%s.key", EASY_PKI_ROOT, EASY_PKI_CA, certName))
	if err != nil {
		return nil, err
	}
	certKeyPair, err := tls.X509KeyPair(cert, key)
	if err != nil {
		return nil, err
	}

	rootCAs, err := a.CACertPool()
	if err != nil {
		return nil, err
	}

	tlsConfig := &tls.Config{
		//RootCAs: rootCAs,

		MinVersion: tls.VersionTLS12,

		// Reject any TLS certificate that cannot be validated
		ClientAuth: tls.RequireAndVerifyClientCert,

		// Ensure that we only use our "CA" to validate certificates
		ClientCAs: rootCAs,
		// Server cert
		Certificates: []tls.Certificate{certKeyPair},

		PreferServerCipherSuites: true,
		//CipherSuites:             []uint16{tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384},
	}
	tlsConfig.BuildNameToCertificate()
	return tlsConfig, nil
}

type TinyCertTLS struct {
}

func (a TinyCertTLS) AppTLSClientConn(addr net.Addr) (capnprpc.App, net.Conn, error) {
	tlsConfig, err := a.ClientTLSConfig()
	if err != nil {
		return capnprpc.App{}, nil, err
	}
	clientConn, err := tls.Dial(addr.Network(), addr.String(), tlsConfig)
	if err != nil {
		return capnprpc.App{}, nil, err
	}
	rpcClient := rpc.NewConn(rpc.StreamTransport(clientConn))
	return capnprpc.App{Client: rpcClient.Bootstrap(context.Background())}, clientConn, nil
}

func (a TinyCertTLS) ServerTLSConfig() (*tls.Config, error) {
	const certName = "server.dev.oysterpack.com"
	cert, err := ioutil.ReadFile(fmt.Sprintf("%s/%s.pem", TINY_CERT_PKI_ROOT, certName))
	if err != nil {
		return nil, err
	}
	key, err := ioutil.ReadFile(fmt.Sprintf("%s/%s.key.pem", TINY_CERT_PKI_ROOT, certName))
	if err != nil {
		return nil, err
	}
	certKeyPair, err := tls.X509KeyPair(cert, key)
	if err != nil {
		return nil, err
	}

	rootCAs, err := a.CACertPool()
	if err != nil {
		return nil, err
	}

	tlsConfig := &tls.Config{
		RootCAs: rootCAs,

		MinVersion: tls.VersionTLS12,

		// Reject any TLS certificate that cannot be validated
		ClientAuth: tls.RequireAndVerifyClientCert,

		// Ensure that we only use our "CA" to validate certificates
		ClientCAs: rootCAs,
		// Server cert
		Certificates: []tls.Certificate{certKeyPair},

		PreferServerCipherSuites: true,
		//CipherSuites:             []uint16{tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384},
	}
	tlsConfig.BuildNameToCertificate()
	return tlsConfig, nil
}

func (a TinyCertTLS) ClientTLSConfig() (*tls.Config, error) {
	const CERT_NAME = "client.dev.oysterpack.com"
	certKeyPair, err := tls.LoadX509KeyPair(
		fmt.Sprintf("%s/%s.pem", TINY_CERT_PKI_ROOT, CERT_NAME),
		fmt.Sprintf("%s/%s.key.pem", TINY_CERT_PKI_ROOT, CERT_NAME),
	)
	if err != nil {
		panic(err)
	}

	rootCAs, err := a.CACertPool()
	if err != nil {
		return nil, err
	}

	tlsConfig := &tls.Config{
		MinVersion:   tls.VersionTLS12,
		RootCAs:      rootCAs,
		Certificates: []tls.Certificate{certKeyPair},

		InsecureSkipVerify: false,
		ServerName:         "server.dev.oysterpack.com",
	}
	tlsConfig.BuildNameToCertificate()
	return tlsConfig, nil
}

func (a TinyCertTLS) CACertPool() (*x509.CertPool, error) {
	pool := x509.NewCertPool()
	caCert := TINY_CERT_PKI_ROOT + "/dev.oysterpack.com.cacert.pem"
	rootPEM, err := ioutil.ReadFile(caCert)
	if err != nil || rootPEM == nil {
		return nil, err
	}
	ok := pool.AppendCertsFromPEM([]byte(rootPEM))
	if !ok {
		return nil, errors.New("rootCA() : failed to parse root certificate")
	}
	return pool, nil
}
