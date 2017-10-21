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

import "time"

// SystemId system id
type SystemId string

// SystemName system name
type SystemName string

// System owns and manages resources
type System interface {
	// Id system id
	Id() SystemId

	// Name system name must be unique within the scope of the domain
	Name() SystemName

	// DomainId refers to the domain that the system belongs to
	DomainId() DomainId

	// OrgId refers to the organization the system belongs to
	OrgId() OrgId

	// Created when the domain was created
	Created() time.Time

	// Enabled is the system enabled.
	// A system can only be enabled / disabled by a domain owner or an organization owner.
	Enabled() bool

	// Owners who own the system
	Owners() []SubjectId
}
