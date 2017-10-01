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

package nats

import (
	"errors"
	"strings"

	"github.com/nats-io/go-nats"
)

// ConnectUrl - if not specified, then the default will be used : "nats://localhost:4222"
func ConnectUrl(url string) nats.Option {
	return func(options *nats.Options) error {
		url = strings.TrimSpace(url)
		if len(url) == 0 {
			return errors.New("url cannot be blank")
		}
		options.Url = url
		return nil
	}
}
