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

import (
	"github.com/rs/zerolog"
	"gopkg.in/tomb.v2"
)

func NewService(id ServiceID) *Service {
	if id == 0 {
		panic(ErrServiceIDZero)
	}
	logLevel := Services.LogLevel(id)
	return &Service{id: id, logger: Logger().With().Uint64("svc", uint64(id)).Logger().Level(logLevel), logLevel: logLevel}
}

type Service struct {
	tomb.Tomb

	id ServiceID

	logLevel zerolog.Level
	logger   zerolog.Logger
}

func (a *Service) ID() ServiceID {
	return a.id
}

func (a *Service) Logger() zerolog.Logger {
	return a.logger
}

func (a *Service) LogLevel() zerolog.Level {
	return a.logLevel
}
