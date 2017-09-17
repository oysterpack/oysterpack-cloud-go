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

package oysterpack_test

import (
	"net"
	"sync"
	"testing"
	"time"
)

type Foo struct{}

func (f Foo) String() string {
	return "Foo"
}

func TestClosingBufferedChannel(t *testing.T) {
	c := make(chan int, 10)
	for i := 0; i < 10; i++ {
		c <- i
	}
	close(c)
	t.Log("closed channel")
	for i := 0; i <= 11; i++ {
		t.Logf("receiving on closed channel [%d]: %v", i, <-c)
	}
}

func TestClosingClosedChannel(t *testing.T) {
	defer func() {
		if p := recover(); p == nil {
			t.Error("closing a closed channel should have triggered a panic")
		}
	}()
	c := make(chan struct{})
	close(c)
	close(c)
}

func TestClosingBlockedChannelOnSend(t *testing.T) {
	msgs := make(chan string)
	wait1 := sync.WaitGroup{}
	wait1.Add(1)
	msgSent := sync.WaitGroup{}
	msgSent.Add(1)
	go func() {
		defer func() {
			switch p := recover().(type) {
			case nil:
				t.Errorf("this should have caused a panic when the channel is closed - message delivery should have failed")
			case error:
				t.Logf("error(%[1]T) : %[1]s", p)
			}
			msgSent.Done()
		}()
		wait1.Done()
		t.Log("msgs <- MSG")
		msgs <- "MSG"
	}()
	wait1.Wait()
	time.Sleep(100 * time.Millisecond)
	t.Log("close(msgs) ...")
	close(msgs)
	t.Log("close(msgs) !!!")
	if msg := <-msgs; msg != "" {
		t.Errorf("Zero value should be received on closed channel : %v", msg)
	}
	msgSent.Wait()
}

func TestGetLocalIP(t *testing.T) {
	t.Logf(GetLocalIP())
}

func GetLocalIP() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return ""
	}
	for _, address := range addrs {
		// check the address type and if it is not a loopback the display it
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String()
			}
		}
	}
	return ""
}
