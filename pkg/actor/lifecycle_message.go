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

var (
	STARTED    = &Started{EMPTY}
	STOPPING   = &Stopping{EMPTY}
	STOPPED    = &Stopped{EMPTY}
	RESTARTING = &Restarting{EMPTY}
)

const (
	MESSAGE_TYPE_STARTED MessageType = iota
	MESSAGE_TYPE_STOPPING
	MESSAGE_TYPE_STOPPED
	MESSAGE_TYPE_RESTARTED
)

type LifeCycleMessage interface {
	SystemMessage
	LifeCycleMessage()
}

type Started struct {
	*Empty
}

func (a *Started) SystemMessage()    {}
func (a *Started) LifeCycleMessage() {}

type Stopping struct {
	*Empty
}

func (a *Stopping) SystemMessage()    {}
func (a *Stopping) LifeCycleMessage() {}

type Stopped struct {
	*Empty
}

func (a *Stopped) SystemMessage()    {}
func (a *Stopped) LifeCycleMessage() {}

type Restarting struct {
	*Empty
}

func (a *Restarting) SystemMessage()    {}
func (a *Restarting) LifeCycleMessage() {}
