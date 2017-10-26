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

	"io/ioutil"

	"fmt"

	"compress/gzip"

	"crypto/tls"

	"compress/zlib"

	"github.com/oysterpack/oysterpack.go/pkg/messaging/nats/server/config"
	"github.com/pierrec/lz4"
	"zombiezen.com/go/capnproto2"
)

/*
Summary : Comparing compression algorithms:
----------------------
no compression 	: 6368
gzip			: 3968 -> 1.6048 (62.31%)
zlib			: 3956 -> 1.6097 (62.12%)
lz4				: 5345 -> 1.1913 (83.93%)

- zlib gave the best compression ratio, but about the same as gzip
- lz4 is suppose to be much faster

TODO: benchmark compression comparing gzip, zlib, lz4 for compression speed

*/

func TestNewNATSServerConfig(t *testing.T) {
	msg, seg, err := capnp.NewMessage(capnp.SingleSegment(nil))
	if err != nil {
		t.Fatal(err)
	}
	cfg, err := config.NewRootNATSServerConfig(seg)
	cfg.SetClusterName("oysterpack")

	routes, err := capnp.NewTextList(seg, 1)
	checkError(t, err)
	routes.Set(0, "tls://localhost:5222")
	cfg.SetRoutes(routes)
	cfg.SetClusterX509KeyPair(readCertKeyPairFiles(t, "cluster.nats.dev.oysterpack.com", seg))
	cfg.SetServerX509KeyPair(readCertKeyPairFiles(t, "nats.dev.oysterpack.com", seg))

	_, serverHostPortSeg, err := capnp.NewMessage(capnp.SingleSegment(nil))
	serverHostPort, err := config.NewNATSServerConfig_HostPort(serverHostPortSeg)
	serverHostPort.SetHost("node-1")
	serverHostPort.SetPort(8080)
	cfg.SetServer(serverHostPort)

	buf := new(bytes.Buffer)
	encoder := capnp.NewEncoder(buf)
	if err := encoder.Encode(msg); err != nil {
		t.Fatal(err)
	}
	t.Logf("encoded buf len  = %d", len(buf.Bytes()))

	decoder := capnp.NewDecoder(buf)
	msg2, err := decoder.Decode()
	checkError(t, err)

	cfg2, err := config.ReadRootNATSServerConfig(msg2)
	checkError(t, err)

	cluster, err := cfg2.ClusterName()
	checkError(t, err)

	cluster2, err := cfg2.ClusterName()
	checkError(t, err)

	serverHostPort2, err := cfg2.Server()
	checkError(t, err)
	serverHost, err := serverHostPort2.Host()
	checkError(t, err)
	if serverHost != "node-1" {
		t.Errorf("Host did not match : %v", serverHostPort2)
	}
	if serverHostPort2.Port() != serverHostPort.Port() {
		t.Errorf("Port did not match : %v", serverHostPort2)
	}

	if cluster != cluster2 {
		t.Errorf("cluster did not match : %s != %s", cluster, cluster2)
	}

	buf.Reset()
	packedEncoder := capnp.NewPackedEncoder(buf)
	if err := packedEncoder.Encode(msg); err != nil {
		t.Fatal(err)
	}
	t.Logf("packed encoded buf len  = %d", len(buf.Bytes()))

	packedDecoder := capnp.NewPackedDecoder(buf)

	msg3, err := packedDecoder.Decode()
	checkError(t, err)

	cfg3, err := config.ReadRootNATSServerConfig(msg3)
	checkError(t, err)

	cluster3, err := cfg3.ClusterName()
	checkError(t, err)

	if cluster3 != cluster2 {
		t.Errorf("cluster did not match : %s != %s", cluster, cluster3)
	}
}

