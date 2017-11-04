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
	"strings"

	"errors"
	"fmt"

	"github.com/rs/zerolog"
)

func NewSystem(name string, logger zerolog.Logger) (*System, error) {
	if name != strings.TrimSpace(name) {
		return nil, fmt.Errorf("name cannot have whitespace passing %q", name)
	}

	if len(name) == 0 {
		return nil, errors.New("name cannot be blank")
	}

	system := &System{
		Actor: &Actor{
			path: []string{name},

			messageProcessorFactory: func() MessageProcessor { return sysMsgProcessor },
			channelSettings:         systemChannelSettings,
			logger:                  logger,
			supervisor:              RESTART_ACTOR_STRATEGY,
		},
	}
	system.system = system

	if err := system.start(); err != nil {
		return nil, err
	}

	return system, nil
}

// System is an actor hierarchy.
type System struct {
	*Actor
}

func (a *System) LookupActor(address *Address) *Actor {
	if address.Path[0] != a.Name() {
		return nil
	}
	return a.lookupActor(address.Path[1:], address.Id)
}

var (
	systemChannelSettings = map[Channel]*ChannelSettings{
		CHANNEL_SYSTEM: &ChannelSettings{
			Channel: CHANNEL_SYSTEM,
			BufSize: 0,
		},
	}
)
