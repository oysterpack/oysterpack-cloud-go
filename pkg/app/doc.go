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
//  1. High performance. Prove it through benchmarking.
//  2. Concurrency safe. Always design the service to be safely used by concurrent goroutines.
//     - Prefer immutability
//	   - Prefer channels
//	   - use locks as a last resort
//  3. Services define their dependencies as functions.
//  4. One service per package, i.e., the package is the service. Public functions exposed by the package are the service interface.
//  5. Each service package will initialize itself automatically, i.e., via package init() functions. Each service package will
//	   register itself with the app.Application service.
//  6. Graceful application shutdown : when the application is killed, all registered services will be killed. The application will wait
//     until all services are dead before exiting.
//     - The following OS signals will trigger graceful application shutdown : SIGINT, SIGTERM, SIGQUIT
//  7. All key components will be assigned a unique numeric id (uint64) for tracking and traceability purposes.
//	   - DomainID
//	   - AppID
//	   - ReleaseID - app release id
//	   - ServiceID
//	   - ErrorID
//	   - LogEventID
//
// 	   Most systems prefer instead to define a symbolic global namespace , but this would have some important disadvantages:
//		1. Programmers often feel the need to change symbolic names and organization in order to make their code cleaner,
//         but the renamed code should still work with existing encoded data.
//		2. It’s easy for symbolic names to collide, and these collisions could be hard to detect in a large distributed
//         system with many different binaries.
//		3. Fully-qualified type names may be large and waste space when transmitted on the wire.
//
//	8. Config is typesafe and secure.
// 	   - Each service will define its config via capnp. The config will be self contained.
//	   - Configurations will be encrypted.
//	   - Docker secrets (https://docs.docker.com/engine/swarm/secrets/) will be supported
//	9. RPC network communication will be secured via mutual TLS
//     - capnp RPC will be supported
package app
