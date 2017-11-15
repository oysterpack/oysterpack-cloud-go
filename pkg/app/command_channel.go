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

package app

func NewServiceCommandChannel(service *Service, chanSize uint16) *ServiceCommandChannel {
	return &ServiceCommandChannel{Service: service, c: make(chan func(), chanSize)}
}

// ServiceCommandChannel pairs a command channel to a service.
//
type ServiceCommandChannel struct {
	*Service

	c chan func()
}

// CommandChan returns the chan to used to receive commands to execute.
func (a *ServiceCommandChannel) CommandChan() <-chan func() {
	return a.c
}

// Submit will put the command function on the service channel.
// If the service is not alive then ErrServiceNotAlive is returned.
func (a *ServiceCommandChannel) Submit(f func()) error {
	select {
	case <-a.Dying():
		return ErrServiceNotAlive
	case a.c <- f:
		return nil
	}
}
