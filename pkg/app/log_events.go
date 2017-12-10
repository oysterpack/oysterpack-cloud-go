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

// log events
const (
	APP_STARTED          = LogEventID(0xa482715a50d67a5f)
	APP_STOPPING         = LogEventID(0xbcdae48c0cb8936e)
	APP_STOPPING_TIMEOUT = LogEventID(0xaa7744d8a20d857a)
	APP_STOPPED          = LogEventID(0xdd0c7775e42d7841)
	APP_RESET            = LogEventID(0xee317bbb0fe0fafe)

	// it is the service's responsibility to log the following Service lifecycle events
	SERVICE_STARTING = LogEventID(0xa3c3eb887d09f9aa)
	SERVICE_STARTED  = LogEventID(0xc27a49a4e5a2a502)

	// the below events are logged by the app using the service logger
	SERVICE_KILLED           = LogEventID(0x85adf7d70dcef626)
	SERVICE_STOPPING         = LogEventID(0x85adbb661141efce)
	SERVICE_STOPPING_TIMEOUT = LogEventID(0x8ed2e400ce585f16)
	SERVICE_STOPPED          = LogEventID(0xfd843c25ce81f841)
	SERVICE_REGISTERED       = LogEventID(0xd8f25797ffa58858)
	SERVICE_UNREGISTERED     = LogEventID(0xa611d10b1dfc880d)

	METRICS_SERVICE_CONFIG_ERROR = LogEventID(0x83ad8592584d0930)

	METRICS_HTTP_REPORTER_SHUTDOWN_ERROR                 = LogEventID(0xab355f6933e9b3fe)
	METRICS_HTTP_REPORTER_START_ERROR                    = LogEventID(0xa1e82667eef867ff)
	METRICS_HTTP_REPORTER_SHUTDOWN_WHILE_SERVICE_RUNNING = LogEventID(0xf8c787086a35cd78)

	METRICS_HTTP_SERVER_STARTING = LogEventID(0xbec4ba15afa97c2b)
	METRICS_HTTP_SERVER_STARTED  = LogEventID(0x8fab3d9ef5011368)
	METRICS_HTTP_SERVER_STOPPED  = LogEventID(0xefd0f72bffc636f6)

	CONFIG_LOADING_ERR = LogEventID(0x83e927cb25aaa032)

	CAPNP_ERR = LogEventID(0x823407a4ed427f33)

	ZERO_HEALTHCHECKS      = LogEventID(0x8c805853a70ec666)
	HEALTHCHECK_REGISTERED = LogEventID(0xb5f3142d410a2375)
	HEALTHCHECK_PAUSED     = LogEventID(0xc7a3b54188e75210)
	HEALTHCHECK_RESUMED    = LogEventID(0xc4dc7b2938caf3a9)
	HEALTHCHECK_RESULT     = LogEventID(0xa68e0475cc1839be)
)
