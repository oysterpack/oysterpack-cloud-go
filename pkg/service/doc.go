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

// Package service is used to design an application as a set of services. Think of an application merely as a service
// container. The application functionality is defined by its services.
//
// Service Design Goals:
//
//  1. Services are defined by an interface.
//  2. Services must be versioned.
//  3. Services must be easy to monitor. Services must expose health checks and metrics. Metrics must be published to prometheus.
//     Alarms should be configurable and driven by metrics.
//  4. Service logging must be standardized per service. The service log events must be documented.
//  5. Logging must be efficient.
//  6. Services have a standard life cyle (init -> run -> destroy) that is manageable. Service state transitions can be observed.
//  7. Services must define their dependencies to other services.
//  8. Services must be designed to be concurrency safe. The recommended approach is to design services as message
//     processors leveraging go channels and goroutines.
//  9. Services must be easy to configure. Service configuration management must be versioned and managed outside the application.
//     The service configuration must be defined by a schema. Configuration also includes secrets.
//  10. Clients can invoke a service through its interface locally or remotely. The client service interface proxy should
//      be easily configured for local or remote access.
//  11. Services must be secured for remote access.
//  12. Service issues, i.e., bugs, must be tracked.
//  13. Distributed service requests must be traceable
//  14. Services can publish events. The service events must be defined. The events should be made available as a stream.
//  15. Services expose a DevOps interface, which must be securely remotely accessible.
//  16. The deployment topology must be discoverable. Services should register with a centralized service registry.
//  17. Services can have dashboards - Grafana or Prometheus template based dashboards.
//
// Additional Application Design Considerations and Goals:
//
//  1. Applications must be deployable as containers
//  2. Applications must be deployable via kubernetes.
//  3. Applications deployments should be defined as Helm charts.
//  4. Applications must expose readiness and liveliness probes for kubernetes.
//  5. Applications must define the resources they need, i.e., cpu, memory, disk.
//     Each service must define its resource needs - the application resources are then the sum of the service resources.
//     External resources must also be defined, i.e., database, queues, etc.
//
// *** Services must provide documentation for all of the above.
//
// The bottom line, is that it takes great discipline to create high quality application services.
//
// Key Interfaces
//
// 	Service
// 	Client
// 	Application
//
// Key Functions
//
//	App() Application
//
// All exported functions and methods are safe to be used concurrently unless specified otherwise.
package service
