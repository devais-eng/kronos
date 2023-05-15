package config

import (
	"fmt"
	"time"
)

const (
	defaultMQTTScheme                = "tcp"
	defaultMQTTHost                  = "localhost"
	defaultMQTTPort                  = 1883
	defaultMQTTClientID              = "kronos"
	defaultMQTTQoS                   = 1
	defaultMQTTCleanSession          = true
	defaultMQTTKeepAlive             = time.Minute
	defaultMQTTCommunicationTimeout  = 30 * time.Second
	defaultMQTTLastWillEnabled       = true
	defaultMQTTMaxEntitiesPerMessage = 50
	defaultMQTTStorageType           = "memory"
	defaultMQTTStoragePath           = "./paho-messages"
	defaultMQTTConnectedTopic        = "/kronos/device/{deviceId}/connected"
	defaultMQTTDisconnectedTopic     = "/kronos/device/{deviceId}/disconnected"
	defaultMQTTEventsTopic           = "/kronos/device/{deviceId}/events"
	defaultMQTTSyncTopicGlobal       = "/kronos/sync"
	defaultMQTTSyncTopicSpecific     = "/kronos/device/{deviceId}/sync"
	defaultMQTTCommandsTopic         = "/kronos/device/{deviceId}/commands"
	defaultMQTTCommandsResponseTopic = "/kronos/device/{deviceId}/commands/{uuid}/response"
)

type MQTTConfig struct {
	// EnablePahoLogging determines if the Paho library logging should be enabled
	EnablePahoLogging bool

	// Scheme is the MQTT connection protocol. It should be of "tcp", "ssl" or "ws"
	Scheme string

	// Host is the ip-address (or hostname) of the MQTT broker
	Host string

	// Port is the MQTT broker port
	Port int

	// ClientID is the MQTT client ID.
	ClientID string

	// RandomizeClientID determines if a random UUID should be appended to
	// the ClientID to increase its uniqueness.
	// NOTE: if enabled, at every restart of the application the client ID will change,
	// triggering the creation of a new session thus making the use of CleanSession=false useless.
	RandomizeClientID bool

	// Username is the MQTT broker username
	Username string

	// Password is the MQTT broker password
	Password string

	// TLS are the MQTTS configuration options
	TLS TLSConfig

	// SubQoS is the Quality of Service used for all subscriptions
	SubQoS byte

	// PubQoS is the Quality of Service used for all publishes
	PubQoS byte

	// CleanSession determines if the MQTT broker should drop the session of this client
	// after every connection/disconnection.
	// If set to false, a persistent session will be established.
	// In a persistent session the broker stores all the client subscriptions and all
	// the messages with a QoS >= 1 that the client missed while offline.
	CleanSession bool

	// KeepAlive is the amount of time that the client should wait before sending a PING
	// request to the broker. This will allow the client to know that a connection has
	// not been lost with the server.
	KeepAlive time.Duration

	// CommunicationTimeout is the timeout for network communication
	CommunicationTimeout time.Duration

	// LastWillEnabled determines whether last will should be enabled or not.
	LastWillEnabled bool

	// Serialization is the configuration for messages serialization
	Serialization SerializationConfig

	// MaxRetries is the maximum number of tries for blocking operations (like publishing of commands response).
	// If 0, operations are retried until the backoff maximum interval is reached.
	MaxRetries int

	// PubRetained determines if messages should always be pushed as retained
	PubRetained bool

	// MaxEntitiesPerMessage determines the maximum number of entities in a single messages.
	// Messages with a higher number of entities will be split into multiple messages.
	MaxEntitiesPerMessage int

	// OrderMatters will set the message routing to guarantee order within
	// each QoS level. By default, this value is true. If set to false (recommended),
	// this flag indicates that messages can be delivered asynchronously
	// from the client to the application and possibly arrive out of order.
	// Specifically, the message handler is called in its own go routine.
	// Note that setting this to true does not guarantee in-order delivery
	// (this is subject to broker settings like "max_inflight_messages=1" in mosquitto)
	// and if true then handlers must not block.
	OrderMatters bool

	// StorageType determines the type of storage to be used to save unhandled raw MQTT messages.
	// Supported types are: "memory", "file" and "badger".
	// Both "memory" and "file" will use the default Paho implementations of each, while "badger"
	// will use the Badger k/v storage library.
	// Badger is a lot more reliable, but its memory usage baseline is higher, thus may not
	// be the ideal choice for constrained environments.
	StorageType string

	// StoragePath is the directory where MQTT messages should be stored.
	// This setting is used by "file" and "badger" storage types.
	StoragePath string

	// ConnectedTopic is the MQTT topic where connection messages are sent.
	// Supports variables
	ConnectedTopic string

	// DisconnectedTopic is the MQTT topic where disconnection messages are sent.
	// This is implemented with MQTT's last will.
	// Supports variables
	DisconnectedTopic string

	// EventsTopic is the MQTT topic where event messages are sent.
	// Supports variables
	EventsTopic string

	// SyncTopicGlobal is a MQTT topic where synchronization messages are received from the server.
	// On this topic global synchronization messages are received.
	// Supports variables.
	SyncTopicGlobal string

	// SyncTopicSpecific is a MQTT topic where synchronization messages are received from the server.
	// On this topic only synchronization messages directed to this device are received.
	// Supports variables.
	SyncTopicSpecific string

	// CommandsTopic is the MQTT topic where command messages are received from the server.
	// Supports variables.
	CommandsTopic string

	// CommandsResponseTopic is the MQTT topic where command response messages are sent.
	// Supports variables.
	CommandsResponseTopic string
}

func (c *MQTTConfig) URL() string {
	return fmt.Sprintf("%s://%s:%d", c.Scheme, c.Host, c.Port)
}

// DefaultMQTTConfig creates a new MQTT configuration structure
// filled with default options
func DefaultMQTTConfig() MQTTConfig {
	return MQTTConfig{
		EnablePahoLogging:     false,
		Scheme:                defaultMQTTScheme,
		Host:                  defaultMQTTHost,
		Port:                  defaultMQTTPort,
		ClientID:              defaultMQTTClientID,
		RandomizeClientID:     false,
		Username:              "",
		Password:              "",
		TLS:                   DefaultTLSConfig(),
		SubQoS:                defaultMQTTQoS,
		PubQoS:                defaultMQTTQoS,
		KeepAlive:             defaultMQTTKeepAlive,
		CleanSession:          defaultMQTTCleanSession,
		CommunicationTimeout:  defaultMQTTCommunicationTimeout,
		LastWillEnabled:       defaultMQTTLastWillEnabled,
		Serialization:         DefaultSerializationConfig(),
		MaxRetries:            0,
		PubRetained:           false,
		MaxEntitiesPerMessage: defaultMQTTMaxEntitiesPerMessage,
		OrderMatters:          false,
		StorageType:           defaultMQTTStorageType,
		StoragePath:           defaultMQTTStoragePath,
		ConnectedTopic:        defaultMQTTConnectedTopic,
		DisconnectedTopic:     defaultMQTTDisconnectedTopic,
		EventsTopic:           defaultMQTTEventsTopic,
		SyncTopicGlobal:       defaultMQTTSyncTopicGlobal,
		SyncTopicSpecific:     defaultMQTTSyncTopicSpecific,
		CommandsTopic:         defaultMQTTCommandsTopic,
		CommandsResponseTopic: defaultMQTTCommandsResponseTopic,
	}
}
