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

// ServiceTicket represents a ticket issued to a user waiting for a service
type ServiceTicket struct {
	// the type of service the user is waiting for
	Interface
	// used to deliver the Client to the user
	channel chan Client

	// when the ticket was created
	time.Time
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
		counts[ticket.Interface]++
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
		func(ticket *ServiceTicket) {
			defer commons.IgnorePanic()
			close(ticket.channel)
		}(ticket)
	}
}
