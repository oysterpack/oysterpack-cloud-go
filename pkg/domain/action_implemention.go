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
	"github.com/oysterpack/oysterpack.go/pkg/comp"
	"github.com/oysterpack/oysterpack.go/pkg/metrics"
)

type ActionImplementation interface {
	Mode() ImplementationMode

	MetricOpts() *metrics.MetricOpts

	HealthChecks() comp.HealthChecks
}

// ImplementationMode indicates the implementation mode
type ImplementationMode uint8

const (
	// LOCAL means the action runs locally, i.e., within the same process
	// This means the server is running locally.
	LOCAL ImplementationMode = iota
	// NATSCLIENT means the action is invoked remotely via NATS, i.e., messages are sent to the server via NATS
	NATS_CLIENT
	// NATS_SERVER means the action is processing messages received via NATS
	NATS_SERVER
)
