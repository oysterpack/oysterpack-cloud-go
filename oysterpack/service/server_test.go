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

	if _, err := startServer(server, t); err != nil {
		t.Error(err)
	}

	if !server.State().Running() {
		t.Errorf("Server state should be 'Running', but instead was : %q", server.State())
	}

	stopServer(server, t)
	if !server.State().Terminated() {
		t.Errorf("Server state should be 'Terminated', but instead was : %q", server.State())
	}
}

func TestNewServer_WithNonNilLifeCycleFunctions(t *testing.T) {
	var init service.Init = func() error {
		t.Log("init")
		return nil
	}
	var run service.Run = func(ctx *service.RunContext) error {
		t.Log("running")
		for {
			select {
			case <-ctx.StopTrigger():
				t.Log("stop triggered")
				return nil
			}
		}
	}
	var destroy service.Destroy = func() error {
		t.Log("destroy")
		return nil
	}

	server := service.NewServer(init, run, destroy)
	if !server.State().New() {
		t.Errorf("Server state should be 'New', but instead was : %q", server.State())
	}

	if _, err := startServer(server, t); err != nil {
		t.Error(err)
	}

	if !server.State().Running() {
		t.Errorf("Server state should be 'Running', but instead was : %q", server.State())
	}

	stopServer(server, t)
	if !server.State().Terminated() {
		t.Errorf("Server state should be 'Terminated', but instead was : %q", server.State())
	}
}

func TestNewServer_InitPanics(t *testing.T) {
	var init service.Init = func() error {
		panic("Init is panicking")
		return nil
	}
	var run service.Run = func(ctx *service.RunContext) error {
		t.Log("running")
		for {
			select {
			case <-ctx.StopTrigger():
				t.Log("stop triggered")
				return nil
			}
		}
	}
	var destroy service.Destroy = func() error {
		t.Log("destroy")
		return nil
	}

	server := service.NewServer(init, run, destroy)
	if !server.State().New() {
		t.Errorf("Server state should be 'New', but instead was : %q", server.State())
	}

	if started, err := startServer(server, t); !started && err != nil {
		switch err.(type) {
		case *service.ServiceError:
		default:
			t.Errorf("Expected a service.ServiceError to be returned, but was %T : %v", err, err)
		}
		expectedError := err.(*service.ServiceError)
		t.Logf("expected error : %v", expectedError)
	} else {
		if started {
			t.Errorf("Expected server to fail to start")
		}

		if err == nil {
			t.Errorf("Expected a service.ServiceError to be returned")
		}
	}

	if !server.State().Failed() {
		t.Errorf("Server state should be 'Failed', but instead was : %q", server.State())
	}

	// stopping a server that is already stopped should be ok
	if !stopServer(server, t) {
		t.Errorf("Server should already be in a stopped state, but we timed out waiting for the server to terminate")
	}
	if !server.State().Failed() {
		t.Errorf("Server state should be 'Failed', but instead was : %q", server.State())
	}
}

func TestNewServer_RunPanics(t *testing.T) {
	var init service.Init = func() error {
		t.Log("init")
		return nil
	}
	var run service.Run = func(ctx *service.RunContext) error {
		panic("Run is panicking")
	}
	var destroy service.Destroy = func() error {
		t.Log("destroy")
		return nil
	}

	server := service.NewServer(init, run, destroy)
	if !server.State().New() {
		t.Errorf("Server state should be 'New', but instead was : %q", server.State())
	}

	if started, err := startServer(server, t); !started && err != nil {
		switch err.(type) {
		case *service.ServiceError:
		default:
			t.Errorf("Expected a service.ServiceError to be returned, but was %T : %v", err, err)
		}
		expectedError := err.(*service.ServiceError)
		t.Logf("expected error : %v", expectedError)
	} else {
		if started {
			t.Errorf("Expected server to fail to start")
		}

		if err == nil {
			t.Errorf("Expected a service.ServiceError to be returned")
		}
	}

	if !server.State().Failed() {
		t.Errorf("Server state should be 'Failed', but instead was : %q", server.State())
	}

	// stopping a server that is already stopped should be ok
	if !stopServer(server, t) {
		t.Errorf("Server should already be in a stopped state, but we timed out waiting for the server to terminate")
	}
	if !server.State().Failed() {
		t.Errorf("Server state should be 'Failed', but instead was : %q", server.State())
	}
}

func TestNewServer_DestroyPanics(t *testing.T) {
	var init service.Init = func() error {
		t.Log("init")
		return nil
	}
	var run service.Run = func(ctx *service.RunContext) error {
		t.Log("running")
		for {
			select {
			case <-ctx.StopTrigger():
				t.Log("stop triggered")
				return nil
			}
		}
	}
	var destroy service.Destroy = func() error {
		panic("Destroy is panicking")
	}

	server := service.NewServer(init, run, destroy)
	if !server.State().New() {
		t.Errorf("Server state should be 'New', but instead was : %q", server.State())
	}

	if _, err := startServer(server, t); err != nil {
		t.Error(err)
	}

	if !server.State().Running() {
		t.Errorf("Server state should be 'Running', but instead was : %q", server.State())
	}

	stopServer(server, t)
	if !server.State().Failed() {
		t.Errorf("Server state should be 'Terminated', but instead was : %q", server.State())
	}
	t.Log(server.FailureCause())
}

// startServer waits up to 3 seconds for the server to start - checking every second
// returns true is the server started
// returns false if we timed out waiting for the server to start
func startServer(server *service.Server, t *testing.T) (bool, error) {
	if err := server.StartAsync(); err != nil {
		return false, err
	}
	for i := 1; i <= 3; i++ {
		if err := server.AwaitRunning(time.Second); err != nil {
			return false, err
		}
		if server.State().Running() {
			return true, nil
		}
		t.Logf("Waiting for server to run for %d sec ...", i)
	}
	return false, nil
}

// stopServer waits up to 3 seconds for the server to stop
// returns true is the server stopped
// returns false if we timed out waiting for the server to stop
func stopServer(server *service.Server, t *testing.T) bool {
	server.StopAsyc()
	for i := 1; i <= 3; i++ {
		server.AwaitTerminated(time.Second)
		if server.State().Stopped() {
			return true
		}
		t.Logf("Waiting for server to run for %d sec ...", i)
	}
	return false
}
