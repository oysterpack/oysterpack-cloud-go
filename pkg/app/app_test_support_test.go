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

import "gopkg.in/tomb.v2"

// Reset is exposed only for testing purposes.
// Reset will kill the app, and then restart the app server.
func Reset() {
	app.Kill(nil)
	app.Wait()

	app = tomb.Tomb{}
	runAppServer()
	runRPCAppServer()
	metricsRegistry = NewMetricsRegistry(true)
	startMetricsHttpReporter()
	APP_RESET.Log(logger.Info()).Msg("reset")
}
