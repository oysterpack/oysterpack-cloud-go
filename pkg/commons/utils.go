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

package commons

import (
	"github.com/rs/zerolog/log"
)

// int max / min values
const (
	MaxUint = ^uint(0)
	MinUint = 0
	MaxInt  = int(MaxUint >> 1)
	MinInt  = -MaxInt - 1
)

// IgnorePanic simply calls recover()
// It is intended to use in defer() functions to explicitly specify the intent.
// If debug logging is enabled, then ignored panics will be logged
func IgnorePanic() {
	if p := recover(); p != nil {
		log.Debug().Msgf("IgnorePanic : %v", p)
	}
}

// CloseQuietly quietly close the channel
// A common use case for such a channel is for event notifications, e.g., stop triggers
func CloseQuietly(c chan struct{}) {
	defer IgnorePanic()
	close(c)
}
