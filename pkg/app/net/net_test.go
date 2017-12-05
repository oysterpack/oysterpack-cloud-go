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

package net_test

import (
	"fmt"

	"io/ioutil"

	"os"

	"github.com/oysterpack/oysterpack.go/pkg/app"
	appconfig "github.com/oysterpack/oysterpack.go/pkg/app/config"
	opnet "github.com/oysterpack/oysterpack.go/pkg/app/net"
	"github.com/oysterpack/oysterpack.go/pkg/app/net/config"
	"zombiezen.com/go/capnproto2"
)

const (
	EASY_PKI_ROOT = "./testdata/.easypki/pki"
)

var PKI = EasyPKI{EASY_PKI_ROOT}

type EasyPKI struct {
	rootDir string
}

func (a EasyPKI) CertFilePath(ca, cn string) string {
	return fmt.Sprintf("%s/%s/certs/%s.crt", a.rootDir, ca, cn)
}

func (a EasyPKI) KeyFilePath(ca, cn string) string {
	return fmt.Sprintf("%s/%s/keys/%s.key", a.rootDir, ca, cn)
}

func (a EasyPKI) ServiceSpec(seg *capnp.Segment, domainID app.DomainID, appID app.AppID, serviceID app.ServiceID, port opnet.ServerPort) (config.ServiceSpec, error) {
	serviceSpec, err := config.NewServiceSpec(seg)
	if err != nil {
		return serviceSpec, err
	}
	serviceSpec.SetDomainID(uint64(domainID))
	serviceSpec.SetAppId(uint64(appID))
	serviceSpec.SetServiceId(uint64(serviceID))
	serviceSpec.SetPort(uint16(port))
	return serviceSpec, nil
}

func (a EasyPKI) ServerSpec(seg *capnp.Segment, serviceSpec config.ServiceSpec, maxConns uint32) (config.ServerSpec, error) {
	serverSpec, err := config.NewRootServerSpec(seg)
	if err != nil {
		return serverSpec, err
	}
	serverSpec.SetMaxConns(maxConns)

	serverSpec.SetServiceSpec(serviceSpec)

	ca := fmt.Sprintf("%x", serviceSpec.DomainID())
	cacert, err := ioutil.ReadFile(a.CertFilePath(ca, ca))
	if err != nil {
		return serverSpec, err
	}
	if err := serverSpec.SetCaCert(cacert); err != nil {
		return serverSpec, err
	}

	serverCN := opnet.ServerCN(app.DomainID(serviceSpec.DomainID()), app.AppID(serviceSpec.AppId()), app.ServiceID(serviceSpec.ServiceId()))
	x509KeyPair, err := a.x509KeyPair(seg, ca, serverCN)
	if err != nil {
		return serverSpec, err
	}
	if err := serverSpec.SetServerCert(x509KeyPair); err != nil {
		return serverSpec, err
	}

	return serverSpec, nil
}

func (a EasyPKI) x509KeyPair(seg *capnp.Segment, ca, cn string) (config.X509KeyPair, error) {
	serverCert, err := ioutil.ReadFile(a.CertFilePath(ca, cn))
	if err != nil {
		return config.X509KeyPair{}, err
	}
	serverKey, err := ioutil.ReadFile(a.KeyFilePath(ca, cn))
	if err != nil {
		return config.X509KeyPair{}, err
	}
	x509KeyPair, err := config.NewX509KeyPair(seg)
	if err != nil {
		return x509KeyPair, err
	}
	if err := x509KeyPair.SetCert(serverCert); err != nil {
		return x509KeyPair, err
	}
	if err := x509KeyPair.SetKey(serverKey); err != nil {
		return x509KeyPair, err
	}
	return x509KeyPair, nil
}

func (a EasyPKI) ClientSpec(seg *capnp.Segment, domainID app.DomainID, appID app.AppID, serviceID app.ServiceID, port opnet.ServerPort, cn string) (config.ClientSpec, error) {
	clientSpec, err := config.NewRootClientSpec(seg)
	if err != nil {
		return clientSpec, err
	}
	serverServiceSpec, err := a.ServiceSpec(seg, app.DomainID(domainID), app.AppID(appID), app.ServiceID(serviceID), opnet.ServerPort(port))
	if err != nil {
		return clientSpec, err
	}
	clientSpec.SetServiceSpec(serverServiceSpec)

	ca := fmt.Sprintf("%x", domainID)
	cacert, err := ioutil.ReadFile(a.CertFilePath(ca, ca))
	if err != nil {
		return clientSpec, err
	}
	if err := clientSpec.SetCaCert(cacert); err != nil {
		return clientSpec, err
	}
	x509KeyPai, err := a.x509KeyPair(seg, ca, cn)
	if err != nil {
		return clientSpec, err
	}
	clientSpec.SetClientCert(x509KeyPai)
	return clientSpec, nil
}

func initConfigDir(configDir string) {
	app.Configs.SetConfigDir(configDir)
	os.RemoveAll(configDir)
	os.MkdirAll(configDir, 0755)
}

func initServerMetricsConfig(serviceID app.ServiceID) error {
	msg, seg, err := capnp.NewMessage(capnp.SingleSegment(nil))
	if err != nil {
		return err
	}
	metricsServiceSpec, err := appconfig.NewRootMetricsServiceSpec(seg)
	if err != nil {
		return err
	}
	metricsSpecs, err := appconfig.NewMetricSpecs(seg)
	if err != nil {
		return err
	}
	metricsServiceSpec.SetMetricSpecs(metricsSpecs)

	// configure gauge specs
	gauges, err := metricsSpecs.NewGaugeSpecs(1)
	if err != nil {
		return err
	}
	gauge, err := appconfig.NewGaugeMetricSpec(seg)
	if err != nil {
		return err
	}
	gauge.SetServiceId(serviceID.UInt64())
	gauge.SetMetricId(opnet.SERVER_CONN_COUNT_METRIC_ID.UInt64())
	if err := gauge.SetHelp("Server connection count"); err != nil {
		return err
	}
	gauges.Set(0, gauge)

	// store the config
	serviceConfigPath := app.Configs.ServiceConfigPath(app.METRICS_SERVICE_ID)
	configFile, err := os.Create(serviceConfigPath)
	if err != nil {
		return err
	}
	app.MarshalCapnpMessage(msg, configFile)
	configFile.Close()

	return nil
}
