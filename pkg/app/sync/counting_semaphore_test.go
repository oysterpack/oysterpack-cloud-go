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

package sync_test

import (
	"testing"

	"github.com/oysterpack/oysterpack.go/pkg/app/sync"
)

func TestNewCountingSemaphore(t *testing.T) {

	t.Run("Token count = 0", func(t *testing.T) {
		defer func() {
			if p := recover(); p == nil {
				t.Error("expected a panic because it is a bug to create a CountingSemaphore with 0 tokens")
			}
		}()

		sync.NewCountingSemaphore(0)
	})

	t.Run("10 tokens", func(t *testing.T) {
		const SIZE = 10
		semaphore := sync.NewCountingSemaphore(SIZE)
		if semaphore.AvailableTokens() != SIZE {
			t.Errorf("AvailableTokens did not match : %d", semaphore.AvailableTokens())
		}

		if semaphore.TotalTokens() != SIZE {
			t.Errorf("TotalTokens did not match : %d", semaphore.AvailableTokens())
		}

		for i := SIZE; i > 0; i-- {
			if semaphore.AvailableTokens() != i {
				t.Errorf("AvailableTokens did not match : %d != %i", semaphore.AvailableTokens(), i)
			}
			<-semaphore.C
		}

		if semaphore.AvailableTokens() != 0 {
			t.Errorf("There should not be any tokens available : %d", semaphore.AvailableTokens())
		}

		select {
		case <-semaphore.C:
			t.Error("There should be no tokens available")
		default:
		}

		semaphore.ReturnToken()
		if semaphore.AvailableTokens() != 1 {
			t.Errorf("There should not be any tokens available : %d", semaphore.AvailableTokens())
		}
		select {
		case <-semaphore.C:
		default:
			t.Error("There should be 1 token available")
		}

		for i := 0; i < SIZE*2; i++ {
			semaphore.ReturnToken()
		}

		if semaphore.AvailableTokens() != SIZE {
			t.Errorf("There should be %d available tokens, but %d was reported", SIZE, semaphore.AvailableTokens())
		}

	})
}