func TestNewNATSServerConfig_GZIP(t *testing.T) {
	msg, seg, err := capnp.NewMessage(capnp.SingleSegment(nil))
	if err != nil {
		t.Fatal(err)
	}
	cfg, err := config.NewRootNATSServerConfig(seg)
	cfg.SetClusterName("oysterpack")

	routes, err := capnp.NewTextList(seg, 1)
	checkError(t, err)
	routes.Set(0, "tls://localhost:5222")
	cfg.SetRoutes(routes)
	cfg.SetClusterX509KeyPair(readCertKeyPairFiles(t, "cluster.nats.dev.oysterpack.com", seg))
	cfg.SetServerX509KeyPair(readCertKeyPairFiles(t, "nats.dev.oysterpack.com", seg))

	buf := new(bytes.Buffer)
	gzipWriter := gzip.NewWriter(buf)
	encoder := capnp.NewEncoder(gzipWriter)
	if err := encoder.Encode(msg); err != nil {
		t.Fatal(err)
	}
	gzipWriter.Close()
	t.Logf("encoded buf len  = %d", len(buf.Bytes()))

	gzipReader, _ := gzip.NewReader(buf)
	decoder := capnp.NewDecoder(gzipReader)
	msg2, err := decoder.Decode()
	checkError(t, err)
	gzipReader.Close()

	cfg2, err := config.ReadRootNATSServerConfig(msg2)
	checkError(t, err)

	cluster, err := cfg2.ClusterName()
	checkError(t, err)

	cluster2, err := cfg2.ClusterName()
	checkError(t, err)

	if cluster != cluster2 {
		t.Errorf("cluster did not match : %s != %s", cluster, cluster2)
	}

	serverX509KeyPair, err := cfg2.ServerX509KeyPair()
	checkError(t, err)
	certPEMBlock, err := serverX509KeyPair.CertPEMBlock()
	checkError(t, err)
	keyPEMBlock, err := serverX509KeyPair.KeyPEMBlock()
	checkError(t, err)
	_, err = tls.X509KeyPair(certPEMBlock, keyPEMBlock)
	checkError(t, err)

	buf.Reset()
	gzipWriter = gzip.NewWriter(buf)
	packedEncoder := capnp.NewPackedEncoder(gzipWriter)
	if err := packedEncoder.Encode(msg); err != nil {
		t.Fatal(err)
	}
	gzipWriter.Close()
	t.Logf("packed encoded buf len  = %d", len(buf.Bytes()))

	gzipReader, _ = gzip.NewReader(buf)
	packedDecoder := capnp.NewPackedDecoder(gzipReader)

	msg3, err := packedDecoder.Decode()
	checkError(t, err)

	cfg3, err := config.ReadRootNATSServerConfig(msg3)
	checkError(t, err)

	cluster3, err := cfg3.ClusterName()
	checkError(t, err)

	if cluster3 != cluster2 {
		t.Errorf("cluster did not match : %s != %s", cluster, cluster3)
	}
}

func TestNewNATSServerConfig_ZLIB(t *testing.T) {
	msg, seg, err := capnp.NewMessage(capnp.SingleSegment(nil))
	if err != nil {
		t.Fatal(err)
	}
	cfg, err := config.NewRootNATSServerConfig(seg)
	cfg.SetClusterName("oysterpack")

	routes, err := capnp.NewTextList(seg, 1)
	checkError(t, err)
	routes.Set(0, "tls://localhost:5222")
	cfg.SetRoutes(routes)
	cfg.SetClusterX509KeyPair(readCertKeyPairFiles(t, "cluster.nats.dev.oysterpack.com", seg))
	cfg.SetServerX509KeyPair(readCertKeyPairFiles(t, "nats.dev.oysterpack.com", seg))

	buf := new(bytes.Buffer)
	compressor := zlib.NewWriter(buf)
	encoder := capnp.NewEncoder(compressor)
	if err := encoder.Encode(msg); err != nil {
		t.Fatal(err)
	}
	compressor.Close()
	t.Logf("encoded buf len  = %d", len(buf.Bytes()))

	decompressor, _ := zlib.NewReader(buf)
	decoder := capnp.NewDecoder(decompressor)
	msg2, err := decoder.Decode()
	checkError(t, err)
	decompressor.Close()

	cfg2, err := config.ReadRootNATSServerConfig(msg2)
	checkError(t, err)

	cluster, err := cfg2.ClusterName()
	checkError(t, err)

	cluster2, err := cfg2.ClusterName()
	checkError(t, err)

	if cluster != cluster2 {
		t.Errorf("cluster did not match : %s != %s", cluster, cluster2)
	}

	serverX509KeyPair, err := cfg2.ServerX509KeyPair()
	checkError(t, err)
	certPEMBlock, err := serverX509KeyPair.CertPEMBlock()
	checkError(t, err)
	keyPEMBlock, err := serverX509KeyPair.KeyPEMBlock()
	checkError(t, err)
	_, err = tls.X509KeyPair(certPEMBlock, keyPEMBlock)
	checkError(t, err)

	buf.Reset()
	compressor = zlib.NewWriter(buf)
	packedEncoder := capnp.NewPackedEncoder(compressor)
	if err := packedEncoder.Encode(msg); err != nil {
		t.Fatal(err)
	}
	compressor.Close()
	t.Logf("packed encoded buf len  = %d", len(buf.Bytes()))

	decompressor, _ = zlib.NewReader(buf)
	packedDecoder := capnp.NewPackedDecoder(decompressor)

	msg3, err := packedDecoder.Decode()
	checkError(t, err)

	cfg3, err := config.ReadRootNATSServerConfig(msg3)
	checkError(t, err)

	cluster3, err := cfg3.ClusterName()
	checkError(t, err)

	if cluster3 != cluster2 {
		t.Errorf("cluster did not match : %s != %s", cluster, cluster3)
	}
}

