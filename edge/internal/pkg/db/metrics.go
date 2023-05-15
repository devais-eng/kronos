package db

import (
	"devais.it/kronos/internal/pkg/config"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/rotisserie/eris"
	log "github.com/sirupsen/logrus"
	gormProm "gorm.io/plugin/prometheus"
)

type Metrics struct {
	dbSize prometheus.Gauge
}

func NewMetrics() *Metrics {
	return &Metrics{
		dbSize: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Name: "kronos_db_size",
				Help: "The SQLite database size",
			},
		),
	}
}

func (m *Metrics) Collectors() []prometheus.Collector {
	return []prometheus.Collector{
		m.dbSize,
	}
}

func (m *Metrics) RefreshMetrics() error {
	size, err := Size()
	if err != nil {
		return err
	}
	m.dbSize.Set(float64(size))

	return nil
}

// InitGormMetrics initializes the Prometheus service of Gorm
func InitGormMetrics(conf *config.Config) error {
	p := gormProm.New(gormProm.Config{
		DBName:          conf.DB.URL,
		RefreshInterval: 0,
		PushAddr:        conf.Prometheus.PushAddress,
		PushUser:        conf.Prometheus.PushUsername,
		PushPassword:    conf.Prometheus.PushPassword,
		// Never start Prometheus HTTP server on Gorm side.
		// We'll initialize it later in our main
		StartServer:    false,
		HTTPServerPort: 0,
	})

	if err := p.Initialize(db); err != nil {
		return eris.Wrap(err, "failed to init Gorm Prometheus service")
	}

	log.Info("GORM Prometheus metrics initialized")

	return nil
}
