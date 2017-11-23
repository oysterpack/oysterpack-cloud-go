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

// NewCountingSemaphore returns a new CountingSemaphore with specified resource count
func NewCountingSemaphore(tokenCount uint) CountingSemaphore {
	if tokenCount == 0 {
		tokenCount = 1
	}
	c := make(chan struct{}, tokenCount)
	token := struct{}{}
	for i := 0; i < cap(c); i++ {
		c <- token
	}
	return c
}

// CountingSemaphore which allow an arbitrary resource count
type CountingSemaphore chan struct{}

// ReturnToken returns a token back
func (a CountingSemaphore) ReturnToken() {
	select {
	case a <- struct{}{}:
	default:
	}
}
