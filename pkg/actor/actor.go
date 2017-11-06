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

package actor

import (
	"time"

	"github.com/nats-io/nuid"
	"github.com/rs/zerolog"
	"gopkg.in/tomb.v2"
)

type Actor struct {
	created time.Time

	name string
	path string
	id   string

	parent   *Actor
	children map[string]*Actor

	system *System

	messageProcessorFactory MessageProcessorFactory
	messageChannel          chan *Envelope

	tomb.Tomb

	*nuid.NUID
	logger zerolog.Logger
}

// Name returns the actor name. The name must be unique with all other siblings.
func (a *Actor) Name() string {
	return a.name
}
