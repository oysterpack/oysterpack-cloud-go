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
)

// RPCServiceSpec is the common RPCService spec shared by the RPC server and client
type RPCServiceSpec struct {
	DomainID
	AppID
	ServiceID

	RPCPort
}

// CN returns the x509 CN - this used by client TLS to set the x509.Config.ServerName
func (a *RPCServiceSpec) CN() string {
	return fmt.Sprintf("%x.%x.%x", a.ServiceID, a.AppID, a.DomainID)
}

// NetworkAddr returns the service network address, which is used by the client to connect to the service
func (a *RPCServiceSpec) NetworkAddr() string {
	return fmt.Sprintf("%x_%x", a.DomainID, a.AppID)
}

// RPCPort represents an RPC port
type RPCPort uint

// RPCServerSpec is the server spec for the RPCService
type RPCServerSpec struct {
	RPCServiceSpec
	ClientCAs *x509.CertPool
	Cert      tls.Certificate
}

// RPCClientSpec is the client spec for the RPCService
type RPCClientSpec struct {
	RPCServiceSpec
	RootCAs *x509.CertPool
	Cert    tls.Certificate
}
