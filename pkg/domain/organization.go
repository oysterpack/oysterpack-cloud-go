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
	"crypto/tls"
	"time"
)

// OrgId Orgaization ID
type OrgId string

// OrgName Organization name
type OrgName string

// Organization owns and manages domains and applications
type Organization interface {
	//Id org id
	Id() OrgId

	// Name org name is globally unique
	Name() OrgName

	// Created when the organization was created
	Created() time.Time

	// Enabled is the organization enabled
	Enabled() bool

	// Owners who own the organization
	// Owners have full control over the entire organization.
	Owners() []SubjectId

	// CACert is the organization CA Certificate that is used to create all certificates for the organization
	CACert() tls.Certificate
}
