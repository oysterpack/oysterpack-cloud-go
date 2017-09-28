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

// Package messaging provides the messaging interfaces designed from the application's perspective
// The goal is to design a distributed message bus with the following goal :
//
// 		The network should be transparent to applications. When network and application problems do occur it should be
// 		easy to determine the source of the problem.
//
// In practice, achieving the previously stated goal is incredibly difficult.
// Messaging attempts to do so by providing the following high level features:
//
// 1. Out of process architecture : The messaging process is designed to run alongside every application server.
//    The message processes form a distributed messaging cluster that will serve as the "nervous system" for high volume messaging
//    systems and resiliency and high availability. Neing out of process enables the messaging infrastructure to be upgraded
//    independent of the application.
// 2. Clients are cluster-aware.
package messaging
