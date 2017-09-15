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

package service

// ServiceDependencies provides functionality to get information about about service dependencies
type ServiceDependencies interface {
	// TODO: add healthchecks and alerts for dependency checks

	// CheckAllServiceDependencies checks that all dependencies for each service are available
	CheckAllServiceDependencies() *ServiceDependencyErrors

	// CheckServiceDependencies checks that the service dependencies are available
	CheckServiceDependencies(client ServiceClient) *ServiceDependencyErrors
}
