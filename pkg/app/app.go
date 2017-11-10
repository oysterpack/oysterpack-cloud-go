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
	"flag"

	"time"

	stdlog "log"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"gopkg.in/tomb.v2"
)

// app vars
var (
	app tomb.Tomb

	logger zerolog.Logger

	services map[ServiceID]*Service

	registerServiceChan      chan registerServiceRequest
	registeredServiceIdsChan chan registeredServiceIdsRequest
	unregisterServiceChan    chan ServiceID
	getServiceChan           chan getServiceRequest
)

func Log() zerolog.Logger {
	return logger
}

// RegisterService will register the service with the app.
//
// errors:
//	- ErrAppNotAlive
//	- ErrServiceAlreadyRegistered
func RegisterService(s *Service) error {
	if s == nil {
		return ErrServiceNil
	}
	if !s.Alive() {
		return ErrServiceNotAlive
	}
	req := registerServiceRequest{s, make(chan error)}
	select {
	case <-app.Dying():
		return ErrAppNotAlive
	case registerServiceChan <- req:
		select {
		case <-app.Dying():
			return ErrAppNotAlive
		case err := <-req.response:
			return err
		}
	}
}

type registerServiceRequest struct {
	*Service
	response chan error
}

func registerService(req registerServiceRequest) {
	if _, ok := services[req.Service.id]; ok {
		req.response <- ErrServiceAlreadyRegistered
	}
	services[req.Service.id] = req.Service
	close(req.response)
	req.Service.Go(func() error {
		select {
		case <-app.Dying():
			return nil
		case <-req.Service.Dying():
			select {
			case <-app.Dying():
				return nil
			case unregisterServiceChan <- req.Service.id:
				return nil
			}
		}
	})
}

// RegisteredServiceIds returns the ServiceID(s) for the currently registered services
//
// errors:
//	- ErrAppNotAlive
func RegisteredServiceIds() ([]ServiceID, error) {
	req := registeredServiceIdsRequest{make(chan []ServiceID)}
	select {
	case <-app.Dying():
		return nil, ErrAppNotAlive
	case registeredServiceIdsChan <- req:
		select {
		case <-app.Dying():
			return nil, ErrAppNotAlive
		case ids := <-req.response:
			return ids, nil
		}
	}
}

type registeredServiceIdsRequest struct {
	response chan []ServiceID
}

func registeredServiceIds(req registeredServiceIdsRequest) {
	ids := make([]ServiceID, len(services))
	i := 0
	for id := range services {
		ids[i] = id
		i++
	}
	select {
	case <-app.Dying():
		return
	case req.response <- ids:
	}
}

// UnregisterService will unregister the service for the specified ServiceID
//
// errors:
//	- ErrAppNotAlive
func UnregisterService(id ServiceID) error {
	select {
	case <-app.Dying():
		return ErrAppNotAlive
	case unregisterServiceChan <- id:
		return nil
	}
}

// GetService will lookup the service for the specified ServiceID
//
// errors:
//	- ErrAppNotAlive
func GetService(id ServiceID) (*Service, error) {
	req := getServiceRequest{id, make(chan *Service)}
	select {
	case <-app.Dying():
		return nil, ErrAppNotAlive
	case getServiceChan <- req:
		select {
		case <-app.Dying():
			return nil, ErrAppNotAlive
		case svc := <-req.response:
			return svc, nil
		}
	}
}

type getServiceRequest struct {
	ServiceID
	response chan *Service
}

// command line flags
var (
	logLevel = flag.String("log-level", "WARN", "valid log levels [DEBUG,INFO,WARN,ERROR] default = INFO")
	appId    = flag.Uint64("app-id", 0, "valid log levels [DEBUG,INFO,WARN,ERROR] default = INFO")
)

// Reset is exposed only for testing purposes.
// Reset will kill the app, and then restart the app server.
func Reset() {
	app.Kill(nil)
	app.Wait()

	app = tomb.Tomb{}
	runAppServer()
	APP_RESET.Log(logger.Info()).Msg("reset")
}

func init() {
	flag.Parse()

	app = tomb.Tomb{}
	services = make(map[ServiceID]*Service)

	makeChans()
	initZerolog()
	runAppServer()
}

func makeChans() {
	registerServiceChan = make(chan registerServiceRequest)
	registeredServiceIdsChan = make(chan registeredServiceIdsRequest)
	unregisterServiceChan = make(chan ServiceID)
	getServiceChan = make(chan getServiceRequest)
}

func initZerolog() {
	// log with nanosecond precision time
	zerolog.TimeFieldFormat = time.RFC3339Nano

	// set the global log level
	log.Logger = log.Logger.Level(LoggingLevel())

	// redirects go's std log to zerolog
	stdlog.SetFlags(0)
	stdlog.SetOutput(log.Logger)

	logger = log.Logger.With().Uint64("app", *appId).Logger().Level(zerolog.InfoLevel)
	APP_STARTED.Log(logger.Info()).Msg("app started")
}

func runAppServer() {
	app.Go(func() error {
		for {
			select {
			case <-app.Dying():
				shutdown()
				return nil
			case req := <-registerServiceChan:
				registerService(req)
			case req := <-registeredServiceIdsChan:
				registeredServiceIds(req)
			case id := <-unregisterServiceChan:
				delete(services, id)
			case req := <-getServiceChan:
				select {
				case <-app.Dying():
				case req.response <- services[req.ServiceID]:
				}
			}
		}
	})
}

func shutdown() {
	APP_STOPPING.Log(logger.Info()).Msg("app stopping")
	defer APP_STOPPED.Log(logger.Info()).Msg("app stopped")

	for _, service := range services {
		service.Kill(nil)
	}

	for _, service := range services {
		if err := service.Wait(); err != nil {
			SERVICE_FAILURE.Log(logger.Error()).Err(err).Msg("service failure during app shutdown")
		}
	}

	services = make(map[ServiceID]*Service)
}
