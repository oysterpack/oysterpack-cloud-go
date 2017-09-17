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

// Package service provides support to build high quality application services. An application is composed of services.
// Application services must be easy to configure, deploy, monitor, operate, scale. Application services must be designed
// to be cloud native.
//
// High quality application services must address for the following design and operational concerns :
// ==================================================================================================
// - services must be designed to be used concurrently safe
// - interfaces must be used to define a service
//    - the interface defines the service's functionality and must be clearly documented
// - services must have a clearly defined lifecycle, i.e., init -> run -> destroy
// - service configuration management must be versioned and managed outside the application.
// - service configuration should have a clearly defined schema, which is versioned
// - each service must have its own logger, i.e., it must be easy to identify and isolate a service's log events
//    - each service must document its log events
//    - log events should be as structural as possible, enabling it to be easily searched
// - services should provide metrics and be instrumented for monitoring and alerting
// - service Dependencies must be defined
// - services must provide health checks
// - service issues, i.e., bugs, must be tracked
// - services must be scalable
// - services must be versioned - the deployed version must be available at runtime
// - services must be secure
// - distributed service requests must be traceable
//
// - applications must be versioned and retain build information
// - applications must be packaged and deployed as containers
// - applications must be able to be deployed via kubernetes
// - applications must expose readiness and liveliness probes
// - applications must define the resources they need, i.e., cpu, memory, disk
//     - each service must define its resource needs - the application resources are then the sum of the service resources
//     - external resources must also be defined, i.e., database, queues, etc
//
// *** Services must provide documentation for all of the above.
//
// The bottom line, is that it takes great discipline to create high quality application services.
//
// Key interfaces
// ==============
// Service interfaces
// Client interface
// Application interface
//
// Key functions
// =============
// App() Application
//
// All exported functions and methods are safe to be used concurrently unless specified otherwise.
package service
