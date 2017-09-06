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

import (
	"reflect"
)

type Service interface {
	Context() Context
}

// TODO: lifecycle
// TODO: dependencies
// TODO: events
// TODO: logging
// TODO: metrics
// TODO: healthchecks
// TODO: alarms
// TODO: config (JSON)
// TODO: readiness probe
// TODO: liveliness probe
// TODO: devops
// TODO: security
// TODO: gRPC - frontend
type Context struct {
	reflect.Type

	Server
}
