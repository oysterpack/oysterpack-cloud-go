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

package counter

import (
	"github.com/oysterpack/oysterpack.go/oysterpack/commons"
	"github.com/oysterpack/oysterpack.go/oysterpack/service"
)

type Interface interface {
	NextInt() uint64
}

var CounterServiceInterface service.ServiceInterface = func() service.ServiceInterface {
	var c Interface = &client{}
	serviceInterface, err := commons.ObjectInterface(&c)
	if err != nil {
		panic(err)
	}
	return serviceInterface
}()

type client struct {
	*service.RestartableService

	counter uint64
	nextInt chan chan<- uint64
}

func (a *client) run(ctx *service.Context) error {
	for {
		select {
		case <-ctx.StopTrigger():
			return nil
		case replyTo := <-a.nextInt:
			a.counter++
			replyTo <- a.counter
		}
	}
}

func (a *client) NextInt() (i uint64) {
	defer func() { a.Service().Logger().Info().Msgf("NextInt() : %d", i) }()
	a.Service().Logger().Info().Msg("NextInt() ...")
	ch := make(chan uint64, 1)
	a.nextInt <- ch
	i = <-ch
	return
}

func (a *client) newService() service.Service {
	return service.NewService(service.ServiceSettings{ServiceInterface: CounterServiceInterface, Run: a.run})
}

func ClientConstructor(app service.Application) service.Client {
	c := &client{
		nextInt: make(chan chan<- uint64),
	}
	c.RestartableService = service.NewRestartableService(c.newService)
	return c
}