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

import "github.com/Masterminds/semver"

// ActionId Action id
type ActionId string

// ActionName Action name
type ActionName string

// Action represents any system Action. For example, it could represent a service, database, queue, etc.
type Action interface {
	// Id Action id
	Id() ActionId

	// Name Action name must be unique within the scope of the resource
	Name() ActionName

	// ResourceId resource id
	ResourceId() ResourceId

	// SystemId refers to the system that the Action belongs to
	SystemId() SystemId

	// DomainId refers to the domain that the Action belongs to
	DomainId() DomainId

	// OrgId refers to the organization the Action belongs to
	OrgId() OrgId

	// Enabled is the Action enabled.
	// An Action can only be enabled / disabled by an owner in the organization hierarchy.
	Enabled() bool

	// MessageDescs defines the types of request-response messages that are supported by this action
	MessageDescs() []MessageDesc

	// Version is the resource action version. The purpose is to ensure the correct version is deployed.
	Version() *semver.Version
}

type MessageDesc interface {
	RequestMessageDesc() RequestMessageDesc
	ResponseMessageDesc() ResponseMessageDesc
}

type RequestMessageDesc interface {
	Type() MessageType
	Streaming() bool
}

type ResponseMessageDesc interface {
	Type() MessageType
	Streaming() bool

	// StatusCodes returns the possible status codes that can be returned on a response.
	StatusCodes() []ActionStatusCode
}
