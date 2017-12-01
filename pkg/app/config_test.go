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
	"fmt"
	"io/ioutil"
	"os"
	"testing"

	"bytes"
	"io"

	"github.com/oysterpack/oysterpack.go/pkg/app/config"
	"zombiezen.com/go/capnproto2"
)

func TestConfigServiceIDs(t *testing.T) {
	previousConfigDir := Configs.ConfigDir()
	defer func() { Configs.SetConfigDir(previousConfigDir) }()
	Configs.SetConfigDir("testdata/TestConfigServiceIDs")

	// Given that the config dir does not exist
	os.RemoveAll(Configs.ConfigDir())

	// When config service ids are looked up
	if _, err := Configs.ConfigServiceIDs(); err == nil {
		t.Errorf("An error should have been returned when the config dir does not exist")
	} else {
		// Then a ConfigError should be returned
		switch err := err.(type) {
		case ConfigError:
			t.Log(err)
		default:
			t.Errorf("Different error type was returned : %[1]T : %[1]v", err)
		}
	}

	// Given the config dir does exist
	if err := os.MkdirAll(Configs.ConfigDir(), 0755); err != nil {
		t.Fatal(err)
	}

	// When the config dir is empty
	if serviceIds, err := Configs.ConfigServiceIDs(); err != nil {
		t.Error(err)
	} else {
		// Then no service ids should be returned
		t.Logf("config service ids : %v", serviceIds)
		if len(serviceIds) != 0 {
			t.Error("There should be 0 service ids returned")
		}
	}

	// Given configs exist
	for i := 0; i < 10; i++ {
		ioutil.WriteFile(fmt.Sprintf("%s/0x%x", Configs.ConfigDir(), uint64(APP_RPC_SERVICE_ID)+uint64(i)), []byte{}, 0444)
	}
	for i := 10; i < 20; i++ {
		ioutil.WriteFile(fmt.Sprintf("%s/%d", Configs.ConfigDir(), uint64(APP_RPC_SERVICE_ID)+uint64(i)), []byte{}, 0444)
	}
	if serviceIds, err := Configs.ConfigServiceIDs(); err != nil {
		t.Error(err)
	} else {
		// Then the service ids should be returned
		t.Logf("config service ids : %v", serviceIds)
		if len(serviceIds) != 20 {
			t.Errorf("There should be 20 service ids returned : %d", len(serviceIds))
		}
	}
}

func TestConfig(t *testing.T) {
	previousConfigDir := Configs.ConfigDir()
	defer func() { Configs.SetConfigDir(previousConfigDir) }()
	Configs.SetConfigDir("testdata/TestConfig")

	os.RemoveAll(Configs.ConfigDir())
	if err := os.MkdirAll(Configs.ConfigDir(), 0755); err != nil {
		t.Fatal(err)
	}

	// Given a capnp message config for a service id
	storeAppRPCServerSpec := func(w io.Writer) {
		msg, seg, err := capnp.NewMessage(capnp.SingleSegment(nil))
		if err != nil {
			t.Fatal(err)
		}
		rpcServerSpec, err := config.NewRootRPCServerSpec(seg)
		if err != nil {
			t.Fatal(err)
		}
		rpcServiceSpec, err := config.NewRPCServiceSpec(seg)
		if err != nil {
			t.Fatal(err)
		}
		rpcServiceSpec.SetDomainID(0xe1b1125eac639831)
		rpcServiceSpec.SetAppId(0x81644b17b1ff35e3)
		rpcServiceSpec.SetServiceId(uint64(APP_RPC_SERVICE_ID))
		rpcServiceSpec.SetPort(uint16(44222))
		rpcServerSpec.SetRpcServiceSpec(rpcServiceSpec)

		MarshalCapnpMessage(msg, w)
	}

	buf := new(bytes.Buffer)
	storeAppRPCServerSpec(buf)
	msg, err := UnmarshalCapnpMessage(buf)
	if err != nil {
		t.Fatal(err)
	}
	spec, err := config.ReadRootRPCServerSpec(msg)
	if err != nil {
		t.Fatal(err)
	}

	checkConfig := func(spec config.RPCServerSpec) {
		t.Helper()
		rpcServiceSpec, err := spec.RpcServiceSpec()
		if err != nil {
			t.Fatal(err)
		}

		t.Logf("spec service id : %x", rpcServiceSpec.ServiceId())

		if rpcServiceSpec.DomainID() != 0xe1b1125eac639831 {
			t.Errorf("DomainID did not match : %x", rpcServiceSpec.DomainID())
		}
		if rpcServiceSpec.AppId() != 0x81644b17b1ff35e3 {
			t.Errorf("AppId did not match : %x", rpcServiceSpec.AppId())
		}
		if rpcServiceSpec.ServiceId() != uint64(APP_RPC_SERVICE_ID) {
			t.Errorf("ServiceId did not match : %x", rpcServiceSpec.ServiceId())
		}
		if rpcServiceSpec.Port() != uint16(44222) {
			t.Errorf("Port did not match : %x", rpcServiceSpec.Port())
		}
	}

	checkConfig(spec)

	configFile, err := os.Create(fmt.Sprintf("%s/0x%x", Configs.ConfigDir(), APP_RPC_SERVICE_ID))
	storeAppRPCServerSpec(configFile)
	configFile.Close()

	// When service ids are looked up
	if serviceIds, err := Configs.ConfigServiceIDs(); err != nil {
		t.Error(err)
	} else {
		// Then the service ids should be returned
		t.Logf("config service ids : %v", serviceIds)
		if len(serviceIds) != 1 {
			t.Errorf("There should be 20 service ids returned : %d", len(serviceIds))
		}
	}

	appRPCServiceSpec, err := Configs.Config(APP_RPC_SERVICE_ID)
	if err != nil {
		t.Fatal(err)
	}
	spec, err = config.ReadRootRPCServerSpec(appRPCServiceSpec)
	if err != nil {
		t.Fatal(err)
	}
	checkConfig(spec)
}
