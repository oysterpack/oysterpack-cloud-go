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
	"fmt"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"os"
	"reflect"
	"sync"
	"testing"
	"time"
)

var logger = log.Output(zerolog.ConsoleWriter{Out: os.Stdout}).Level(zerolog.WarnLevel)

type Foo struct{}

func (f Foo) String() string {
	return "Foo"
}

func TestClosingBufferedChannel(t *testing.T) {
	c := make(chan int, 10)
	for i := 0; i < 10; i++ {
		c <- i
	}
	close(c)
	t.Log("closed channel")
	for i := 0; i <= 11; i++ {
		t.Logf("receiving on closed channel [%d]: %v", i, <-c)
	}
}

func TestClosingClosedChannel(t *testing.T) {
	defer func() {
		if p := recover(); p == nil {
			t.Error("closing a closed channel should have triggered a panic")
		}
	}()
	c := make(chan struct{})
	close(c)
	close(c)
}

func TestClosingBlockedChannelOnSend(t *testing.T) {
	msgs := make(chan string)
	waitGroup := sync.WaitGroup{}
	waitGroup.Add(1)
	go func() {
		defer func() {
			switch p := recover().(type) {
			case nil:
				log.Info().Msg("no error")
			case error:
				logger.Warn().Msgf("error(%[1]T) : %[1]s", p)
			}
		}()
		waitGroup.Done()
		log.Info().Msg("msgs <- MSG")
		msgs <- "MSG"
	}()
	waitGroup.Wait()
	time.Sleep(100 * time.Millisecond)
	log.Info().Msg("close(msgs) ...")
	close(msgs)
	log.Info().Msg("close(msgs) !!!")
}

func TestReflect(t *testing.T) {
	var foo fmt.Stringer = fmt.Stringer(Foo{})
	log.Info().Msgf("TypeOf(foo) = %s", reflect.TypeOf(&foo).Elem())
	log.Info().Msgf("TypeOf(foo) == TypeOf(foo) is %v", reflect.TypeOf(&foo).Elem() == reflect.TypeOf(&foo).Elem())
	log.Info().Msgf("ValueOf(foo) = %s", reflect.ValueOf(&foo))
}

func TestEchoService(t *testing.T) {
	StartService(EchoService)
	reqWaitGroup := sync.WaitGroup{}
	for i := 0; i < 100; i++ {
		req := NewEchoRequest(fmt.Sprintf("REQ #%d", i))
		logger.Info().Msgf("TestEchoService : REQ #%d", i)
		reqWaitGroup.Add(1)
		go func() {
			Echo <- req
			response := <-req.ReplyTo
			logger.Info().Msgf("response : %v", response)
			reqWaitGroup.Done()
		}()
	}

	reqWaitGroup.Wait()
	StopServices()
	ServicesWaitGroup.Wait()
}

func BenchmarkEchoService(t *testing.B) {
	Stop = make(chan struct{})
	StartService(EchoService)
	reqWaitGroup := sync.WaitGroup{}

	t.ResetTimer()
	for i := 0; i < t.N; i++ {
		req := NewEchoRequest(fmt.Sprintf("REQ #%d", i))
		reqWaitGroup.Add(1)
		go func() {
			Echo <- req
			<-req.ReplyTo
			reqWaitGroup.Done()
		}()
	}
	reqWaitGroup.Wait()
	t.StopTimer()

	StopServices()
	ServicesWaitGroup.Wait()
}

var Stop = make(chan struct{})

var Echo = make(chan *EchoRequest, 1000)

var ServicesWaitGroup sync.WaitGroup

func StartService(service func(stop <-chan struct{})) {
	ServicesWaitGroup.Add(1)
	log.Info().Msgf("StartService : %T", service)
	go service(Stop)
}

func StopServices() {
	close(Stop)
}

func EchoService(stop <-chan struct{}) {
	logger.Info().Msg("EchoService : Running")
	for {
		select {
		case <-stop:
			logger.Info().Msg("EchoService : Stop")
			ServicesWaitGroup.Done()
			return
		case req := <-Echo:
			logger.Info().Msgf("EchoService: echo(%v)", req.Msg)
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

func TestNegativeDuration(t *testing.T) {
	var d time.Duration = -1
	if d < 0 {
		t.Log("negative durations are allowed")
	}
}

func TestRangeOnNilSlice(t *testing.T) {
	slice := []int{1, 2, 3}
	for i, v := range slice {
		t.Log(i, v)
		slice = nil
	}
}
