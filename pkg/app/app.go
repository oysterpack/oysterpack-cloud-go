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

	"os"
	"os/signal"
	"syscall"

	"strings"

	"github.com/nats-io/nuid"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"gopkg.in/tomb.v2"
)

// app vars
var (
	logLevel string
	appID    uint64

	appInstanceId = InstanceID(nuid.Next())
	createdOn     = time.Now()

	app tomb.Tomb

	logger zerolog.Logger

	services map[ServiceID]*Service

	registerServiceChan      chan registerServiceRequest
	registeredServiceIdsChan chan registeredServiceIdsRequest
	unregisterServiceChan    chan ServiceID
	getServiceChan           chan getServiceRequest
)

type InstanceID string

// ID returns the AppID which is specified via a command line argument
func ID() AppID {
	return AppID(appID)
}

func InstanceId() InstanceID {
	return appInstanceId
}

func CreatedOn() time.Time {
	return createdOn
}

// Logger returns the app logger
func Logger() zerolog.Logger {
	return logger
}

// LogLevel returns the application log level.
// If the command line is parsed, then the -loglevel flag will be inspected. Valid values for -loglevel are : [DEBUG,INFO,WARN,ERROR]
// If not specified on the command line, then the defauly value is INFO.
// The log level is used to configure the log level for loggers returned via NewTypeLogger() and NewPackageLogger().
// It is also used to initialize zerolog's global logger level.
func LogLevel() zerolog.Level {
	switch logLevel {
	case "DEBUG":
		return zerolog.DebugLevel
	case "INFO":
		return zerolog.InfoLevel
	case "WARN":
		return zerolog.WarnLevel
	case "ERROR":
		return zerolog.ErrorLevel
	default:
		return zerolog.WarnLevel
	}
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

	// signal that the service registration was completed successfully
	close(req.response)

	SERVICE_REGISTERED.Log(req.Service.Logger().Info()).Msg("registered")

	// watch the service
	// when it dies, then unregister it
	req.Service.Go(func() error {
		select {
		case <-app.Dying():
			return nil
		case <-req.Service.Dying():
			SERVICE_STOPPING.Log(req.Service.Logger().Info()).Msg("stopping")
			select {
			case <-app.Dying():
				return nil
			case unregisterServiceChan <- req.Service.id:
				app.Go(func() error {
					select {
					case <-app.Dying():
						return nil
					case <-req.Service.Dead():
						logServiceDeath(req.Service)
						return nil
					}
				})
				return nil
			}
		}
	})
}

func logServiceDeath(service *Service) {
	logEvent := SERVICE_STOPPED.Log(service.Logger().Info())
	if err := service.Err(); err != nil {
		logEvent.Err(err)
	}
	logEvent.Msg("stopped")
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

// Reset is exposed only for testing purposes.
// Reset will kill the app, and then restart the app server.
func Reset() {
	app.Kill(nil)
	app.Wait()

	app = tomb.Tomb{}
	runAppServer()
	APP_RESET.Log(logger.Info()).Msg("reset")
}

// Kill triggers app shutdown
func Kill() {
	app.Kill(nil)
}

func init() {
	flag.Uint64Var(&appID, "app-id", 0, "AppID")
	flag.StringVar(&logLevel, "log-level", "WARN", "valid log levels [DEBUG,INFO,WARN,ERROR] default = WARN")
	flag.Parse()
	logLevel = strings.ToUpper(logLevel)

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
	log.Logger = log.Logger.Level(LogLevel())

	// redirects go's std log to zerolog
	stdlog.SetFlags(0)
	stdlog.SetOutput(log.Logger)

	logger = log.Logger.With().Uint64("app", appID).Str("instance", string(appInstanceId)).Logger().Level(zerolog.InfoLevel)
	APP_STARTED.Log(logger.Info()).Msg("started")
}

func runAppServer() {
	app.Go(func() error {
		sigs := make(chan os.Signal, 1)
		signal.Notify(sigs, syscall.SIGTERM)
		for {
			select {
			case <-sigs:
				app.Kill(nil)
			case <-app.Dying():
				shutdown()
				return nil
			case req := <-registerServiceChan:
				registerService(req)
			case req := <-registeredServiceIdsChan:
				registeredServiceIds(req)
			case id := <-unregisterServiceChan:
				if service, ok := services[id]; ok {
					SERVICE_UNREGISTERED.Log(service.Logger().Info()).Msg("unregistered")
				}
				delete(services, id)
			case req := <-getServiceChan:
				select {
				case req.response <- services[req.ServiceID]:
				case <-app.Dying():
				}
			}
		}
	})
}

func shutdown() {
	APP_STOPPING.Log(logger.Info()).Msg("stopping")
	defer APP_STOPPED.Log(logger.Info()).Msg("stopped")

	for _, service := range services {
		service.Kill(nil)
		SERVICE_KILLED.Log(service.Logger().Info()).Msg("killed")
	}

	ticker := time.NewTicker(time.Second * 10)
	defer ticker.Stop()
SERVICE_LOOP:
	for _, service := range services {
		for {
			select {
			case <-service.Dead():
				logServiceDeath(service)
				continue SERVICE_LOOP
			case <-ticker.C:
				SERVICE_STOPPING.Log(service.Logger().Warn()).Msg("waiting for service to stop")
			}
		}
	}

	services = make(map[ServiceID]*Service)
}
