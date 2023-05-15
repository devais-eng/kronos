package config

import (
	"fmt"
	"time"
)

const (
	defaultHTTPHost    = "localhost"
	defaultHTTPPort    = 5000
	defaultHTTPTimeout = 5 * time.Second
)

type HTTPSentryConfig struct {
	// Enabled determines if HTTP server errors should be forwarded to Sentry.
	Enabled bool

	// WaitForDelivery determines if the error handler should wait until
	// errors are delivered to Sentry.
	WaitForDelivery bool

	// DeliveryTimeout is the
	DeliveryTimeout time.Duration
}

func DefaultHTTPSentryConfig() HTTPSentryConfig {
	return HTTPSentryConfig{
		Enabled:         true,
		WaitForDelivery: false,
		DeliveryTimeout: time.Duration(0),
	}
}

type HTTPConfig struct {
	// Enabled determines if the HTTP server should be run
	Enabled bool

	// DebugMode enables debug mode on the Gin HTTP server
	DebugMode bool

	// PprofEnabled determines if pprof should be exposed on the HTTP server
	PprofEnabled bool

	// Sentry is the Sentry configuration for this component
	Sentry HTTPSentryConfig

	// ReplyCreatedData determines if create APIs should reply with the whole
	// created data.
	// If false, only the record ID is replied instead.
	ReplyCreatedData bool

	// Host is the host where the HTTP server will be bound to
	Host string

	// Port is the port where the HTTP server will listen for connections
	Port int

	// Timeout is the server startup/shutdown timeout
	Timeout time.Duration
}

func (c *HTTPConfig) Address() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}

func DefaultHTTPConfig() HTTPConfig {
	return HTTPConfig{
		Enabled:          false,
		DebugMode:        false,
		PprofEnabled:     false,
		Sentry:           DefaultHTTPSentryConfig(),
		ReplyCreatedData: true,
		Host:             defaultHTTPHost,
		Port:             defaultHTTPPort,
		Timeout:          defaultHTTPTimeout,
	}
}
