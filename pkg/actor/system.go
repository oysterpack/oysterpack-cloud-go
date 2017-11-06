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

import "sync"

type System struct {
	*Actor

	ActorRegistry
}

type ActorRegistry struct {
	sync.RWMutex
	registry map[string]*Actor
}

func (a ActorRegistry) registerActor(actor *Actor) bool {
	if actor == nil {
		return false
	}
	a.Lock()
	key := actor.address.path.PathKey()
	if _, ok := a.registry[key]; ok {
		return false
	}
	a.registry[key] = actor
	a.Unlock()
	return true
}

func (a ActorRegistry) unregisterActor(actor *Actor) {
	if actor == nil {
		return
	}
	a.Lock()
	key := actor.address.path.PathKey()
	delete(a.registry, key)
	a.Unlock()
}

func (a ActorRegistry) Actor(address *Address) (actor *Actor, exists bool) {
	if address == nil {
		return nil, false
	}
	a.RLock()
	actor, exists = a.registry[address.Path().PathKey()]
	a.RUnlock()
	return
}
