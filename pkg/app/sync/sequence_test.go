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
	"sync"
	"testing"

	opsync "github.com/oysterpack/oysterpack.go/pkg/app/sync"
)

func TestNewSequence(t *testing.T) {

	t.Run("NewSequence(0)", func(t *testing.T) {
		seq := opsync.NewSequence(0)

		wg := sync.WaitGroup{}

		for i := 0; i < 10; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for i := 0; i < 1000; i++ {
					seq.Next()
				}
			}()
		}
		wg.Wait()
		if seq.Value() != 10*1000 {
			t.Errorf("Seq value did not match : %d", seq.Value())
		}
	})

	t.Run("NewSequence(100)", func(t *testing.T) {
		seq := opsync.NewSequence(100)

		wg := sync.WaitGroup{}

		for i := 0; i < 10; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for i := 0; i < 1000; i++ {
					seq.Next()
				}
			}()
		}
		wg.Wait()
		if seq.Value() != 10*1000+100 {
			t.Errorf("Seq value did not match : %d", seq.Value())
		}
	})

	t.Run("Default Sequence", func(t *testing.T) {
		var seq opsync.Sequence

		wg := sync.WaitGroup{}

		for i := 0; i < 10; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for i := 0; i < 1000; i++ {
					seq.Next()
				}
			}()
		}
		wg.Wait()
		if seq.Value() != 10*1000 {
			t.Errorf("Seq value did not match : %d", seq.Value())
		}
	})

}
