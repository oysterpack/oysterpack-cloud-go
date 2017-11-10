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

// Package app provides the application service model and platform for building message driven concurrent and distributed
// services.
//
// Design Principles
//	1. High performance. Prove it through benchmarking.
//  2. Concurrency safe. Always design the service to be safely used by concurrent goroutines.
//     - Prefer immutability
//	   - Prefer channels
//	   - use locks as a last resort
//	3. Services define their dependencies as functions.
//  4. One service per package, i.e., the package is the service. Public functions exposed by the package are the service interface.
//	5. Each service package will initialize itself automatically, i.e., via package init() functions. Each service package will
//	   register itself with the app.Application service.
//  6. When Application shutdown is triggered, services that need to gracefully shutdown should listen on the Application Stopping channel.
//  7. The Application has a lifecycle. Each lifecycle state has a channel that can be used to listen for lifecyle events.
//  8. All services that must shutdown gracefully, must launch a goroutine through the Application. When the Application shutdown
//     is triggered, the Application will not transition to the Dead state until all registered goroutines have completed.
//  9. All key components will be assigned a unique numeric id (uint64) for tracking purposes.
//
// 	   Most systems prefer instead to define a symbolic global namespace , but this would have some important disadvantages:
//		1. Programmers often feel the need to change symbolic names and organization in order to make their code cleaner,
//         but the renamed code should still work with existing encoded data.
//		2. It’s easy for symbolic names to collide, and these collisions could be hard to detect in a large distributed
//         system with many different binaries.
//		3. Fully-qualified type names may be large and waste space when transmitted on the wire.
package app
