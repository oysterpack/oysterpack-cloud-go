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

import (
	"io"
	"runtime"
)

// DumpAllStacks dumps all goroutine stacks to the specified writer.
// Used for troubleshooting purposes.
func DumpAllStacks(out io.Writer) {
	size := 1024 * 8
	for {
		buf := make([]byte, size)
		if len := runtime.Stack(buf, true); len <= size {
			out.Write(buf[:len])
			return
		}
		size = size + (1024 * 8)
	}
}
