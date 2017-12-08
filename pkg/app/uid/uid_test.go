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

package uid_test

import (
	"testing"

	"hash/fnv"

	"github.com/nats-io/nuid"
	"github.com/oysterpack/oysterpack.go/pkg/app/uid"
)

func TestNextUID(t *testing.T) {
	size := 1 * 1000 * 1000
	hashes := make(map[uid.UID]struct{}, size)

	for i := 0; i < size; i++ {
		hashes[uid.NextUID()] = struct{}{}
	}

	if len(hashes) != size {
		t.Errorf("Dups occurrec : %d - %d = %d", size, len(hashes), size-len(hashes))
	}
}

func TestNextUIDHash(t *testing.T) {
	uidHash := uid.NextUIDHash()
	if uint64(uidHash) != uidHash.UInt64() {
		t.Fatal("values do not match : %v != %v", uint64(uidHash), uidHash.UInt64())
	}

	size := 1 * 1000 * 1000
	hashes := make(map[uid.UIDHash]struct{}, size)

	for i := 0; i < size; i++ {
		hashes[uid.NextUIDHash()] = struct{}{}
	}

	if len(hashes) != size {
		t.Errorf("Dups occurrec : %d - %d = %d", size, len(hashes), size-len(hashes))
	}
}

// BenchmarkUID/uidGen.Next()-8            20000000                98.8 ns/op            32 B/op          1 allocs/op
// BenchmarkUID/nuid.Next()-8              10000000               250 ns/op              72 B/op          3 allocs/op
// BenchmarkUID/NextUID-8                  10000000               129 ns/op              32 B/op          1 allocs/op
// BenchmarkUID/NextUIDHash-8               5000000               289 ns/op              72 B/op          3 allocs/op
func BenchmarkUID(b *testing.B) {
	uidGen := nuid.New()
	b.Run("uidGen.Next()", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			uidGen.Next()
		}
	})

	b.Run("nuid.Next()", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			hasher := fnv.New64()
			hasher.Write([]byte(uidGen.Next()))
			hasher.Sum64()
		}
	})

	b.Run("NextUID", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			uid.NextUID()
		}
	})

	b.Run("NextUIDHash", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			uid.NextUIDHash()
		}
	})

}
