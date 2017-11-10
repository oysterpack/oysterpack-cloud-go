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

const (
	APP_STARTED  = LogEventID(0xa482715a50d67a5f)
	APP_STOPPING = LogEventID(0xbcdae48c0cb8936e)
	APP_STOPPED  = LogEventID(0xdd0c7775e42d7841)
	APP_RESET    = LogEventID(0xee317bbb0fe0fafe)

	SERVICE_FAILURE = LogEventID(0xd8f25797ffa58858)
)
