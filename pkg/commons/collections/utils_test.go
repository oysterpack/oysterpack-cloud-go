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

package collections_test

import (
	"testing"

	"github.com/oysterpack/oysterpack.go/pkg/commons/collections"
)

func TestStringMapEquals(t *testing.T) {
	m1 := map[string]string{}
	m2 := map[string]string{}

	if !collections.StringMapEquals(m1, m2) {
		t.Errorf("both are empty maps, and thus should be equal")
	}

	m1["a"] = "a"
	if collections.StringMapEquals(m1, m2) {
		t.Errorf("maps should not be equal")
	}

	m2["a"] = "a"
	if !collections.StringMapEquals(m1, m2) {
		t.Errorf("both maps should be equal")
	}

	m2["a"] = "b"
	if collections.StringMapEquals(m1, m2) {
		t.Errorf("both maps should not be equal")
	}

	m1["a"] = "b"
	if !collections.StringMapEquals(m1, m2) {
		t.Errorf("maps should be equal")
	}

	m1["b"] = "b"
	if collections.StringMapEquals(m1, m2) {
		t.Errorf("maps should not be equal")
	}
}
