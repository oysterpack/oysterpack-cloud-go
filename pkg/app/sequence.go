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

import "sync"

// NewSequence creates a new Sequence using n as the current sequence value.
func NewSequence(n uint64) *Sequence {
	return &Sequence{n: n}
}

// Sequence is used to get an incrementing sequence number
type Sequence struct {
	m sync.Mutex
	n uint64
}

// Next returns the next sequence value
func (a *Sequence) Next() uint64 {
	a.m.Lock()
	a.n++
	n := a.n
	a.m.Unlock()
	return n
}

// Value returns the current sequence value
func (a *Sequence) Value() uint64 {
	a.m.Lock()
	n := a.n
	a.m.Unlock()
	return n
}
