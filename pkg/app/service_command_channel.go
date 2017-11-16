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

// NewServiceCommandChannel creates a new ServiceCommandChannel with the specified channel buffer size.
func NewServiceCommandChannel(service *Service, chanSize uint16) *ServiceCommandChannel {
	return &ServiceCommandChannel{Service: service, c: make(chan func(), chanSize)}
}

// ServiceCommandChannel pairs a command channel to a service.
//
// Service Channel Server Design Pattern
// - The server's main goroutine will listen on the command channel, and when a command function is received it will
//   execute the function.
// - The service state is accessed within the service's main goroutine, i.e., the code executed by the command can assume
//   that it will be run in a threadsafe manner.
// - Data is passed into the service via the function closure.
// - Data is passed out over channels that are enclosed by the closure function.
// - The service's main goroutine acts like an event loop which executes code that is received on the service commmand channel
// - NOTE: the service may embed more than 1 service command channel.
// - example service main goroutine :
//
//		a.Go(func() error {
//			for {								// command event loop
//				select {
//				case <- a.Dying(): 				// listen for kill signal
//					a.stop()					// stop the service gracefully
//					return nil					// server exit
//			    case f := <- a.CommandChan(): 	// listen for commands
//					f()							// execute the command
//			    }
//		     }
//		}
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
// Submit will block until the command function can be put on the channel.
//
// errors:
//  - ErrServiceNotAlive
func (a *ServiceCommandChannel) Submit(f func()) error {
	select {
	case <-a.Dying():
		return ErrServiceNotAlive
	case a.c <- f:
		return nil
	}
}