func TestNewNATSServerConfig_LZ4(t *testing.T) {
	msg, seg, err := capnp.NewMessage(capnp.SingleSegment(nil))
	if err != nil {
		t.Fatal(err)
	}
	cfg, err := config.NewRootNATSServerConfig(seg)
	cfg.SetClusterName("oysterpack")

	routes, err := capnp.NewTextList(seg, 1)
	checkError(t, err)
	routes.Set(0, "tls://localhost:5222")
	cfg.SetRoutes(routes)
	cfg.SetClusterX509KeyPair(readCertKeyPairFiles(t, "cluster.nats.dev.oysterpack.com", seg))
	cfg.SetServerX509KeyPair(readCertKeyPairFiles(t, "nats.dev.oysterpack.com", seg))

	buf := new(bytes.Buffer)
	compressor := lz4.NewWriter(buf)
	encoder := capnp.NewEncoder(compressor)
	if err := encoder.Encode(msg); err != nil {
		t.Fatal(err)
	}
	compressor.Close()
	t.Logf("encoded buf len  = %d", len(buf.Bytes()))

	decoder := capnp.NewDecoder(lz4.NewReader(buf))
	msg2, err := decoder.Decode()
	checkError(t, err)

	cfg2, err := config.ReadRootNATSServerConfig(msg2)
	checkError(t, err)

	cluster, err := cfg2.ClusterName()
	checkError(t, err)

	cluster2, err := cfg2.ClusterName()
	checkError(t, err)

	if cluster != cluster2 {
		t.Errorf("cluster did not match : %s != %s", cluster, cluster2)
	}

	serverX509KeyPair, err := cfg2.ServerX509KeyPair()
	checkError(t, err)
	certPEMBlock, err := serverX509KeyPair.CertPEMBlock()
	checkError(t, err)
	keyPEMBlock, err := serverX509KeyPair.KeyPEMBlock()
	checkError(t, err)
	_, err = tls.X509KeyPair(certPEMBlock, keyPEMBlock)
	checkError(t, err)

	buf.Reset()
	compressor = lz4.NewWriter(buf)
	packedEncoder := capnp.NewPackedEncoder(compressor)
	if err := packedEncoder.Encode(msg); err != nil {
		t.Fatal(err)
	}
	compressor.Close()
	t.Logf("packed encoded buf len  = %d", len(buf.Bytes()))

	packedDecoder := capnp.NewPackedDecoder(lz4.NewReader(buf))

	msg3, err := packedDecoder.Decode()
	checkError(t, err)

	cfg3, err := config.ReadRootNATSServerConfig(msg3)
	checkError(t, err)

	cluster3, err := cfg3.ClusterName()
	checkError(t, err)

	if cluster3 != cluster2 {
		t.Errorf("cluster did not match : %s != %s", cluster, cluster3)
	}
}

func TestNewRootNATSServerConfig_MultiMessages(t *testing.T) {
	buf := new(bytes.Buffer)
	compressor := zlib.NewWriter(buf)
	encoder := capnp.NewEncoder(compressor)
	func() {
		defer compressor.Close()
		for i := 0; i < 10; i++ {
			msg, seg, err := capnp.NewMessage(capnp.SingleSegment(nil))
			if err != nil {
				t.Fatal(err)
			}
			cfg, err := config.NewRootNATSServerConfig(seg)
			cfg.SetClusterName(fmt.Sprintf("oysterpack-%d", i))
			encoder.Encode(msg)
		}
	}()

	reader, _ := zlib.NewReader(buf)
	decoder := capnp.NewDecoder(reader)
	func() {
		defer reader.Close()
		for i := 0; i < 10; i++ {
			msg, err := decoder.Decode()
			checkError(t, err)
			cfg, err := config.ReadRootNATSServerConfig(msg)
			checkError(t, err)
			cluster, _ := cfg.ClusterName()
			t.Log(cluster)
		}
	}()
}

func readCertKeyPairFiles(t *testing.T, name string, seg *capnp.Segment) config.NATSServerConfig_X509KeyPair {
	t.Helper()
	const PKI_ROOT = "../testdata/.easypki/pki/oysterpack"

	keypair, err := config.NewNATSServerConfig_X509KeyPair(seg)
	checkError(t, err)

	cert, err := ioutil.ReadFile(fmt.Sprintf("%s/certs/%s.crt", PKI_ROOT, name))
	checkError(t, err)

	key, err := ioutil.ReadFile(fmt.Sprintf("%s/keys/%s.key", PKI_ROOT, name))
	checkError(t, err)

	keypair.SetCertPEMBlock(cert)
	keypair.SetKeyPEMBlock(key)

	return keypair
}

func checkError(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatal(err)
	}
}
