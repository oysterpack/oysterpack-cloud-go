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
	"github.com/Masterminds/semver"
	"github.com/oysterpack/oysterpack.go/pkg/commons/reflect"
	"github.com/oysterpack/oysterpack.go/pkg/service"
)

// Service is the service server interface
type Service interface {
	NextInt() uint64
}

// CounterServiceInterface is the ServiceInterface
var CounterServiceInterface service.ServiceInterface = func() service.ServiceInterface {
	var c Service = &server{}
	serviceInterface, err := reflect.ObjectInterface(&c)
	if err != nil {
		panic(err)
	}
	return serviceInterface
}()

type server struct {
	*service.RestartableService

	counter uint64
	nextInt chan chan<- uint64
}

func (a *server) run(ctx *service.Context) error {
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

func (a *server) NextInt() (i uint64) {
	defer func() { a.Service().Logger().Info().Msgf("NextInt() : %d", i) }()
	a.Service().Logger().Info().Msg("NextInt() ...")
	ch := make(chan uint64, 1)
	a.nextInt <- ch
	i = <-ch
	return
}

func (a *server) newService() service.Service {
	version, err := semver.NewVersion("1.0.0")
	if err != nil {
		panic(err)
	}
	return service.NewService(service.ServiceSettings{ServiceInterface: CounterServiceInterface, Version: version, Run: a.run})
}

// ClientConstructor is the service ClientConstructor
func ClientConstructor(app service.Application) service.Client {
	c := &server{
		nextInt: make(chan chan<- uint64),
	}
	c.RestartableService = service.NewRestartableService(c.newService)
	return c
}
