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
)

// DomainID domain id
type DomainId string

// DomainName domain name
type DomainName string

// Domain owns and manages systems
type Domain interface {
	// Id domain id
	Id() DomainId

	// Name domain name must be unique within the scope of the organization
	Name() DomainName

	// OrgId refers to the Organization this domain belongs to
	OrgId() OrgId

	// Created when the domain was created
	Created() time.Time

	// Enabled is the domain enabled.
	// A domain can only be enabled / disabled by an organization owner.
	Enabled() bool

	// Owners who own the domain
	// Domain owners have full control over the domain, except for disabling / enabling the domain.
	Owners() []SubjectId
}
