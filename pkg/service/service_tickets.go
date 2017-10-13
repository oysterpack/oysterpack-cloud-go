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

import (
	"sync"
	"time"

	"github.com/oysterpack/oysterpack.go/pkg/commons"
)

// NewServiceTicket creates a new ServiceTicket for the specified service Interface
func NewServiceTicket(serviceInterface Interface) *ServiceTicket {
	return &ServiceTicket{ServiceInterface: serviceInterface, channel: make(chan Client, 1), Time: time.Now()}
}

// ServiceTicket represents a ticket issued to a user waiting for a service
type ServiceTicket struct {
	mutex sync.Mutex

	// the type of service the user is waiting for
	ServiceInterface Interface
	// used to deliver the Client to the user
	channel chan Client

	// when the ticket was created
	Time time.Time
}

func (a *ServiceTicket) Close(client Client) {
	a.mutex.Lock()
	defer a.mutex.Unlock()
	defer commons.IgnorePanic()
	if client != nil {
		a.channel <- client
	}
	close(a.channel)
}

// used to keep track of users who are waiting on services
type serviceTickets struct {
	// used to keep track of users who are waiting on services
	ticketsMutex sync.RWMutex
	tickets      []*ServiceTicket
}

// ServiceTicketCounts returns the number of tickets that have been issued per service
func (a *serviceTickets) ServiceTicketCounts() map[Interface]int {
	a.ticketsMutex.RLock()
	defer a.ticketsMutex.RUnlock()
	counts := make(map[Interface]int)
	for _, ticket := range a.tickets {
		counts[ticket.ServiceInterface]++
	}
	return counts
}

func (a *serviceTickets) add(ticket *ServiceTicket) {
	a.ticketsMutex.Lock()
	a.tickets = append(a.tickets, ticket)
	a.ticketsMutex.Unlock()
}

func (a *serviceTickets) deleteServiceTicket(ticket *ServiceTicket) {
	a.ticketsMutex.Lock()
	defer a.ticketsMutex.Unlock()
	for i, elem := range a.tickets {
		if elem == ticket {
			a.tickets[i] = a.tickets[len(a.tickets)-1] // Replace it with the last one.
			a.tickets = a.tickets[:len(a.tickets)-1]
			return
		}
	}
}

func (a *serviceTickets) closeAllServiceTickets() {
	a.ticketsMutex.RLock()
	defer a.ticketsMutex.RUnlock()
	for _, ticket := range a.tickets {
		ticket.Close(nil)
	}
}

// checks if any open service tickets can be closed, i.e., it checks if the services that tickets are waiting on are available in the registry
func (a *serviceTickets) checkServiceTickets(registry Registry) {
	a.ticketsMutex.RLock()
	defer a.ticketsMutex.RUnlock()
	for _, ticket := range a.tickets {
		serviceClient := registry.ServiceByType(ticket.ServiceInterface)
		if serviceClient != nil {
			go ticket.Close(serviceClient)
			go a.deleteServiceTicket(ticket)
		}
	}
}
