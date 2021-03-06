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

package pkg_test

import (
	"errors"
	"testing"

	"time"

	"fmt"

	"golang.org/x/net/context"
	"gopkg.in/tomb.v2"
)

type Foo interface {
	Foo()
}

type Bar interface {
	Bar()
}

type FooBar struct {
}

func (a FooBar) Foo() {}
func (a FooBar) Bar() {}

func TestFooBar(t *testing.T) {
	var foo Foo = &FooBar{}
	var bar Bar = foo.(Bar)
	bar.Bar()
}

func TestContext(t *testing.T) {

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	result := make(chan error)
	go func(ctx context.Context, result chan<- error) {
		select {
		case <-ctx.Done():
			result <- errors.New("Cancelled")
			close(result)
		default:
			close(result)
		}
	}(ctx, result)

	t.Logf("result : %v", <-result)

}

func TestContext_Cancelled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	result := make(chan error)
	go func(ctx context.Context, result chan<- error) {
		select {
		case <-ctx.Done():
			result <- errors.New("Cancelled")
			close(result)
		default:
			close(result)
		}
	}(ctx, result)

	t.Logf("result : %v", <-result)

}

type Error1 struct {
}

func (a *Error1) Error() string {
	return "ERR1"
}

func TestErrorTypeAssertion(t *testing.T) {
	var err error = &Error1{}
	switch err := err.(type) {
	case *Error1:
		t.Log(err)
	default:
		t.Error("Unable to match type to *Error1")
	}

}

func TestTimeUnix(t *testing.T) {
	now := time.Now()

	t.Logf("now = %v", now)
	t.Logf("time.Unix(now.Unix(),0) -> %v", time.Unix(now.Unix(), 0))
	t.Logf("time.Unix(now.Unix(),now.UnixNano()) -> %v", time.Unix(now.Unix(), now.UnixNano()))
	t.Logf("time.Unix(0,now.UnixNano()) -> %v", time.Unix(0, now.UnixNano()))
	t.Logf("time.Unix(now.Unix(),now.UnixNano()).Equal(now) -> %v", time.Unix(now.Unix(), now.UnixNano()).Equal(now))
	t.Logf("time.Unix(now.Unix(), 0).Equal(now) -> %v", time.Unix(now.Unix(), 0).Equal(now))
	t.Logf("time.Unix(0,now.UnixNano()).Equal(now) -> %v", time.Unix(0, now.UnixNano()).Equal(now))

}

func TestTomb(t *testing.T) {
	a := tomb.Tomb{}

	for i := 0; i < 100; i++ {
		ii := i
		a.Go(func() error {
			t.Log(ii)
			<-a.Dying()
			t.Log(ii, "DONE")
			return nil
		})
	}

	a.Kill(nil)
	a.Wait()

	a.Kill(nil)
	a.Wait()

	c := make(chan interface{})
	close(c)

	defer func() {
		if p := recover(); p != nil {
			t.Logf("%[1]T panic : %[1]v", p)
		}
	}()
	select {
	case <-a.Dying():
	case c <- 1:
	}

}

func TestReturnValue(t *testing.T) {
	var failure error

	func() (err error) {
		defer func() {
			if err != nil {
				failure = err
			}
		}()

		return errors.New("BOOM!!!")
	}()

	t.Logf("failure : %v", failure)

	if failure == nil {
		t.Error("Should have failed")
	}
}

func TestNilMap(t *testing.T) {
	var m map[string]int

	t.Log(len(m))
	t.Log(m["a"])
}

func TestDefer(t *testing.T) {
	defer t.Log("A")
	defer t.Log("B")
	defer t.Log("C")
}

func TestFmtSprint(t *testing.T) {
	t.Log(fmt.Sprint("abc", 1))
}

func TestChanSelect(t *testing.T) {
	c := make(chan interface{})
	go func() {
		defer close(c)
		c <- 1
	}()
LOOP:
	for {
		select {
		case v, ok := <-c:
			t.Logf("v: %v, ok: %v", v, ok)
			if !ok {
				t.Log("chan was closed")
				break LOOP
			}
		}
	}
}

// the goal was to see if there was any value in creating predefined variables for empty structs
//
// BenchmarkEmptyStruct/context.WithValue(a{},a{})-8               20000000                80.7 ns/op            48 B/op          1 allocs/op
// BenchmarkEmptyStruct/context.WithValue(aa,aa)-8                 20000000                82.8 ns/op            48 B/op          1 allocs/op
//
// ANSWER : There was no value add, i.e., empty structs do not allocate memory - period.
func BenchmarkEmptyStruct(b *testing.B) {
	type Key struct{}
	type a Key

	b.Run("context.WithValue(a{},a{})", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			context.WithValue(context.Background(), a{}, a{})
		}
	})

	aa := a{}

	b.Run("context.WithValue(aa,aa)", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			context.WithValue(context.Background(), aa, aa)
		}
	})
}
