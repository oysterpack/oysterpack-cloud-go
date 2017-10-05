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

package pkg_test

import (
	"testing"

	"regexp"

	"github.com/json-iterator/go"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

type Name string

type Foo struct {
	Name
}

var clusterRegex = regexp.MustCompile(`^[a-z][0-9a-z_\-]+$`)

func TestRegex(t *testing.T) {
	t.Logf("%v", clusterRegex)

	if !clusterRegex.MatchString("oysterpack-nats") {
		t.Error("match failed")
	}

	if clusterRegex.MatchString("oysterpack-nats:") {
		t.Error("should not have matched")
	}
}

func TestMarshalCustomAlias(t *testing.T) {

	foo := Foo{Name: "Alfio Zappala"}

	bytes, err := json.Marshal(&foo)
	if err != nil {
		t.Fatalf("Marshal failed : %v", err)
	}

	foo2 := &Foo{}
	if err := json.Unmarshal(bytes, foo2); err != nil {
		t.Fatalf("Unmarshal failed : %v", err)
	}

	t.Logf("foo2 : %v", foo2)

	if foo.Name != foo2.Name {
		t.Errorf("Marshal / Unmarshal did not work : [%v] != [%v]", foo, foo2)
	}

}

func TestStructKey(t *testing.T) {
	m := map[Key]int{
		Key{"a", "b"}: 1,
		Key{"a", "b"}: 2,
		Key{"a", "z"}: 3,
	}
	t.Logf("%v", m)
	m[Key{"a", "b"}] = 4
	t.Logf("%v", m)
}

type Key struct {
	topic string
	queue string
}
