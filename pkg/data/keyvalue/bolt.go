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

package keyvalue

import "github.com/coreos/bbolt"

func lookupBucket(tx *bolt.Tx, path []string) *bolt.Bucket {
	return lookupChildBucket(tx, nil, path)
}

func lookupChildBucket(tx *bolt.Tx, parent *bolt.Bucket, path []string) *bolt.Bucket {
	if len(path) == 0 {
		return parent
	}

	if parent != nil {
		parent = parent.Bucket([]byte(path[0]))
	} else {
		parent = tx.Bucket([]byte(path[0]))
	}
	if parent == nil {
		return nil
	}
	return lookupChildBucket(tx, parent, path[1:])
}
