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
func NewServiceCommandChannel(service *Service, chanSize uint16) (*ServiceCommandChannel, error) {
	if service == nil {
		return nil, ErrServiceNil
	}
	if !service.Alive() {
		return nil, ErrServiceNotAlive
	}
	return &ServiceCommandChannel{Service: service, c: make(chan func(), chanSize)}, nil
}

// ServiceCommandChannel pairs a command channel to a service. It is the embedding service's responsibility to subscribe to
// service command channel and execute the command functions.
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
// CommandServer is reference implementation and is designed to be embedded into services.
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

// NewCommandServer creates a new CommandServer and returns it
func NewCommandServer(service *Service, chanSize uint16, name string, init func(), destroy func()) (*CommandServer, error) {
	serviceCommandChannel, err := NewServiceCommandChannel(service, chanSize)
	if err != nil {
		return nil, err
	}
	cmdServer := &CommandServer{
		ServiceCommandChannel: serviceCommandChannel,
		Name:    name,
		Init:    init,
		Destroy: destroy,
	}
	cmdServer.start()
	return cmdServer, nil
}

// CommandServer will execute commands in a background goroutine serially.
// Think of the command server as a command event loop.
// The CommandServer can be supplied with optional Init and Destroy funcs.
type CommandServer struct {
	*ServiceCommandChannel

	// OPTIONAL - used for logging purposes
	Name string

	// OPTIONAL - invoked before the any commands are run
	Init func()
	// OPTIONAL - invoked when the service has been killed
	Destroy func()
}

// Start starts the server's main goroutine, i.e., the command event loop.
//
// 1. Log SERVICE_STARTING event
// 2. Init()
// 3. Log the SERVICE_STARTED event
// 4. Run the command event loop until the service kill signal is received
// 5. Destroy()
//
// NOTE: the SERVICE_STOPPING and SERVICE_STOPPED log events are logged by the app
//
// errors:
// - ErrServiceNotAlive - if trying to start after the service has been killed
func (a *CommandServer) start() error {
	if !a.Alive() {
		return ErrServiceNotAlive
	}

	a.Go(func() error {
		a.starting()
		if a.Init != nil {
			a.Init()
		}
		a.started()

		for {
			select {
			case <-a.Dying():
				if a.Destroy != nil {
					a.Destroy()
				}
				return nil
			case f := <-a.CommandChan():
				f()
			}
		}
	})
	return nil
}

func (a *CommandServer) starting() {
	evt := SERVICE_STARTING.Log(a.Logger().Info())
	if a.Name != "" {
		evt.Str("cmdsvr", a.Name)
	}
	evt.Msg("starting")
}

func (a *CommandServer) started() {
	evt := SERVICE_STARTED.Log(a.Logger().Info())
	if a.Name != "" {
		evt.Str("cmdsvr", a.Name)
	}
	evt.Msg("started")
}
