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

package app_test

import (
	"testing"

	"github.com/oysterpack/oysterpack.go/pkg/app"
)

type FooCommandService struct {
	*app.CommandServer

	N uint64
}

func (a *FooCommandService) Foo() error {
	return a.Submit(func() {})
}

func (a *FooCommandService) GetN1() (uint64, error) {
	// Using a buffered chan enables the command func to be decoupled from the chan
	// The command function can put the result on the channel in a non-blocking fashion
	// The tradeoff is that the memory allocation overhead
	//
	// In benchmarks, this appraoch appears to provide ~30% performance boost as compared to the approach used in the GetN0
	// function.
	c := make(chan uint64, 1)
	a.Submit(func() {
		c <- a.N
	})

	select {
	case n := <-c:
		return n, nil
	case <-a.Dying():
		return 0, app.ErrServiceNotAlive
	}
}

func (a *FooCommandService) GetN0() (uint64, error) {
	c := make(chan uint64)
	a.Submit(func() {
		select {
		case c <- a.N:
		case <-a.Dying():
		}
	})

	select {
	case n := <-c:
		return n, nil
	case <-a.Dying():
		return 0, app.ErrServiceNotAlive
	}
}

func (a *FooCommandService) IncN() error {
	return a.Submit(func() {
		a.N++
	})
}

func BenchmarkCommandServer(b *testing.B) {
	const SERVICE_ID = app.ServiceID(1)

	commandServer, err := app.NewCommandServer(app.NewService(SERVICE_ID), 0, "", nil, nil)
	if err != nil {
		b.Fatal(err)
	}

	foo := &FooCommandService{CommandServer: commandServer}
	foo.Foo()

	//BenchmarkCommandServer/Foo_-_chan_buf_size_=_0-8                 3000000               772 ns/op               0 B/op          0 allocs/op
	b.Run("Foo - chan buf size = 0", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			foo.Foo()
		}
	})

	//BenchmarkCommandServer/GetN1_-_chan_buf_size_=_0-8               1000000              1754 ns/op             144 B/op          2 allocs/op
	b.Run("GetN1 - chan buf size = 0", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			foo.GetN1()
		}
	})

	//BenchmarkCommandServer/GetN0_-_chan_buf_size_=_0-8                500000              2287 ns/op             128 B/op          2 allocs/op
	b.Run("GetN0 - chan buf size = 0", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			foo.GetN0()
		}
	})

	//BenchmarkCommandServer/IncN_-_chan_buf_size_=_0-8                2000000               960 ns/op              16 B/op          1 allocs/op
	N := 0
	b.Run("IncN - chan buf size = 0", func(b *testing.B) {
		N += b.N
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			foo.IncN()
		}
	})

	n, err := foo.GetN1()
	if err != nil {
		b.Error(err)
	}
	if int(n) != N {
		b.Errorf("%d != %d", n, N)
	}

}
