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
	"compress/zlib"
	"io"

	"os"

	"strconv"

	"fmt"
	"io/ioutil"

	"bytes"

	"zombiezen.com/go/capnproto2"
)

const (
	CONFIG_SERVICE_ID = ServiceID(0x86813dd78fd207f5)
)

func initConfigService() {
	service := Services.Service(CONFIG_SERVICE_ID)
	if service == nil {
		Services.Register(NewService(CONFIG_SERVICE_ID))
	}
}

// AppConfig is used to group together config related functions
type AppConfig struct{}

// ConfigDir returns the config dir path
func (a AppConfig) ConfigDir() string {
	configDirMutex.RLock()
	defer configDirMutex.RUnlock()
	return configDir
}

// SetConfigDir is only exposed for testing purposes.
// This enables tests to setup test configurations
func (a AppConfig) SetConfigDir(dir string) {
	configDirMutex.Lock()
	defer configDirMutex.Unlock()
	configDir = dir
}

// ConfigServiceIDs returns the list of ServiceID(s) for which configs exist
// The ServiceID is used as the config id.
func (a AppConfig) ConfigServiceIDs() ([]ServiceID, error) {
	dir, err := os.Open(a.ConfigDir())
	if err != nil {
		if err == os.ErrNotExist {
			return []ServiceID{}, nil
		}
		return nil, ConfigError(CONFIG_SERVICE_ID, err, fmt.Sprintf("Failed to open the config directory : %v", a.ConfigDir()))
	}
	defer dir.Close()
	dirStat, err := dir.Stat()
	if err != nil {
		return nil, ConfigError(CONFIG_SERVICE_ID, err, fmt.Sprintf("Failed to stat the config directory : %v", a.ConfigDir()))
	}
	if !dirStat.IsDir() {
		return nil, ConfigError(CONFIG_SERVICE_ID, fmt.Errorf("Config dir is not a dir : %s", a.ConfigDir()), "")
	}

	names, err := dir.Readdirnames(0)
	if err != nil {
		return nil, ConfigError(CONFIG_SERVICE_ID, err, "Failed to get directory listing")
	}
	serviceIds := make([]ServiceID, len(names))
	for i, name := range names {
		if id, err := strconv.ParseUint(name, 0, 64); err != nil {
			return nil, ConfigError(CONFIG_SERVICE_ID, fmt.Errorf("Invalid ServiceID : %v : %v", name, err), "")
		} else {
			serviceIds[i] = ServiceID(id)
		}
	}

	return serviceIds, nil
}

// ServiceConfigID is used to convert the ServiceID into the config id. The config id will be the ServiceID in HEX format.
func (a AppConfig) ServiceConfigID(id ServiceID) string {
	return fmt.Sprintf("0x%x", id)
}

// ServiceConfigExists returns true if a config exists for the specified ServiceID
func (a AppConfig) ServiceConfigExists(id ServiceID) bool {
	f, err := os.Open(a.ServiceConfigPath(id))
	if err != nil {
		return false
	}
	f.Close()
	return true
}

// ConfigDirExists returns true if the config dir exists
func (a AppConfig) ConfigDirExists() bool {
	dir, err := os.Open(a.ConfigDir())
	if err != nil {
		return false
	}
	dir.Close()
	return true
}

// ServiceConfigPath returns the service config file path
func (a AppConfig) ServiceConfigPath(id ServiceID) string {
	return fmt.Sprintf("%s/%s", a.ConfigDir(), a.ServiceConfigID(id))
}

// Config returns the config message for the specified ServiceID.
// The config is expected to exist at {configDir}/{ServiceID}, where the ServiceID is specified in HEX format, e.g.,
//
//	/run/secrets/0xe49214fa20b35ba8
//
// If no config exists for the specified service, then nil is returned.
//
// errors:
//	- ConfigError
func (a AppConfig) Config(id ServiceID) (*capnp.Message, error) {
	c, err := ioutil.ReadFile(a.ServiceConfigPath(id))
	if err != nil {
		if err == os.ErrNotExist {
			return nil, nil
		}
		switch err := err.(type) {
		case *os.PathError:
			return nil, nil
		default:
			return nil, ConfigError(id, err, "Failed to read service config file")
		}
	}

	msg, err := UnmarshalCapnpMessage(bytes.NewBuffer(c))
	if err != nil {
		return nil, ConfigError(id, err, "Failed to unmarshal service config file")
	}
	return msg, err
}

// MarshalCapnpMessage marshals the capnp message to the writer.
// The message will be compressed using zlib.
func MarshalCapnpMessage(msg *capnp.Message, out io.Writer) error {
	compressor := zlib.NewWriter(out)
	encoder := capnp.NewPackedEncoder(compressor)
	if err := encoder.Encode(msg); err != nil {
		return err
	}
	compressor.Close()
	return nil
}

// UnmarshalCapnpMessage unmarshals the capnp message using the specified reader.
// The message should have have marshalled via MarshalCapnpMessage(), i.e., it must be zlib compressed.
func UnmarshalCapnpMessage(in io.Reader) (*capnp.Message, error) {
	decompressor, _ := zlib.NewReader(in)
	decoder := capnp.NewPackedDecoder(decompressor)
	msg, err := decoder.Decode()
	if err != nil {
		return nil, err
	}
	return msg, nil
}
