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

package domain

import (
	"time"

	"github.com/Masterminds/semver"
)

// ResourceId Resource id
type ResourceId string

// ResourceName Resource name
type ResourceName string

// Resource represents any system resource. For example, it could represent a service, database, queue, etc.
type Resource interface {
	// Id Resource id
	Id() ResourceId

	// Name Resource name must be unique within the scope of the system
	Name() ResourceName

	// Version is the resource version. The purpose is to ensure the correct version is deployed.
	Version() *semver.Version

	// SystemId refers to the system that the resource belongs to
	SystemId() SystemId

	// DomainId refers to the domain that the Resource belongs to
	DomainId() DomainId

	// OrgId refers to the organization the Resource belongs to
	OrgId() OrgId

	// Created when the domain was created
	Created() time.Time

	// Enabled is the Resource enabled.
	// A Resource can only be enabled / disabled by a domain owner or an organization owner.
	Enabled() bool

	// Owners who own the Resource
	Owners() []SubjectId
}
