package config

import (
	"github.com/getsentry/sentry-go"
	"github.com/rotisserie/eris"
	log "github.com/sirupsen/logrus"
)

type SentryConfig struct {
	// Enabled determines whether Sentry integration is enabled or not
	Enabled bool

	// Debug determines if Sentry debug mode should be enabled.
	// When debug mode is enabled, debug information is logged
	// to help understand what Sentry is doing.
	Debug bool

	// Dsn is the Sentry connection string. You can obtain it
	// from your Sentry project page
	Dsn string

	// AttachStacktrace determines if stack traces should be
	// attached to Sentry messages
	AttachStacktrace bool

	// SampleRate is the sample rate for event submission in the range [0.0, 1.0].
	// By default all events are sent.
	// The sample rate 0.0 is treated the same as 1.0
	SampleRate float64

	// TracesSampleRate is the sample rate for traces
	TracesSampleRate float64

	// MaxBreadcrumbs is the maximum number of breadcrumbs
	MaxBreadCrumbs int
}

func DefaultSentryConfig() SentryConfig {
	return SentryConfig{
		Enabled:          false,
		Debug:            false,
		Dsn:              "",
		AttachStacktrace: false,
		SampleRate:       0.0,
		TracesSampleRate: 0.0,
		MaxBreadCrumbs:   0,
	}
}

type sentryDebugWriter struct{}

func (w *sentryDebugWriter) Write(p []byte) (n int, err error) {
	str := string(p)
	log.Debug("[sentry] " + str)
	return len(p), nil
}

func (c *SentryConfig) ToSentryClientOptions() sentry.ClientOptions {
	options := sentry.ClientOptions{
		Dsn:              c.Dsn,
		Debug:            c.Debug,
		AttachStacktrace: c.AttachStacktrace,
		SampleRate:       c.SampleRate,
		TracesSampleRate: c.TracesSampleRate,
		MaxBreadcrumbs:   c.MaxBreadCrumbs,
	}

	if c.Debug {
		options.DebugWriter = &sentryDebugWriter{}
	}

	return options
}

func (c *SentryConfig) InitSentry() error {
	clientOptions := c.ToSentryClientOptions()
	if err := sentry.Init(clientOptions); err != nil {
		return eris.Wrap(err, "failed to init Sentry")
	}
	return nil
}
