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
	domainID  DomainID
	appID     AppID
	releaseID ReleaseID

	appInstanceId = InstanceID(nuid.Next())
	startedOn     = time.Now()

	app         = tomb.Tomb{}
	commandChan = make(chan func(), 32)

	logger           zerolog.Logger
	appLogLevel      zerolog.Level
	serviceLogLevels map[ServiceID]zerolog.Level

	services    = make(map[ServiceID]*Service)
	rpcServices = make(map[ServiceID]*RPCService)

	configDir string
)

func submitCommand(f func()) error {
	select {
	case <-app.Dying():
		return ErrAppNotAlive
	case commandChan <- f:
		return nil
	}
}

func Domain() DomainID {
	return domainID
}

// ID returns the AppID which is specified via a command line argument
func ID() AppID {
	return appID
}

// Release returns the application ReleaseID
func Release() ReleaseID {
	return releaseID
}

// InstanceID returns a new unique instance id each time the app is started, i.e., when the process is started.
// The instance id remains for the lifetime of the process
func Instance() InstanceID {
	return appInstanceId
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

// ServiceLogLevel returns the service log level.
// The service log level will use the application log level unless it is overridden via the -service-log-level command line flag
func ServiceLogLevel(id ServiceID) zerolog.Level {
	logLevel, ok := serviceLogLevels[id]
	if ok {
		return logLevel
	}
	return LogLevel()
}

func zerologLevel(logLevel string) zerolog.Level {
	switch strings.ToUpper(logLevel) {
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

	c := make(chan error, 1)
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
				if err := s.Err(); err != nil {
					SERVICE_STOPPING.Log(s.Logger().Error()).Err(err).Msg("stopping")
				} else {
					SERVICE_STOPPING.Log(s.Logger().Info()).Msg("stopping")
				}

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

// UnregisterService will unregister the service for the specified ServiceID
//
// errors:
//	- ErrAppNotAlive
//  - ErrServiceNotRegistered
func UnregisterService(id ServiceID) error {
	c := make(chan error, 1)
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

func logServiceDeath(service *Service) {
	logEvent := SERVICE_STOPPED.Log(service.Logger().Info())
	if err := service.Err(); err != nil {
		logEvent.Err(err)
	}
	logEvent.Msg("stopped")
}

// ServiceIDs returns the ServiceID(s) for the currently registered services
//
// errors:
//	- ErrAppNotAlive
func ServiceIDs() ([]ServiceID, error) {
	c := make(chan []ServiceID, 1)

	if err := submitCommand(func() {
		ids := make([]ServiceID, len(services))
		i := 0
		for id := range services {
			ids[i] = id
			i++
		}
		c <- ids
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

// GetService will lookup the service for the specified ServiceID
//
// errors:
//	- ErrAppNotAlive
func GetService(id ServiceID) (*Service, error) {
	c := make(chan *Service, 1)

	if err := submitCommand(func() {
		c <- services[id]
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

// RPCServiceIDs returns ServiceID(s) for registered RPCService(s)
//
// errors:
//	- ErrAppNotAlive
func RPCServiceIDs() ([]ServiceID, error) {
	c := make(chan []ServiceID, 1)

	if err := submitCommand(func() {
		ids := make([]ServiceID, len(rpcServices))
		i := 0
		for id := range rpcServices {
			ids[i] = id
			i++
		}
		c <- ids
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

// GetRPCService returns the registered *RPCService
//
// errors :
// - ErrAppNotAlive
// - ErrServiceNotRegistered
func GetRPCService(id ServiceID) (*RPCService, error) {
	c := make(chan *RPCService, 1)

	submitCommand(func() {
		service, ok := rpcServices[id]
		if !ok {
			close(c)
			return
		}
		c <- service
	})

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

// Alive returns true if the app is still alive
func Alive() bool {
	return app.Alive()
}

// Kill triggers app shutdown
func Kill() {
	app.Kill(nil)
}

// Dying returns a channel that is used to signal that the app has been killed and shutting down
func Dying() <-chan struct{} {
	return app.Dying()
}

// Dead returns a channel that signals the app has been shutdown
func Dead() <-chan struct{} {
	return app.Dead()
}

func init() {
	var domainIDVar uint64
	flag.Uint64Var(&domainIDVar, "domain-id", 0, "DomainID")

	var appIDVar uint64
	flag.Uint64Var(&appIDVar, "app-id", 0, "AppID")

	var releaseIDVar uint64
	flag.Uint64Var(&releaseIDVar, "release-id", 0, "ReleaseID")

	var logLevelVar string
	flag.StringVar(&logLevelVar, "log-level", "WARN", "[DEBUG,INFO,WARN,ERROR] default = WARN")

	var serviceLogLevelsVar string
	flag.StringVar(&serviceLogLevelsVar, "service-log-level", "", "ServiceID=LogLevel[,ServiceID=LogLevel]")

	flag.StringVar(&configDir, "config-dir", "/run/secrets", "App config directory - default is Docker's secrets dir")

	flag.Parse()

	domainID = DomainID(domainID)
	appID = AppID(appIDVar)
	releaseID = ReleaseID(releaseIDVar)

	initZerolog(logLevelVar)
	initServiceLogLevels(serviceLogLevelsVar)

	runAppServer()
	runRPCAppServer()
	startMetricsHttpReporter()
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

	loggerCtx := log.Logger.With().
		Uint64("domain", uint64(domainID)).
		Uint64("app", uint64(appID)).
		Uint64("release", uint64(releaseID)).
		Str("instance", string(appInstanceId))
	if appLogLevel == zerolog.DebugLevel {
		logger = loggerCtx.Logger().Level(zerolog.DebugLevel)
	} else {
		logger = loggerCtx.Logger().Level(zerolog.InfoLevel)
	}
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
		signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
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

	// Kill each registered service
	for _, service := range services {
		service.Kill(nil)
		SERVICE_KILLED.Log(service.Logger().Info()).Msg("killed")
	}

	// Wait until all registered srevices are shutdown.
	// If the service takes longer than 10 seconds to shutdown, then log a warning and move on.
	// Wait a maximum of 2 minute for the app to shutdown, after which we move on. We don't want to hang the whole app
	// because we are waiting for a service to shutdown.
	maxWaitTime := time.NewTicker(time.Minute * 2)
	defer maxWaitTime.Stop()
SERVICE_LOOP:
	for _, service := range services {
		ticker := time.NewTicker(time.Second * 10)
		defer ticker.Stop()
		for {
			select {
			case <-service.Dead():
				logServiceDeath(service)
				continue SERVICE_LOOP
			case <-ticker.C:
				ticker.Stop()
				SERVICE_STOPPING_TIMEOUT.Log(service.Logger().Warn()).Msg("service is taking too long to stop")
				continue SERVICE_LOOP
			case <-maxWaitTime.C:
				APP_STOPPING_TIMEOUT.Log(Logger().Warn()).Msg("app is taking too long to stop")
			}
		}
	}

	services = make(map[ServiceID]*Service)
	rpcServices = make(map[ServiceID]*RPCService)
}
