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

package sets_test

import (
	"fmt"
	"testing"

	"github.com/oysterpack/oysterpack.go/pkg/commons/collections/sets"
)

func TestNewStrings_EmptySet(t *testing.T) {
	s := sets.NewStrings()

	// exercise empty set
	if !s.Empty() || s.Size() != 0 || s.Contains("a") || s.Remove("a") || len(s.Values()) != 0 {
		t.Error("set should be empty")
	}
	s.Clear()
}

func TestStrings_AddRemove(t *testing.T) {
	s := sets.NewStrings()

	for i := 0; i < 10; i++ {
		s.Add(fmt.Sprintf("#%v", i))
	}
	t.Logf("set : %v", s)
	if s.Size() != 10 {
		t.Error("There should be 10 elements in the set")
	}
	for i := 0; i < 10; i++ {
		value := fmt.Sprintf("#%v", i)
		if !s.Contains(value) {
			t.Errorf("set should have contained: %v", value)
		}
		if s.Add(value) {
			t.Errorf("should not have added: %v", value)
		}

		if !s.Remove(value) {
			t.Errorf("should have removed: %v", value)
		}
		if s.Remove(value) {
			t.Errorf("should have been already removed: %v", value)
		}
	}
}

func TestStrings_ContainsAll_Equals(t *testing.T) {
	s := sets.NewStrings()
	for i := 0; i < 10; i++ {
		s.Add(fmt.Sprintf("#%v", i))
	}

	s2 := sets.NewStrings()
	if s2.ContainsAll(s) {
		t.Error("ERROR: s2 is empty")
	}
	if s2.Equals(s) {
		t.Error("ERROR: s2 is empty")
	}
	for i := 0; i < 10; i++ {
		s2.Add(fmt.Sprintf("#%v", i))
		if !s.ContainsAll(s2) {
			t.Errorf("ERROR: s should contain s2 : %v : %v", s, s2)
		}
	}
	if !s2.ContainsAll(s) || !s2.Equals(s) {
		t.Error("ERROR: s2 should == s")
	}
	s2.Add("Z")
	s.Add("A")
	if s2.ContainsAll(s) {
		t.Error("ERROR: s2 should contain oll of s")
	}
	if s.ContainsAll(s2) {
		t.Error("ERROR: s is a subset s2")
	}
	if s2.Equals(s) {
		t.Error("ERROR: s2 should != s")
	}
}
