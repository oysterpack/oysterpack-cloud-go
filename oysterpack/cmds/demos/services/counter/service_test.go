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

package counter_test

import (
	"testing"

	"github.com/oysterpack/oysterpack.go/oysterpack/cmds/demos/services/counter"
	"github.com/oysterpack/oysterpack.go/oysterpack/service"
)

func TestClientConstructor(t *testing.T) {
	app := service.NewApplication(service.ApplicationSettings{})
	app.Service().StartAsync()
	app.Service().AwaitUntilRunning()
	defer app.Service().Stop()

	client := app.MustRegisterService(counter.ClientConstructor)
	client.Service().AwaitUntilRunning()

	counter := client.(counter.Interface)

	var count uint64 = 0
	for i := 0; i < 100; i++ {
		count = counter.NextInt()
	}

	if count != 100 {
		t.Errorf("Expected count=100, but count=%d", count)
	}
}
