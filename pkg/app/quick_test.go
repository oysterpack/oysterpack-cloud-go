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

package app_test

import (
	"fmt"
	"testing"
	"time"
)

type HexID uint64

func (a HexID) String() string { return fmt.Sprintf("%x", uint64(a)) }

func TestHexFormat(t *testing.T) {

	t.Logf("%%v = %v", HexID(0xe1b1125eac639831))
	t.Logf("%%s = %s", HexID(0xe1b1125eac639831))
	t.Log(HexID(0xe1b1125eac639831).String())
	t.Log(uint64(HexID(0xe1b1125eac639831)))

	// I would expect this to format the number as HEX, but instead some unexpected number gets printed
	t.Logf("%%x = %x", HexID(0xe1b1125eac639831))

	// TEST OUTPUT
	// quick_test.go:28: %v = e1b1125eac639831
	// quick_test.go:29: %s = e1b1125eac639831
	// quick_test.go:30: e1b1125eac639831
	// quick_test.go:31: 16262799927240005681
	// quick_test.go:34: %x = 65316231313235656163363339383331
}

func TestPointers(t *testing.T) {
	s := "HELLO"
	var a *string = &s
	var b *string = a

	t.Log(*a, *b, s)
	s = "CIAO"
	t.Log(*a, *b, s)
}

func TestDurationZero(t *testing.T) {
	d := time.Duration(0)
	if d != 0 {
		t.Error("duration should be zero")
	}

	d = time.Second
	if !(d > 0) {
		t.Error("duration should be > 0")
	}
}
