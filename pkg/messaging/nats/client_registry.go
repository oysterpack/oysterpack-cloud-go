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

type clientRegistry struct {
	*service.RestartableService

	sync.RWMutex
	registry map[messaging.ClusterName]messaging.Client
}

func (a *clientRegistry) Client(cluster messaging.ClusterName) messaging.Client {
	a.RLock()
	defer a.RUnlock()
	return a.registry[cluster]
}

func (a *clientRegistry) Clusters() []messaging.ClusterName {
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

func (a *clientRegistry) MustRegister(client messaging.Client) {
	a.Lock()
	defer a.Unlock()
	if _, exists := a.registry[client.Cluster()]; exists {
		logger.Panic().Msgf("A ConnManager is already registered with name %q", client.Cluster())
	}
	a.registry[client.Cluster()] = client
}

func (a *clientRegistry) Unregister(cluster messaging.ClusterName) messaging.Client {
	a.Lock()
	defer a.Unlock()
	client := a.registry[cluster]
	delete(a.registry, cluster)
	return client
}

func (a *clientRegistry) Destroy(ctx *service.Context) error {
	a.Lock()
	defer a.Unlock()
	keys := make([]messaging.ClusterName, len(a.registry))
	i := 0
	for key, c := range a.registry {
		c.CloseAllConns()
		keys[i] = key
		i++
	}
	for _, key := range keys {
		delete(a.registry, key)
	}
	return nil
}
