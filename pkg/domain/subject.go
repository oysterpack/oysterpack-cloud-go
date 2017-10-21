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

// SubjectId subject id
type SubjectId string

// Subject subject
type Subject interface {
	// Id subject id
	Id() SubjectId

	// OrgId the organization the subject belongs to
	OrgId() OrgId

	Created() time.Time

	// Enabled can only be changed by an organization owner
	Enabled() bool

	// Certificate which is signed by the organization's CA cert
	Cert() tls.Certificate
}
