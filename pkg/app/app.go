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

	"strconv"

	"github.com/nats-io/nuid"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"gopkg.in/tomb.v2"
)

// app vars
var (
	appID     AppID
	releaseID ReleaseID

	appInstanceId = InstanceID(nuid.Next())
	startedOn     = time.Now()

	app         tomb.Tomb
	commandChan chan func()

	logger           zerolog.Logger
	appLogLevel      zerolog.Level
	serviceLogLevels map[ServiceID]zerolog.Level

	services map[ServiceID]*Service
)

func submitCommand(f func()) error {
	select {
	case <-app.Dying():
		return ErrAppNotAlive
	case commandChan <- f:
		return nil
	}
}

// ID returns the AppID which is specified via a command line argument
func ID() AppID {
	return appID
}

// InstanceID returns a new unique instance id each time the app is started, i.e., when the process is started.
// The instance id remains for the lifetime of the process
func Instance() InstanceID {
	return appInstanceId
}

// Release returns the application ReleaseID
func Release() ReleaseID {
	return releaseID
}

//
func StartedOn() time.Time {
	return startedOn
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
	return appLogLevel
}

func ServiceLogLevel(id ServiceID) zerolog.Level {
	logLevel, ok := serviceLogLevels[id]
	if ok {
		return logLevel
	}
	return LogLevel()
}

func zerologLevel(logLevel string) zerolog.Level {
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

	c := make(chan error)

	err := submitCommand(func() {
		if _, ok := services[s.id]; ok {
			c <- ErrServiceAlreadyRegistered
		}
		services[s.id] = s
		SERVICE_REGISTERED.Log(s.logger.Info()).Msg("registered")
		// signal that the service registration was completed successfully
		close(c)

		// watch the service
		// when it dies, then unregister it
		s.Go(func() error {
			select {
			case <-app.Dying():
				return nil
			case <-s.Dying():
				SERVICE_STOPPING.Log(s.Logger().Info()).Msg("stopping")
				UnregisterService(s.id)
				return nil
			}
		})
	})

	if err != nil {
		return err
	}

	select {
	case <-app.Dying():
		return ErrAppNotAlive
	case err := <-c:
		return err
	}
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
	c := make(chan []ServiceID)

	if err := submitCommand(func() {
		ids := make([]ServiceID, len(services))
		i := 0
		for id := range services {
			ids[i] = id
			i++
		}
		select {
		case <-app.Dying():
		case c <- ids:
		}
	}); err != nil {
		return nil, err
	}

	select {
	case <-app.Dying():
		return nil, ErrAppNotAlive
	case ids := <-c:
		return ids, nil
	}
}

// UnregisterService will unregister the service for the specified ServiceID
//
// errors:
//	- ErrAppNotAlive
//  - ErrServiceNotRegistered
func UnregisterService(id ServiceID) error {
	c := make(chan error)
	if err := submitCommand(func() {
		service, exists := services[id]
		if !exists {
			c <- ErrServiceNotRegistered
			return
		}

		// log an event when the service is dead
		app.Go(func() error {
			select {
			case <-app.Dying():
				return nil
			case <-service.Dead():
				logServiceDeath(service)
				return nil
			}
		})

		delete(services, id)
		SERVICE_UNREGISTERED.Log(service.Logger().Info()).Msg("unregistered")
		c <- nil
	}); err != nil {
		return err
	}

	select {
	case <-app.Dying():
		return ErrAppNotAlive
	case err := <-c:
		return err
	}
}

// GetService will lookup the service for the specified ServiceID
//
// errors:
//	- ErrAppNotAlive
func GetService(id ServiceID) (*Service, error) {
	c := make(chan *Service)

	if err := submitCommand(func() {
		select {
		case <-app.Dying():
		case c <- services[id]:
		}
	}); err != nil {
		return nil, err
	}

	select {
	case <-app.Dying():
		return nil, ErrAppNotAlive
	case svc := <-c:
		if svc == nil {
			return nil, ErrServiceNotRegistered
		}
		return svc, nil
	}
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

// Alive returns true if the app is still alive
func Alive() bool {
	return app.Alive()
}

// Kill triggers app shutdown
func Kill() {
	app.Kill(nil)
}

func init() {
	var appIDVar uint64
	flag.Uint64Var(&appIDVar, "app-id", 0, "AppID")

	var releaseIDVar uint64
	flag.Uint64Var(&releaseIDVar, "release-id", 0, "AppID")

	var logLevelVar string
	flag.StringVar(&logLevelVar, "log-level", "WARN", "valid log levels [DEBUG,INFO,WARN,ERROR] default = WARN")

	var serviceLogLevelsVar string
	flag.StringVar(&serviceLogLevelsVar, "service-log-level", "", "ServiceID=LogLevel[,ServiceID=LogLevel]")

	flag.Parse()

	logLevelVar = strings.ToUpper(logLevelVar)
	appID = AppID(appIDVar)
	releaseID = ReleaseID(releaseIDVar)

	app = tomb.Tomb{}
	services = make(map[ServiceID]*Service)

	commandChan = make(chan func())
	initZerolog(logLevelVar)
	initServiceLogLevels(serviceLogLevelsVar)
	runAppServer()
}

func initZerolog(logLevel string) {
	// log with nanosecond precision time
	zerolog.TimeFieldFormat = time.RFC3339Nano

	// set the global log level
	appLogLevel = zerologLevel(logLevel)
	log.Logger = log.Logger.Level(appLogLevel)

	// redirects go's std log to zerolog
	stdlog.SetFlags(0)
	stdlog.SetOutput(log.Logger)

	logger = log.Logger.With().
		Uint64("app", uint64(appID)).
		Uint64("release", uint64(releaseID)).
		Str("instance", string(appInstanceId)).
		Logger().
		Level(zerolog.InfoLevel)
}

func initServiceLogLevels(serviceLogLevelsFlag string) {
	serviceLogLevelsFlag = strings.TrimSpace(serviceLogLevelsFlag)
	if serviceLogLevelsFlag == "" {
		return
	}

	serviceLogLevels = make(map[ServiceID]zerolog.Level)
	for _, serviceLogLevel := range strings.Split(serviceLogLevelsFlag, ",") {
		tokens := strings.Split(serviceLogLevel, "=")
		if len(tokens) != 2 {
			logger.Warn().Str("service-log-level", serviceLogLevel).Msg("invalid service log level flag")
			continue
		}

		serviceId, err := strconv.ParseUint(tokens[0], 0, 64)
		if err != nil {
			logger.Warn().Str("service-log-level", serviceLogLevel).Err(err).Msg("invalid service id")
			continue
		}
		serviceLogLevels[ServiceID(serviceId)] = zerologLevel(tokens[1])
	}

}

func runAppServer() {
	app.Go(func() error {
		APP_STARTED.Log(logger.Info()).Msg("started")

		sigs := make(chan os.Signal, 1)
		signal.Notify(sigs, syscall.SIGTERM)
		for {
			select {
			case <-sigs:
				app.Kill(nil)
			case <-app.Dying():
				shutdown()
				return nil
			case f := <-commandChan:
				f()
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
