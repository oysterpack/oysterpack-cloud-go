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

// Config returns the config message for the specified ServiceID.
// The config is expected to exist at {configDir}/{ServiceID}, where the ServiceID is specified in HEX format, e.g.,
//
//	/run/secrets/0xe49214fa20b35ba8
//
// errors:
//	- ServiceConfigNotExistError
//	- ConfigError - for unmarshalling errors
func Config(id ServiceID) (*capnp.Message, error) {
	c, err := ioutil.ReadFile(fmt.Sprintf("%s/0x%x", configDir, id))
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

func MarshalCapnpMessage(msg *capnp.Message, out io.Writer) error {
	compressor := zlib.NewWriter(out)
	encoder := capnp.NewPackedEncoder(compressor)
	if err := encoder.Encode(msg); err != nil {
		return err
	}
	compressor.Close()
	return nil
}

func UnmarshalCapnpMessage(in io.Reader) (*capnp.Message, error) {
	decompressor, _ := zlib.NewReader(in)
	decoder := capnp.NewPackedDecoder(decompressor)
	msg, err := decoder.Decode()
	if err != nil {
		return nil, err
	}
	return msg, nil
}
