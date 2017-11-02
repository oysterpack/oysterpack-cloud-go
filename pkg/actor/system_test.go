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

package actor_test

import (
	"testing"

	"github.com/oysterpack/oysterpack.go/pkg/actor"
	"github.com/rs/zerolog/log"
)

func TestNewSystem(t *testing.T) {

	system, err := actor.NewSystem("oysterpack", log.Logger)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		log.Logger.Info().Msg("Killing system ...")
		system.Kill(nil)
		log.Logger.Info().Msg("Kill signalled. Waiting for system to terminate ...")
		if err := system.Wait(); err != nil {
			t.Error(err)
		}
		log.Logger.Info().Msg("System terminated")
	}()

	log.Logger.Info().Msgf("address = %v, name = %s, path = %v, id = %s", system.Address(), system.Name(), system.Path(), system.Id())

	if !system.Alive() {
		t.Error("system is not alive")
	}

	if system.Name() != "oysterpack" {
		t.Errorf("name did not match : %v", system.Name())
	}

}
