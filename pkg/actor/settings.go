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

package actor

type Settings struct {
	InstanceFactory
	SupervisorStrategy
	Channels map[string]ChannelSettings
}

type ChannelSettings struct {
	channel        string
	bufSize        int
	middleware     []Middleware
	messageFactory MessageFactory
}

func (a *ChannelSettings) Channel() string {
	return a.channel
}

func (a *ChannelSettings) BufSize() int {
	return a.bufSize
}

func (a *ChannelSettings) MessageInterceptor(finalInterceptor MessageInterceptor) MessageInterceptor {
	if len(a.middleware) == 0 {
		return func(system *System, msg *Envelope) (*Envelope, error) {
			if err := a.messageFactory.Validate(msg.message); err != nil {
				return nil, err
			}
			return msg, nil
		}
	}

	middleware := make([]Middleware, len(a.middleware)+1)

	middleware[0] = func(next MessageInterceptor) MessageInterceptor {
		return func(system *System, msg *Envelope) (*Envelope, error) {
			if err := a.messageFactory.Validate(msg.message); err != nil {
				return nil, err
			}
			return next(system, msg)
		}
	}

	for i := 0; i < len(a.middleware); i++ {
		middleware[i+1] = a.middleware[i]
	}

	interceptor := middleware[len(middleware)-1](finalInterceptor)
	for i := len(middleware) - 2; i >= 0; i-- {
		interceptor = middleware[i](interceptor)
	}
	return interceptor
}

func (a *ChannelSettings) MessageFactory() MessageFactory {
	return a.messageFactory
}

type MessageInterceptor func(system *System, msg *Envelope) (*Envelope, error)

type Middleware func(next MessageInterceptor) MessageInterceptor
