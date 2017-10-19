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

package config_test

import (
	"testing"

	"bytes"

	"github.com/oysterpack/oysterpack.go/pkg/messaging/nats/server/config"
	"zombiezen.com/go/capnproto2"
)

func TestNewNATSServerConfig(t *testing.T) {
	msg, seg, err := capnp.NewMessage(capnp.SingleSegment(nil))
	if err != nil {
		t.Fatal(err)
	}
	cfg, err := config.NewRootNATSServerConfig(seg)
	cfg.SetClusterName("oysterpack")

	buf := new(bytes.Buffer)
	encoder := capnp.NewEncoder(buf)
	if err := encoder.Encode(msg); err != nil {
		t.Error(err)
	}
	t.Logf("encoded buf len  = %d", len(buf.Bytes()))

	decoder := capnp.NewDecoder(buf)
	msg2, err := decoder.Decode()
	if err != nil {
		t.Fatal(err)
	}
	cfg2, err := config.ReadRootNATSServerConfig(msg2)
	if err != nil {
		t.Fatal(err)
	}
	cluster, err := cfg2.ClusterName()
	if err != nil {
		t.Fatal(err)
	}
	cluster2, err := cfg2.ClusterName()
	if err != nil {
		t.Fatal(err)
	}
	if cluster != cluster2 {
		t.Errorf("cluster did not match : %s != %s", cluster, cluster2)
	}

	buf.Reset()
	packedEncoder := capnp.NewPackedEncoder(buf)
	if err := packedEncoder.Encode(msg); err != nil {
		t.Error(err)
	}
	t.Logf("packed encoded buf len  = %d", len(buf.Bytes()))

	packedDecoder := capnp.NewPackedDecoder(buf)
	msg3, err := packedDecoder.Decode()
	if err != nil {
		t.Fatal(err)
	}
	cfg3, err := config.ReadRootNATSServerConfig(msg3)
	if err != nil {
		t.Fatal(err)
	}
	cluster3, err := cfg3.ClusterName()
	if err != nil {
		t.Fatal(err)
	}
	if cluster3 != cluster2 {
		t.Errorf("cluster did not match : %s != %s", cluster, cluster3)
	}

}
