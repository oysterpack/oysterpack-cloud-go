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

package net

import (
	"errors"
	"net"
)

var localIP *net.IP

// LocalIP returns the cached local IP
func LocalIP() *net.IP {
	if localIP != nil {
		return localIP
	}
	localIP, _ := GetLocalIP()
	return localIP
}

// GetLocalIP looks up the local IP and returns the first non-loopback IP
func GetLocalIP() (*net.IP, error) {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return nil, err
	}
	for _, address := range addrs {
		// check the address type and if it is not a loopback the display it
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			return &ipnet.IP, nil
		}
	}
	return nil, errors.New("unable to obtain non loopback local ip address")
}
