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

package comp

// Context represents the Container context that is made available to Component(s)
type Context interface {
	Registry

	// Go runs f in a new goroutine and tracks its termination.
	// If f returns a non-nil error, then the container is killed with that error as the death reason parameter.
	//
	// It is f's responsibility to monitor the stopping channel and return appropriately once it is triggered, i.e., the stopping
	// channel is used to notify the goroutine that the container is stopping, signalling the goroutine to terminate.
	//
	// It is safe for the f function to call the Go method again to create additional tracked goroutines.
	//
	// Calling the Go method after the container has stopped causes a runtime panic.
	// For that reason, calling the Go method a second time out of a tracked goroutine is unsafe.
	Go(f func(stopping <-chan struct{}) error)
}
