package config

import (
	"fmt"
	"time"
)

const (
	defaultPrometheusPushJobName     = "KronosPusher"
	defaultPrometheusHost            = "localhost"
	defaultPrometheusPort            = 2112
	defaultPrometheusTimeout         = 5 * time.Second
	defaultPrometheusPushInterval    = 5 * time.Second
	defaultPrometheusRefreshInterval = 5 * time.Second
)

type PrometheusConfig struct {
	// Enabled determines whether Prometheus metrics should be reported or not
	Enabled bool

	// PushAddress is the address of the Prometheus Pushgateway service
	PushAddress string

	// If authentication is enabled on the Prometheus PushGateway
	// PushUsername is service username
	PushUsername string

	// If authentication is enabled on the Prometheus PushGateway
	// PushUsername is service password
	PushPassword string

	// PushJobName is the name of the Pushgateway job displayed by Prometheus
	PushJobName string

	// PushInterval is the interval at which metrics should be pushed to the
	// Prometheus Pushgateway service
	PushInterval time.Duration

	// RefreshInterval is the interval at which metrics' values should be refreshed
	RefreshInterval time.Duration

	// StartServer determines if an HTTP server exposing Prometheus
	// metrics should be started
	StartServer bool

	// Host is the Prometheus HTTP server hostname
	Host string

	// Port is the Prometheus HTTP server port
	Port int

	// Timeout is the HTTP server operations' timeout
	Timeout time.Duration
}

func (c *PrometheusConfig) Address() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}

func DefaultPrometheusConfig() PrometheusConfig {
	return PrometheusConfig{
		Enabled:         false,
		PushJobName:     defaultPrometheusPushJobName,
		PushInterval:    defaultPrometheusPushInterval,
		RefreshInterval: defaultPrometheusRefreshInterval,
		StartServer:     true,
		Host:            defaultPrometheusHost,
		Port:            defaultPrometheusPort,
		Timeout:         defaultPrometheusTimeout,
	}
}
