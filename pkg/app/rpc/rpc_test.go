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

package rpc_test

import (
	"net"
	"sync"
	"testing"

	"hash/fnv"

	"github.com/nats-io/nuid"
	"github.com/oysterpack/oysterpack.go/pkg/app/message"
	"github.com/rs/zerolog/log"
	"gopkg.in/tomb.v2"
	"zombiezen.com/go/capnproto2"
)

func TestRPCMessaging(t *testing.T) {
	l, err := net.Listen("tcp", ":0")
	if err != nil {
		t.Fatal(err)
	}
	defer l.Close()

	listenerRunning := sync.WaitGroup{}
	listenerRunning.Add(1)
	listenerTomb := tomb.Tomb{}

	listenerTomb.Go(func() error {
		for {
			select {
			case <-listenerTomb.Dying():
				return nil
			default:
				conn, err := l.Accept()
				if err != nil {
					t.Error(err)
					return err
				}

				listenerTomb.Go(func() error {
					defer conn.Close()

					decoder := capnp.NewPackedDecoder(conn)
					decoder.ReuseBuffer()

					for {
						select {
						case <-listenerTomb.Dying():
							return nil
						default:
							if msg, err := decoder.Decode(); err != nil {
								m, err := message.ReadRootMessage(msg)
								if err != nil {
									log.Error().Err(err).Msg("message.ReadRootMessage(msg) failed")
									return err
								}

								t.Logf("id : %d", m.Id())
							}
						}
					}

				})

			}
		}
	})
}

func BenchmarkHash(b *testing.B) {

	b.Run("uid.Next", func(b *testing.B) {
		uid := nuid.New()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			uid.Next()
		}
	})

	b.Run("nuid.Next", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			nuid.Next()
		}
	})

	hashes := make(map[uint64]struct{}, 1000*1000*3)
	hasher := fnv.New64()
	count := 0
	b.Run("nuid hash - checking for dups", func(b *testing.B) {
		ids := make([]string, b.N)
		for i := 0; i < b.N; i++ {
			ids[i] = nuid.Next()
		}
		count = count + b.N

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			hasher.Reset()
			hasher.Write([]byte(ids[i]))
			hashes[hasher.Sum64()] = struct{}{}
		}
	})

	b.Logf("count = %d, len(hashes) = %d", count, len(hashes))
	if len(hashes) != count {
		b.Errorf("Dups occurrec : %d - %d = %d", count, len(hashes), count-len(hashes))
	}

	b.Run("hash", func(b *testing.B) {
		ids := make([]string, b.N)
		for i := 0; i < b.N; i++ {
			ids[i] = nuid.Next()
		}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			hasher.Reset()
			hasher.Write([]byte(ids[i]))
		}
	})

	b.Run("nuid hash", func(b *testing.B) {
		uid := nuid.New()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			hasher.Reset()
			hasher.Write([]byte(uid.Next()))
		}
	})

}
