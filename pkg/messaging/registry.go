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

package messaging

import "sync"

// Registry is the global ClientRegistry
var Registry = NewClientRegistry()

// ClientRegistry
type ClientRegistry interface {

	// MustRegister will panic if a Client is already registered for the same cluster
	MustRegister(client Client)

	// Unregister will remove the Client for the specified cluster and is returned
	Unregister(cluster ClusterName) Client

	// Client will return the Client registered for the specified cluster
	Client(cluster ClusterName) Client

	// Clusters returns the cluster names for all the registered messaging Client(s)
	Clusters() []ClusterName
}

func NewClientRegistry() ClientRegistry {
	return &registry{clients: make(map[ClusterName]Client)}
}

type registry struct {
	sync.RWMutex
	clients map[ClusterName]Client
}

func (a *registry) MustRegister(client Client) {
	a.Lock()
	defer a.Unlock()
	if a.clients[client.Cluster()] != nil {
		logger.Panic().Err(ErrClientAlreadyRegsiteredForSameCluster).Msg("")
	}
	a.clients[client.Cluster()] = client
}

func (a *registry) Unregister(cluster ClusterName) Client {
	a.Lock()
	defer a.Unlock()
	client := a.clients[cluster]
	delete(a.clients, cluster)
	return client
}

func (a *registry) Client(cluster ClusterName) Client {
	a.RLock()
	defer a.RUnlock()
	return a.clients[cluster]
}

func (a *registry) Clusters() []ClusterName {
	a.RLock()
	defer a.RUnlock()
	names := make([]ClusterName, len(a.clients))
	i := 0
	for name := range a.clients {
		names[i] = name
		i++
	}
	return names
}
