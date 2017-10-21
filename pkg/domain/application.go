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

	"github.com/Masterminds/semver"
)

// ApplicationId application id
type ApplicationId string

// ApplicationName must be unique within the owning organization
type ApplicationName string

// Application is used to package resource actions as an application.
// All who have access to the application have access to all of the app's resource actions.
type Application interface {
	// Id is the app id
	Id() ApplicationId

	// Name is the app name
	Name() ApplicationName

	// Version is the application version. The purpose is to ensure the correct version is deployed.
	Version() *semver.Version

	// OrgId is the owning organization
	OrgId() OrgId

	// CACert is an intermediate CA cert that is signed by the organization's CA cert.
	// It is used to issue client certs used to access the application
	CACert() tls.Certificate

	// ActionIds are the actions that are exposed via the app.
	ActionIds() []ActionId

	// Created is when the application was created
	Created() time.Time
}
