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

package uid

import (
	"hash/fnv"

	"github.com/nats-io/nuid"
)

type UID string

func (a UID) Hash() UIDHash {
	hasher := fnv.New64()
	hasher.Write([]byte(a))
	return UIDHash(hasher.Sum64())
}

type UIDHash uint64

func (a UIDHash) UInt64() uint64 {
	return uint64(a)
}

type UIDHashProducer func() UIDHash

func NextUID() UID {
	return UID(nuid.Next())
}

func NextUIDHash() UIDHash {
	return NextUID().Hash()
}
