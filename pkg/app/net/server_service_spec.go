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
	"fmt"

	"github.com/oysterpack/oysterpack.go/pkg/app"
	"github.com/oysterpack/oysterpack.go/pkg/app/net/config"
	"zombiezen.com/go/capnproto2"
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
	domainID  app.DomainID
	appID     app.AppID
	serviceID app.ServiceID

	serverPort ServerPort
}

func (a *ServerServiceSpec) DomainID() app.DomainID {
	return a.domainID
}

func (a *ServerServiceSpec) AppID() app.AppID {
	return a.appID
}

func (a *ServerServiceSpec) ServiceID() app.ServiceID {
	return a.serviceID
}

func (a *ServerServiceSpec) ServerPort() ServerPort {
	return a.serverPort
}

// CN returns the x509 CN - this used by client TLS to set the x509.Config.ServerName
func (a *ServerServiceSpec) CN() string {
	return ServerCN(a.domainID, a.appID, a.serviceID)
}

// NetworkAddr returns the service network address, which is used by the client to connect to the service.
// The network address uses the following naming convention :
//
//		fmt.Sprintf("%x_%x", a.DomainID, a.AppID)
//
//		e.g. ed5cf026e8734361-d113a2e016e12f0f
func (a *ServerServiceSpec) NetworkAddr() string {
	return fmt.Sprintf("%x_%x", a.domainID, a.appID)
}

func (a *ServerServiceSpec) ToCapnp(s *capnp.Segment) (config.ServiceSpec, error) {
	spec, err := config.NewServiceSpec(s)
	if err != nil {
		return spec, err
	}
	spec.SetDomainID(uint64(a.domainID))
	spec.SetAppId(uint64(a.appID))
	spec.SetServiceId(uint64(a.serviceID))
	spec.SetPort(uint16(a.serverPort))
	return spec, nil
}

func (a *ServerServiceSpec) Validate() error {
	if a.domainID == app.DomainID(0) {
		return app.ErrDomainIDZero
	}
	if a.appID == app.AppID(0) {
		return app.ErrAppIDZero
	}
	if a.serviceID == app.ServiceID(0) {
		return app.ErrServiceIDZero
	}
	if a.serverPort == ServerPort(0) {
		return ErrServerPortZero
	}
	return nil
}

// ServerPort represents a server network port
type ServerPort uint16
