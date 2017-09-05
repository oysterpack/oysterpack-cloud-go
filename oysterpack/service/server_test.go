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

package service_test

import (
	"github.com/oysterpack/oysterpack.go/oysterpack/service"
	"testing"
	"time"
)

func TestNewServer_WithNilLifeCycleFunctions(t *testing.T) {
	var init service.Init = nil
	var run service.Run = nil
	var destroy service.Destroy = nil

	server := service.NewServer(init, run, destroy)
	if !server.State().New() {
		t.Errorf("Server state should be 'New', but instead was : %q", server.State())
	}

	if err := server.StartAsync();err != nil {
		t.Error(err)
	}
	for i:=0;i<3;i++ {
		if err := server.AwaitRunning(time.Second);err != nil {
			t.Error(err)
			break
		}
		if server.State().Running() {
			break
		}
		t.Log("Waiting for server to run ...")
	}

	if !server.State().Running() {
		t.Errorf("Server state should be 'Running', but instead was : %q", server.State())
	}
	server.StopAsyc()
	for i:=0;i<3;i++ {
		server.AwaitTerminated(time.Second)
		if server.State().Terminated() {
			break
		}
		t.Log("Waiting for server to terminate ...")
	}

	if !server.State().Terminated() {
		t.Errorf("Server state should be 'Terminated', but instead was : %q", server.State())
	}
}
