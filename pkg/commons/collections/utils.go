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

package collections

// StringMapsAreEqual returns true of the 2 maps of strings are the same
func StringMapsAreEqual(m1, m2 map[string]string) bool {
	if len(m1) != len(m2) {
		return false
	}

	for k, v := range m1 {
		if m2[k] != v {
			return false
		}
	}

	return true
}

// Float64MapsAreEqual returns true of the 2 maps of strings are the same
func Float64MapsAreEqual(m1, m2 map[float64]float64) bool {
	if len(m1) != len(m2) {
		return false
	}

	for k, v := range m1 {
		if m2[k] != v {
			return false
		}
	}

	return true
}

// StringSlicesAreEqual returns true if the slice contents are the same
func StringSlicesAreEqual(s1, s2 []string) bool {
	if len(s1) != len(s2) {
		return false
	}

	for i, v := range s1 {
		if s2[i] != v {
			return false
		}
	}

	return true
}

// Float64SlicesAreEqual returns true if the slice contents are the same
func Float64SlicesAreEqual(s1, s2 []float64) bool {
	if len(s1) != len(s2) {
		return false
	}

	for i, v := range s1 {
		if s2[i] != v {
			return false
		}
	}

	return true
}
