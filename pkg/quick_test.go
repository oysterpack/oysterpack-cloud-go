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

	"golang.org/x/net/context"
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
