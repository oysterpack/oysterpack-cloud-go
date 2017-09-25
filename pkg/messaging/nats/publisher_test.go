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

package nats_test

import (
	"testing"

	"github.com/nats-io/go-nats"

	natsop "github.com/oysterpack/oysterpack.go/pkg/messaging/nats"
)

func TestNewGobPublisher(t *testing.T) {
	nc, _ := nats.Connect(nats.DefaultURL)
	defer nc.Close()

	subject := "ping"

	publisher := natsop.NewPublisher(nc, subject, 10)
	defer publisher.Close()

	for i := 1; i <= 100; i++ {
		if err := publisher.Publish(&Person{"alfio", "zappala", i}); err != nil {
			t.Errorf("Publish failed : %v", err)
		}
		t.Logf("published message #%d", i)
	}

}

type Person struct {
	Fname string
	Lname string
	Id    int
}
