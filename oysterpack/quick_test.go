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

package oysterpack_test

import (
	"testing"
	"github.com/rs/zerolog/log"
	"sync"
)

func TestEchoService(t *testing.T) {
	StartService(EchoService)
	req := NewEchoRequest("CIAO MUNDO !!!")
	Echo <- req
	log.Info().Msgf("response : %v\n", <-req.ReplyTo)
	StopServices()
	ServicesWaitGroup.Wait()
}

var Stop = make(chan struct{})

var Echo = make(chan *EchoRequest)

var ServicesWaitGroup sync.WaitGroup

func StartService(service func(stop <-chan struct{})) {
	ServicesWaitGroup.Add(1)
	go service(Stop)
}

func StopServices() {
	close(Stop)
}

func EchoService(stop <-chan struct{}) {
	log.Info().Msg("EchoService : Running")
	for {
		select {
		case <-stop:
			log.Info().Msg("EchoService : Stop")
			ServicesWaitGroup.Done()
			return
		case req := <-Echo :
			log.Info().Msgf("EchoService: echo(%v)\n", req.Msg)
			go func() { req.ReplyTo <- req.Msg }()
		}
	}
}

type EchoRequest struct {
	Msg     interface{}
	ReplyTo chan interface{}
}

func NewEchoRequest(msg interface{}) *EchoRequest {
	return &EchoRequest{
		msg,
		make(chan interface{}),
	}
}
