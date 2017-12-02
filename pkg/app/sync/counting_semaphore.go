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

package sync

// NewCountingSemaphore returns a new CountingSemaphore with specified resource count.
// A panic is triggered if the token count == 0.
func NewCountingSemaphore(tokenCount uint) *CountingSemaphore {
	if tokenCount == 0 {
		panic("token count must be > 0")
	}
	c := make(chan struct{}, tokenCount)
	token := struct{}{}
	for i := 0; i < cap(c); i++ {
		c <- token
	}
	return &CountingSemaphore{c, func() {
		select {
		case c <- struct{}{}:
		default:
		}
	}}
}

// CountingSemaphore which allow an arbitrary resource count.
// Tokens are simply pulled from the channel. Make sure tokens are returned back to the pool via ReturnToken()
type CountingSemaphore struct {
	C           <-chan struct{}
	ReturnToken func()
}

// TotalTokens returns the total number of tokens
func (a CountingSemaphore) TotalTokens() int {
	return cap(a.C)
}

// AvailableTokens returns the number of tokens that are available
func (a CountingSemaphore) AvailableTokens() int {
	return len(a.C)
}
