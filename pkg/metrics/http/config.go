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

package http

// ServiceConfig defines the config for the metrics service
type ServiceConfig struct {
	httpPort int
}

// HTTPPort returns the http port for the metrics http server for exposing metrics to prometheus on /metrics
func (a *ServiceConfig) HTTPPort() int {
	return a.httpPort
}

// Config exposes service config - used by http clients to obtain the HTTP port
var Config = &ServiceConfig{httpPort: 4444}
