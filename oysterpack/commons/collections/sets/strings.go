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

package sets

import (
	"fmt"
	"sync"
)

// Strings represents a set of strings.
// The collection is concurrency safe.
type Strings interface {
	// Values returns the set of strings
	Values() []string

	// Add returns true if the string was added to the set.
	// false is returned if the string already exists in the set.
	Add(s string) bool

	// Remove returns true if the string was removed from the set.
	// false is returned if the string did not exist in the set
	Remove(s string) bool

	// Clear clears the set
	Clear()

	// Size returns the collection size
	Size() int

	// Empty return true if Size() == 0
	Empty() bool

	Contains(s string) bool

	Equals(set Strings) bool

	ContainsAll(set Strings) bool
}

type set struct {
	mutex  sync.RWMutex
	values map[string]struct{}
}

var entry = struct{}{}

func NewStrings() Strings {
	return &set{
		values: make(map[string]struct{}),
	}
}

func (a *set) Values() []string {
	a.mutex.RLock()
	defer a.mutex.RUnlock()
	values := make([]string, len(a.values))
	i := 0
	for k, _ := range a.values {
		values[i] = k
		i++
	}
	return values
}

func (a *set) Add(s string) bool {
	a.mutex.Lock()
	defer a.mutex.Unlock()
	if a.contains(s) {
		return false
	}
	a.values[s] = entry
	return true
}

func (a *set) contains(s string) bool {
	_, exists := a.values[s]
	return exists
}

func (a *set) Contains(s string) bool {
	a.mutex.RLock()
	defer a.mutex.RUnlock()
	return a.contains(s)
}

func (a *set) Remove(s string) bool {
	a.mutex.Lock()
	defer a.mutex.Unlock()
	contained := a.contains(s)
	delete(a.values, s)
	return contained
}

func (a *set) Clear() {
	a.mutex.Lock()
	defer a.mutex.Unlock()
	a.values = make(map[string]struct{})
}

func (a *set) Size() int {
	a.mutex.RLock()
	defer a.mutex.RUnlock()
	return len(a.values)
}

func (a *set) Empty() bool {
	return a.Size() == 0
}

func (a *set) String() string { return fmt.Sprintf("%v", a.Values()) }

func (a *set) ContainsAll(set Strings) bool {
	if a.Size() < set.Size() {
		return false
	}
	a.mutex.RLock()
	defer a.mutex.RUnlock()
	for _, s := range set.Values() {
		if !a.contains(s) {
			return false
		}
	}
	return true
}

func (a *set) Equals(set Strings) bool {
	if a.Size() != set.Size() {
		return false
	}
	a.mutex.RLock()
	defer a.mutex.RUnlock()
	for _, s := range set.Values() {
		if !a.contains(s) {
			return false
		}
	}
	return true
}
