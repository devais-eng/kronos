package kronos

import (
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"runtime"
	"time"

	"devais.it/kronos/internal/pkg/telemetry"
	"devais.it/kronos/internal/pkg/util"

	"devais.it/kronos/internal/pkg/api/dbus"
	"devais.it/kronos/internal/pkg/api/http"
	"devais.it/kronos/internal/pkg/build"
	"devais.it/kronos/internal/pkg/config"
	"devais.it/kronos/internal/pkg/db"
	"devais.it/kronos/internal/pkg/logging"
	"devais.it/kronos/internal/pkg/prometheus"
	"devais.it/kronos/internal/pkg/sync"
	"devais.it/kronos/internal/pkg/version"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

const startupBanner = `
  _  __                          
 | |/ /                          
 | ' / _ __ ___  _ __   ___  ___ 
 |  < | '__/ _ \| '_ \ / _ \/ __|
 | . \| | | (_) | | | | (_) \__ \  devAIs S.R.L.
 |_|\_\_|  \___/|_| |_|\___/|___/  Version: %s
`

func printBanner() {
	fmt.Printf(startupBanner, version.GetVersionString(false))
	fmt.Println()
}

func run(configFile string, verbose bool) int {
	printBanner()

	startTime := time.Now()
	conf, err := config.Parse(viper.GetViper(), configFile)
	confParseTime := time.Since(startTime)
	if err != nil {
		logging.Panic(err, "Failed to load configuration")
	}

	if verbose {
		conf.Logging.Level = log.TraceLevel
	}

	err = logging.Setup(&conf.Logging)
	if err != nil {
		logging.Panic(err, "Failed to setup logging")
	}

	log.Infof("Configuration loaded [%v]", confParseTime)

	if log.IsLevelEnabled(log.TraceLevel) {
		config.PrintDebug(viper.GetViper())
	}

	if conf.MaxProcs > 0 {
		oldProcs := runtime.GOMAXPROCS(conf.MaxProcs)
		log.Debugf("Set GOMAXPROCS from %d to %d", oldProcs, conf.MaxProcs)
	}

	if util.IsInDocker() {
		log.Debug("Running inside Docker")
	}

	// Init variables global environment
	err = config.InitGlobalEnvironment()
	if err != nil {
		logging.Panic(err, "Failed to setup configuration variables")
	}

	if conf.Sync.MQTT.EnablePahoLogging {
		logging.InitPahoLogger()
		log.Info("Paho logging enabled")
	}

	var promAgent *prometheus.Agent

	if !build.Light && conf.Prometheus.Enabled {
		promAgent = prometheus.NewAgent(&conf.Prometheus)
	}

	err = db.OpenDB(&conf.DB)
	if err != nil {
		logging.Panic(err, "Failed to open Database")
	}

	log.Info("Database ready")

	defer func() {
		if err := db.Close(); err != nil {
			logging.Error(err, "Failed to close database")
		}
		log.Info("Database closed")
	}()

	var dbusServer *dbus.Server

	if conf.DBus.Enabled {
		dbusServer, err = dbus.NewServer(&conf.DBus)
		if err != nil {
			logging.Panic(err, "Failed to connect to DBus")
		}

		err = dbusServer.Start()
		if err != nil {
			logging.Panic(err, "Failed to start DBus server")
		}

		defer func() {
			if err := dbusServer.Stop(); err != nil {
				logging.Error(err, "Failed to stop DBus server")
			}
		}()
	}

	if !build.Light && conf.HTTP.Enabled {
		server := http.NewServer(&conf.HTTP)
		server.Start()
		log.Info("HTTP server listening on ", conf.HTTP.Address())
		defer func() {
			if err := server.Stop(); err != nil {
				logging.Error(err, "Failed to stop HTTP server")
			}
		}()
	}

	if !build.Light && conf.Sentry.Enabled {
		err = conf.Sentry.InitSentry()
		if err != nil {
			logging.Panic(err, "Failed to init Sentry")
		}
		log.Info("Sentry initialized")
	}

	// Create sync worker
	syncWorker, err := sync.NewWorker(&conf.Sync)
	if err != nil {
		logging.Panic(err, "Failed to create sync worker")
	}

	// Set sync callback
	if dbusServer != nil {
		syncWorker.AddSyncCallback(dbusServer.SignalSyncEvent)
	}

	err = syncWorker.Start()
	if err != nil {
		logging.Panic(err, "Failed to start sync worker")
	}
	defer func() {
		if err := syncWorker.Stop(); err != nil {
			logging.Error(err, "Failed to stop sync worker")
		}
	}()

	if !build.Light && promAgent != nil {
		// Register GORM metrics
		err = db.InitGormMetrics(conf)
		if err != nil {
			logging.Panic(err, "Failed to register GORM database metrics")
		}

		err = promAgent.RegisterMetrics(db.NewMetrics(), syncWorker)
		if err != nil {
			logging.Panic(err, "Failed to register Prometheus metrics")
		}
		log.Debug("Prometheus metrics registered")

		err = promAgent.Start()
		if err != nil {
			logging.Panic(err, "Failed to start Prometheus agent")
		}
		defer func() {
			if err := promAgent.Stop(); err != nil {
				logging.Error(err, "Failed to stop Prometheus agent")
			}
		}()
	}

	if conf.Sync.TelemetryEnabled {
		log.Info("Telemetry enabled")
	} else {
		log.Info("Telemetry disabled")
	}

	if log.IsLevelEnabled(log.DebugLevel) {
		telData, err := telemetry.Get()
		if err != nil {
			logging.Error(err, "Failed to get startup telemetry")
		} else {
			telJSON, err := json.Marshal(telData)
			if err != nil {
				logging.Panic(err, "Failed to marshal startup telemetry to JSON")
			}
			log.Debugf("Startup telemetry: %s", string(telJSON))
		}
	}

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	for range c {
		// Handle Signals
		log.Info(":(")
		return 1
	}

	return 0
}
