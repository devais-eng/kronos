package prometheus

import (
	"context"
	"devais.it/kronos/internal/pkg/config"
	"devais.it/kronos/internal/pkg/logging"
	"devais.it/kronos/internal/pkg/util"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/client_golang/prometheus/push"
	"github.com/rotisserie/eris"
	log "github.com/sirupsen/logrus"
	"net/http"
	"sync"
	"time"
)

var (
	ErrAgentNotConfigured = eris.New("No Prometheus server or pusher configured")
)

type Agent struct {
	conf              *config.PrometheusConfig
	registeredMetrics []Metrics
	metricsMu         sync.RWMutex
	server            *http.Server
	quitCond          util.ChanCond
}

func NewAgent(conf *config.PrometheusConfig) *Agent {
	var server *http.Server

	if conf.StartServer {
		server = &http.Server{
			Addr:    conf.Address(),
			Handler: promhttp.Handler(),
		}
	}

	return &Agent{
		conf:   conf,
		server: server,
	}
}

func (a *Agent) RegisterMetrics(metrics ...Metrics) error {
	a.metricsMu.Lock()
	defer a.metricsMu.Unlock()

	for _, m := range metrics {
		// Make sure metrics are updated
		if err := m.RefreshMetrics(); err != nil {
			return eris.Wrap(err, "failed to refresh metrics")
		}

		// Register collectors to Prometheus
		for _, collector := range m.Collectors() {
			if err := prometheus.Register(collector); err != nil {
				return eris.Wrap(err, "failed to register Prometheus collector")
			}
		}
	}

	a.registeredMetrics = append(a.registeredMetrics, metrics...)

	return nil
}

func (a *Agent) RefreshMetrics() error {
	a.metricsMu.RLock()
	defer a.metricsMu.RUnlock()

	for _, metric := range a.registeredMetrics {
		if err := metric.RefreshMetrics(); err != nil {
			return eris.Wrap(err, "failed to refresh metrics")
		}
	}

	return nil
}

func (a *Agent) Start() error {
	if !a.conf.StartServer && a.conf.PushAddress == "" {
		return ErrAgentNotConfigured
	}

	go a.startRefresher()
	log.Debug("Prometheus refresher started")

	if a.conf.StartServer {
		go a.startServer()
		log.Debug("Prometheus server started")
	}

	if a.conf.PushAddress != "" {
		go a.startPusher()
		log.Debug("Prometheus pusher started")
	}

	log.Info("Prometheus agent started")

	return nil
}

func (a *Agent) Stop() error {
	a.quitCond.Broadcast()

	if a.conf.StartServer {
		if err := a.stopServer(); err != nil {
			return eris.Wrap(err, "failed to stop Prometheus server")
		}
	}

	log.Info("Prometheus agent stopped")

	return nil
}

func (a *Agent) startRefresher() {
	t := time.NewTicker(a.conf.RefreshInterval)

	for {
		select {
		case <-a.quitCond.Wait():
			log.Debug("Prometheus refresher quit")
			return
		case <-t.C:
			if err := a.RefreshMetrics(); err != nil {
				logging.Error(err, "Failed to refresh Prometheus metrics")
			}
		}
	}
}

func (a *Agent) startPusher() {
	conf := a.conf

	pusher := push.New(conf.PushAddress, conf.PushJobName)

	if conf.PushUsername != "" || conf.PushPassword != "" {
		pusher.BasicAuth(conf.PushUsername, conf.PushPassword)
	}

	a.metricsMu.RLock()
	for _, metric := range a.registeredMetrics {
		for _, collector := range metric.Collectors() {
			pusher = pusher.Collector(collector)
		}
	}
	a.metricsMu.RUnlock()

	a.quitCond = util.ChanCond{}

	t := time.NewTicker(conf.PushInterval)

	for {
		select {
		case <-a.quitCond.Wait():
			log.Debug("Prometheus pusher quit")
			return
		case <-t.C:
			if err := a.RefreshMetrics(); err != nil {
				logging.Error(err, "Failed to refresh Prometheus metrics")
			}

			if err := pusher.Push(); err != nil {
				logging.Error(err, "Failed to push metrics to Prometheus pusher")
			} else {
				log.Trace("Prometheus metrics pushed")
			}
		}
	}
}

func (a *Agent) startServer() {
	if err := a.server.ListenAndServe(); err != nil {
		if eris.Is(err, http.ErrServerClosed) {
			log.Info("Prometheus HTTP server closed")
		} else {
			logging.Error(err, "Prometheus HTTP server error")
		}
	}
}

func (a *Agent) stopServer() error {
	ctx, cancel := context.WithTimeout(context.Background(), a.conf.Timeout)
	defer cancel()
	return a.server.Shutdown(ctx)
}
