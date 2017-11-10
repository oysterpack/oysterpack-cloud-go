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
	"sync"

	"gopkg.in/tomb.v2"
)

var gofuncs chan func()

func init() {
	gofuncs = make(chan func(), 1024)
	go func() {
		for {
			select {
			case f := <-gofuncs:
				go f()
			}
		}
	}()
}

func NewSystem() *System {
	return &System{registry: make(map[string]*Actor)}
}

type System struct {
	lock     sync.RWMutex
	registry map[string]*Actor
}

func (a *System) KillRootActors(t tomb.Tomb) {
	for actor := range a.RootActors(t) {
		actor.Kill()

		select {
		case <-actor.Dead():
		case <-t.Dying():
			return
		}
	}
}

func (a *System) registerActor(actor *Actor) bool {
	if actor == nil {
		return false
	}
	a.lock.Lock()
	key := actor.path
	if _, ok := a.registry[key]; ok {
		a.lock.Unlock()
		return false
	}
	a.registry[key] = actor
	a.lock.Unlock()
	LOG_EVENT_REGISTERED.Log(actor.logger.Debug()).Msg("")
	return true
}

func (a *System) unregisterActor(actor *Actor) {
	if actor == nil {
		return
	}
	a.lock.Lock()
	key := actor.path
	delete(a.registry, key)
	a.lock.Unlock()
}

func (a *System) Actor(address *Address) (*Actor, bool) {
	a.lock.RLock()
	actor, exists := a.registry[address.Path]
	a.lock.RUnlock()
	if exists && address.Id != nil && actor.id != *address.Id {
		return nil, false
	}
	return actor, exists
}

func (a *System) ActorCount() (count int) {
	a.lock.RLock()
	count = len(a.registry)
	a.lock.RUnlock()
	return
}

func (a *System) RootActors(t tomb.Tomb) <-chan *Actor {
	c := make(chan *Actor)

	t.Go(func() error {
		defer close(c)
		a.lock.RLock()
		rootActors := []*Actor{}
		for _, actor := range a.registry {
			if actor.parent == nil {
				rootActors = append(rootActors, actor)
			}
		}
		a.lock.RUnlock()

		for _, actor := range rootActors {
			select {
			case <-t.Dying():
				return nil
			case c <- actor:
			}
		}
		return nil
	})

	return c
}
