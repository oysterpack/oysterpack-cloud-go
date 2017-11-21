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
	"errors"
	"io/ioutil"
	"path/filepath"
	"strings"
	"sync"
)

var (
	certPoolsLock sync.Mutex
	certPools     = make(map[string]*x509.CertPool)

	tlsConfigsLock sync.Mutex
	tlsConfigs     = make(map[string]*tls.Config)
)

func CACertPool(pkiDir, cacert string) (*x509.CertPool, error) {
	pkiDir = strings.TrimSpace(pkiDir)
	cacert = strings.TrimSpace(cacert)
	certPoolsLock.Lock()
	defer certPoolsLock.Unlock()
	certPath := filepath.Join(pkiDir, cacert, "certs", cacert+".crt")
	Logger().Debug().Str("certPath", certPath).Msg("CACertPool")
	certPool := certPools[certPath]
	if certPool != nil {
		return certPool, nil
	}

	certPool = x509.NewCertPool()
	rootPEM, err := ioutil.ReadFile(certPath)
	if err != nil || rootPEM == nil {
		return nil, err
	}
	ok := certPool.AppendCertsFromPEM([]byte(rootPEM))
	if !ok {
		return nil, errors.New("rootCA() : failed to parse root certificate")
	}

	certPools[certPath] = certPool
	Logger().Debug().Str("certPath", certPath).Msg("CACertPool - registered")
	return certPool, nil
}

func ServerTLSConfig(pkiDir, cacert, cert string) (*tls.Config, error) {
	pkiDir = strings.TrimSpace(pkiDir)
	cacert = strings.TrimSpace(cacert)
	cert = strings.TrimSpace(cert)

	if pkiDir == "" && cacert == "" && cert == "" {
		return nil, nil
	}
	certPath := filepath.Join(pkiDir, cacert, "certs", cert+".crt")
	tlsConfig := tlsConfigs[certPath]
	if tlsConfig != nil {
		return tlsConfig, nil
	}

	certPool, err := CACertPool(pkiDir, cacert)
	if err != nil {
		return nil, err
	}

	certPEM, err := ioutil.ReadFile(certPath)
	if err != nil {
		return nil, err
	}
	keyPem, err := ioutil.ReadFile(filepath.Join(pkiDir, cacert, "keys", cert+".key"))
	if err != nil {
		return nil, err
	}
	certKeyPair, err := tls.X509KeyPair(certPEM, keyPem)
	if err != nil {
		return nil, err
	}

	tlsConfig = &tls.Config{
		RootCAs: certPool,

		MinVersion: tls.VersionTLS12,

		// Reject any TLS certificate that cannot be validated
		ClientAuth: tls.RequireAndVerifyClientCert,

		// Ensure that we only use our "CA" to validate certificates
		ClientCAs: certPool,
		// Server cert
		Certificates: []tls.Certificate{certKeyPair},

		PreferServerCipherSuites: true,
		//CipherSuites:             []uint16{tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384},
	}
	tlsConfig.BuildNameToCertificate()

	tlsConfigs[certPath] = tlsConfig
	Logger().Debug().Str("certPath", certPath).Msg("ServerTLSConfig - registered")
	return tlsConfig, nil
}

func ClientTLSConfig(pkiDir, cacert, cert string) (*tls.Config, error) {
	pkiDir = strings.TrimSpace(pkiDir)
	cacert = strings.TrimSpace(cacert)
	cert = strings.TrimSpace(cert)

	if pkiDir == "" && cacert == "" && cert == "" {
		return nil, nil
	}

	certPath := filepath.Join(pkiDir, cacert, "certs", cert+".crt")
	tlsConfig := tlsConfigs[certPath]
	if tlsConfig != nil {
		return tlsConfig, nil
	}

	certPool, err := CACertPool(pkiDir, cacert)
	if err != nil {
		return nil, err
	}

	certPEM, err := ioutil.ReadFile(certPath)
	if err != nil {
		return nil, err
	}
	keyPem, err := ioutil.ReadFile(filepath.Join(pkiDir, cacert, "keys", cert+".key"))
	if err != nil {
		return nil, err
	}
	certKeyPair, err := tls.X509KeyPair(certPEM, keyPem)
	if err != nil {
		return nil, err
	}

	tlsConfig = &tls.Config{
		Certificates: []tls.Certificate{certKeyPair},
		RootCAs:      certPool,

		MinVersion:         tls.VersionTLS12,
		InsecureSkipVerify: false,
		ServerName:         "server.dev.oysterpack.com",
	}
	tlsConfig.BuildNameToCertificate()

	tlsConfigs[certPath] = tlsConfig
	Logger().Debug().Str("certPath", certPath).Msg("ClientTLSConfig - registered")
	return tlsConfig, nil
}
