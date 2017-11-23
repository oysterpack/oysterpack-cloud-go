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

// ConfigServiceIDs returns the list of ServiceID(s) for which configs exist
// The ServiceID is used as the config id.
func ConfigServiceIDs() ([]ServiceID, error) {
	dir, err := os.Open(configDir)
	if err != nil {
		if err == os.ErrNotExist {
			return []ServiceID{}, nil
		}
		return nil, NewConfigError(err)
	}
	defer dir.Close()
	dirStat, err := dir.Stat()
	if err != nil {
		return nil, NewConfigError(err)
	}
	if !dirStat.IsDir() {
		return nil, NewConfigError(fmt.Errorf("Config dir is not a dir : %s", configDir))
	}

	names, err := dir.Readdirnames(0)
	if err != nil {
		return nil, NewConfigError(err)
	}
	serviceIds := make([]ServiceID, len(names))
	for i, name := range names {
		if id, err := strconv.ParseUint(name, 0, 64); err != nil {
			return nil, NewConfigError(fmt.Errorf("Invalid ServiceID : %v : %v", name, err))
		} else {
			serviceIds[i] = ServiceID(id)
		}
	}

	return serviceIds, nil
}

// ServiceConfigID is used to convert the ServiceID into the config id. The config id will be the ServiceID in HEX format.
func ServiceConfigID(id ServiceID) string {
	return fmt.Sprintf("0x%x", id)
}

// ServiceConfigExists returns true if a config exists for the specified ServiceID
func ServiceConfigExists(id ServiceID) bool {
	f, err := os.Open(serviceConfigPath(id))
	if err != nil {
		return false
	}
	f.Close()
	return true
}

// ConfigDir returns the config dir path
func ConfigDir() string {
	return configDir
}

// ConfigDirExists returns true if the config dir exists
func ConfigDirExists() bool {
	dir, err := os.Open(configDir)
	if err != nil {
		return false
	}
	dir.Close()
	return true
}

func serviceConfigPath(id ServiceID) string {
	return fmt.Sprintf("%s/%s", configDir, ServiceConfigID(id))
}

// Config returns the config message for the specified ServiceID.
// The config is expected to exist at {configDir}/{ServiceID}, where the ServiceID is specified in HEX format, e.g.,
//
//	/run/secrets/0xe49214fa20b35ba8
//
// errors:
//	- ServiceConfigNotExistError
//	- ConfigError - for unmarshalling errors
func Config(id ServiceID) (*capnp.Message, error) {
	c, err := ioutil.ReadFile(serviceConfigPath(id))
	if err != nil {
		if err == os.ErrNotExist {
			return nil, NewServiceConfigNotExistError(id)
		}
		switch err := err.(type) {
		case *os.PathError:
			return nil, NewServiceConfigNotExistError(id)
		default:
			return nil, NewConfigError(err)
		}
	}

	msg, err := UnmarshalCapnpMessage(bytes.NewBuffer(c))
	if err != nil {
		return nil, NewConfigError(err)
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
