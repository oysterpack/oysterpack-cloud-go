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

package app

import (
	"os"
	"sync"
	"testing"
	"time"
)

func TestHealthcheckService(t *testing.T) {
	previousConfigDir := Configs.ConfigDir()
	Configs.SetConfigDir("testdata/healthchecks_test/TestHealthcheckService")

	// clean the config dir
	os.RemoveAll(Configs.ConfigDir())
	os.MkdirAll(Configs.ConfigDir(), 0755)

	// Reset the app
	Reset()
	defer func() {
		Configs.SetConfigDir(previousConfigDir)
		Reset()
	}()

	const (
		HEALTHCHECK_1 = HealthCheckID(1)
		HEALTHCHECK_2 = HealthCheckID(2)
		HEALTHCHECK_3 = HealthCheckID(3)
	)

	t.Run("No HealthCheckServiceSpec - no healthchecks", func(t *testing.T) {
		// clean the config dir
		os.RemoveAll(Configs.ConfigDir())
		os.MkdirAll(Configs.ConfigDir(), 0755)
		Reset()

		ids := HealthChecks.HealthCheckIDs()
		t.Logf("HealthCheckIDs : %v", ids)
		if len(ids) != 0 {
			t.Errorf("there should be no healthchecks registered : %v", ids)
		}

		results := HealthChecks.HealthCheckResults()

		t.Logf("HealthCheckResults : %v", results)
		if len(results) != 0 {
			t.Errorf("there should be no healthchecks registered : %v", results)
		}

		if HealthChecks.Registered(HEALTHCHECK_1) {
			t.Error("there should be no healthchecks registered")
		}

		if _, err := HealthChecks.HealthCheckResult(HEALTHCHECK_1); !IsError(err, ErrSpec_HealthCheckNotRegistered.ErrorID) {
			t.Errorf("Expected ErrHealthCheckNotRegistered : %[1]T : %[1]v", err)
		}

		if HealthChecks.HealthCheckSpec(HEALTHCHECK_1) != nil {
			t.Error("there should be no healthchecks registered")
		}

		if _, err := HealthChecks.Run(HEALTHCHECK_1); !IsError(err, ErrSpec_HealthCheckNotRegistered.ErrorID) {
			t.Errorf("Expected ErrHealthCheckNotRegistered : %[1]T : %[1]v", err)
		}
	})

	t.Run("No HealthCheckServiceSpec - register healthchecks", func(t *testing.T) {
		// clean the config dir
		os.RemoveAll(Configs.ConfigDir())
		os.MkdirAll(Configs.ConfigDir(), 0755)
		Reset()

		HealthChecks.Register(HEALTHCHECK_1, func(result chan<- error, cancel <-chan struct{}) {
			close(result)
		})

		if !HealthChecks.Registered(HEALTHCHECK_1) {
			t.Error("Should be registered")
		}

		beforeRunningHealthCheck := time.Now()
		if result, err := HealthChecks.Run(HEALTHCHECK_1); err != nil {
			t.Errorf("HealthCheck should have passed : %[1]T : %[1]v", err)
		} else {
			t.Logf("HEALTHCHECK_1 result : %v", result)
			if result.Err != nil {
				t.Errorf("HealthCheck should have passed : %[1]T : %[1]v", err)
			}
			if !result.Time.After(beforeRunningHealthCheck) {
				t.Errorf("The healthcheck run timestamp should have been later : %v : %v", beforeRunningHealthCheck, result.Time)
			}
		}

		result1, err := HealthChecks.Run(HEALTHCHECK_1)
		if err != nil {
			t.Fatal(err)
		}
		for i := 0; i < 10; i++ {
			result2, err := HealthChecks.Run(HEALTHCHECK_1)
			if err != nil {
				t.Fatal(err)
			}
			if !result2.Time.After(result1.Time) {
				t.Fatalf("result2 time was not afte result1 time : %v : %v", result1.Time, result2.Time)
			}
			t.Log(result2)
			result1 = result2
		}

		results := HealthChecks.HealthCheckResults()
		for k, v := range results {
			t.Log(k, v)
		}

		wg := sync.WaitGroup{}
		wg.Add(1)
		go func() {
			defer wg.Done()
			results := HealthChecks.HealthCheckResults()
			for k, v := range results {
				t.Log(k, v)
			}
		}()
		wg.Wait()

		paused := HealthChecks.PausedHealthChecks()
		if len(paused) > 0 {
			t.Errorf("No healthchecks should be paused : %v", paused)
		}
		if !HealthChecks.PauseHealthCheck(HEALTHCHECK_1) {
			t.Error("Should have paused")
		}
		paused = HealthChecks.PausedHealthChecks()
		t.Logf("paused : %v", paused)
		if len(paused) != 1 {
			t.Errorf("HEALTHCHECK_1 should be paused : %v", paused)
		}
		if err := HealthChecks.ResumeHealthCheck(HEALTHCHECK_1); err != nil {
			t.Error(err)
		}
		paused = HealthChecks.PausedHealthChecks()
		t.Logf("paused : %v", paused)
		if len(paused) != 0 {
			t.Errorf("No healthchecks should be paused : %v", paused)
		}

	})
}
