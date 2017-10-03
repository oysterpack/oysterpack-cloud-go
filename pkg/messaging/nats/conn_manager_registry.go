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

package nats

import (
	"sync"

	"github.com/oysterpack/oysterpack.go/pkg/messaging"
	"github.com/oysterpack/oysterpack.go/pkg/service"
)

// ConnManagerRegistry is a registry for ConnManager instances
type ConnManagerRegistry interface {
	// Clusters returns the names of all registered ConnManager(s)
	Clusters() []messaging.ClusterName

	// ConnManager returns a registered ConnManager
	// false indicates that there is no ConnManager registered with the specified name
	ConnManager(cluster messaging.ClusterName) (ConnManager, bool)

	// MustRegister panics if a ConnManager is already registered with the same ClusterName
	MustRegister(connManager ConnManager)

	// CloseAll will close all connections for all registered ConnManager(s)
	CloseAll()
}

type connManagerRegistry struct {
	*service.RestartableService

	sync.RWMutex
	registry map[messaging.ClusterName]ConnManager
}

func (a *connManagerRegistry) Clusters() []messaging.ClusterName {
	a.RLock()
	defer a.RUnlock()
	names := make([]messaging.ClusterName, len(a.registry))
	i := 0
	for name := range a.registry {
		names[i] = name
		i++
	}
	return names
}

func (a *connManagerRegistry) ConnManager(cluster messaging.ClusterName) (connMgr ConnManager, exists bool) {
	a.RLock()
	defer a.RUnlock()
	connMgr, exists = a.registry[cluster]
	return
}

func (a *connManagerRegistry) MustRegister(connManager ConnManager) {
	a.Lock()
	defer a.Unlock()
	if _, exists := a.registry[connManager.Cluster()]; exists {
		logger.Panic().Msgf("A ConnManager is already registered with name %q", connManager.Cluster())
	}
	a.registry[connManager.Cluster()] = connManager
}

func (a *connManagerRegistry) CloseAll() {
	a.RLock()
	defer a.RUnlock()
	for _, c := range a.registry {
		c.CloseAll()
	}
}
