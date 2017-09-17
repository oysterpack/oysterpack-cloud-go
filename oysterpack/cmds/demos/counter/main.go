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

package main

import (
	"bufio"
	"fmt"
	"os"

	"github.com/oysterpack/oysterpack.go/oysterpack/cmds/demos/services/counter"
	"github.com/oysterpack/oysterpack.go/oysterpack/service"
)

// this app will increment and print the counter value each time the user hits the return key
func main() {
	defer service.App().Stop()
	service.App().Start()
	client := service.App().RegisterService(counter.ClientConstructor)
	client.Service().AwaitUntilRunning()
	counter := client.(counter.Interface)

	go func() {
		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			fmt.Println(counter.NextInt())
		}
	}()

	service.App().Service().AwaitUntilStopped()
}