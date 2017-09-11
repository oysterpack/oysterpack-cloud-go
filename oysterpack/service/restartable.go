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

package service

import "sync"

// RestartableService manages provides support to "restart" the service.
// The service is not really restarted because it is illegal to start a service that has been stopped.
// Instead a new instance of the service is created (via the provided ServiceConstructor) and then started
// New instances should be created using NewRestartableService()
type RestartableService struct {
	serviceMutex sync.RWMutex
	service      *Service
	restartCount int

	NewService ServiceConstructor
}

// Service returns the managed service instance
func (a *RestartableService) Service() *Service {
	a.serviceMutex.RLock()
	defer a.serviceMutex.RUnlock()
	return a.service
}

// RestartService restarts the service. It performs the following steps:
// 1. stops the service
// 2. waits for the service to stop
// 3. starts the service async
// 4. increments the restart counter
func (a *RestartableService) RestartService() {
	a.serviceMutex.Lock()
	defer a.serviceMutex.Unlock()
	if a.service != nil {
		a.service.StopAsyc()
		a.service.AwaitUntilStopped()
	}
	a.service = a.NewService()
	a.service.StartAsync()
	a.restartCount++
}

// RestartCount returns the number of times the service has been restarted
func (a *RestartableService) RestartCount() int {
	return a.restartCount
}

// NewRestartableService is used to create a new RestartableService instance.
func NewRestartableService(newService ServiceConstructor) *RestartableService {
	return &RestartableService{
		service:    newService(),
		NewService: newService,
	}
}
