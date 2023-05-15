package config

import "time"

const (
	defaultSyncTelemetryEnabled         = true
	defaultSyncNotifyGracefulDisconnect = true
	defaultSyncMaxEvents                = 100
	defaultSyncStopTimeout              = 10 * time.Second
)

type SyncClientType string

var (
	SyncClientMQTT SyncClientType = "MQTT"
)

type SyncConfig struct {
	// ClientType is the type of synchronization client to use
	ClientType SyncClientType

	// PublishVersions if set to true, version of all entities will
	// be sent after the first connection, in order to perform a full
	// re-synchronization.
	// This is not recommended for most cases
	PublishVersions bool

	// TelemetryEnabled if set to true, telemetry data will be reported
	// after a successful connection.
	TelemetryEnabled bool

	// NotifyGracefulDisconnect determines if a disconnect message should
	// be sent before a graceful disconnection.
	NotifyGracefulDisconnect bool

	// MaxEvents determines the maximum number of events to send in a single message
	MaxEvents int

	// StopTimeout is the maximum time to wait for synchronization worker stop
	StopTimeout time.Duration

	// MinSleepTime sets the minimum amount of time between
	// sync worker iterations.
	// This is useful to throttle outgoing traffic or limit CPU usage.
	MinSleepTime time.Duration

	// Backoff is the backoff configuration for most fallible operations
	Backoff BackoffConfig

	// Sentry is the configuration of the Sentry client.
	// Sentry will be used to report errors happening inside the synchronization
	// worker
	Sentry SyncSentryConfig

	// MQTT is the configuration of the MQTT client.
	// The MQTT client will be used to communicate with the server application
	MQTT MQTTConfig
}

func DefaultSyncConfig() SyncConfig {
	return SyncConfig{
		ClientType:               SyncClientMQTT,
		PublishVersions:          false,
		TelemetryEnabled:         defaultSyncTelemetryEnabled,
		NotifyGracefulDisconnect: defaultSyncNotifyGracefulDisconnect,
		MaxEvents:                defaultSyncMaxEvents,
		StopTimeout:              defaultSyncStopTimeout,
		MinSleepTime:             0,
		Backoff:                  DefaultBackoffConfig(),
		Sentry:                   DefaultSyncSentryConfig(),
		MQTT:                     DefaultMQTTConfig(),
	}
}

type SyncSentryConfig struct {
	// Enabled determines if Sentry should be used to report
	// errors happening inside the synchronization worker.
	// Sentry should be enabled and configured globally to make
	// this useful
	Enabled bool

	// WaitForDelivery determines if the Sentry client should wait
	// for the delivery of traces
	WaitForDelivery bool

	// DeliveryTimeout is the maximum time to wait for delivery of
	// traces
	DeliveryTimeout time.Duration
}

func DefaultSyncSentryConfig() SyncSentryConfig {
	return SyncSentryConfig{
		Enabled:         false,
		WaitForDelivery: false,
		DeliveryTimeout: time.Duration(0),
	}
}
