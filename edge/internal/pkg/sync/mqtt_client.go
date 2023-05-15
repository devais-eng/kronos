package sync

import (
	"devais.it/kronos/internal/pkg/telemetry"
	"time"

	"devais.it/kronos/internal/pkg/config"
	"devais.it/kronos/internal/pkg/logging"
	"devais.it/kronos/internal/pkg/serialization"
	"devais.it/kronos/internal/pkg/services"
	"devais.it/kronos/internal/pkg/sync/messages"
	"devais.it/kronos/internal/pkg/types"
	"devais.it/kronos/internal/pkg/util"
	MQTT "github.com/eclipse/paho.mqtt.golang"
	"github.com/google/uuid"
	"github.com/rotisserie/eris"
	log "github.com/sirupsen/logrus"
)

type MQTTClient struct {
	conf            *config.MQTTConfig
	syncConf        *config.SyncConfig
	client          MQTT.Client
	connectionCb    ConnectionCallback
	disconnectionCb DisconnectionCallback
	syncCb          SyncCallback
	commandCb       CommandCallback
	serializer      serialization.Serializer
	deserializer    serialization.Deserializer

	baseEnv *util.Environment
}

func (c *MQTTClient) deviceID() string {
	return c.baseEnv.Get("deviceID")
}

func NewMQTTClient(syncConf *config.SyncConfig) (*MQTTClient, error) {
	conf := &syncConf.MQTT

	globalEnv, err := config.GetGlobalEnvironment()
	if err != nil {
		return nil, err
	}

	baseEnv := util.NewEnvironment(globalEnv)
	baseEnv.Set("username", conf.Username)

	c := &MQTTClient{
		conf:     conf,
		syncConf: syncConf,
		baseEnv:  baseEnv,
	}

	c.serializer, c.deserializer, err = conf.Serialization.NewSerializer()
	if err != nil {
		return nil, err
	}

	err = c.createPahoMQTTClient()
	if err != nil {
		return nil, eris.Wrap(err, "failed to create the Paho MQTT client")
	}

	if log.IsLevelEnabled(log.DebugLevel) {
		fields := log.Fields{
			"url":                  conf.URL(),
			"clientID":             conf.ClientID,
			"cleanSession":         conf.CleanSession,
			"keepAlive":            conf.KeepAlive,
			"orderMatters":         conf.OrderMatters,
			"communicationTimeout": conf.CommunicationTimeout,
			"storageType":          conf.StorageType,
		}
		log.WithFields(fields).Debug("MQTT client options")
	}

	return c, nil
}

func createMQTTStore(conf *config.MQTTConfig) (MQTT.Store, error) {
	var store MQTT.Store

	switch conf.StorageType {
	case "file":
		store = MQTT.NewFileStore(conf.StoragePath)
	case "badger":
		store = NewBadgerStore(conf.StoragePath)
	case "memory":
		store = MQTT.NewMemoryStore()
	default:
		return nil, eris.Errorf("Unknown MQTT storage type: '%s'", conf.StorageType)
	}

	return store, nil
}

// createPahoMQTTClient creates the underlying Paho client of an MQTTClient.
// It is assumed that the parent client is initialized.
func (c *MQTTClient) createPahoMQTTClient() error {
	conf := c.conf

	clientID := conf.ClientID
	if conf.RandomizeClientID {
		// Append a new UUID to ClientID
		clientID = clientID + "-" + uuid.NewString()
	}

	// Create messages store
	store, err := createMQTTStore(conf)
	if err != nil {
		return err
	}
	log.Infof("Using '%s' store with Paho", conf.StorageType)

	// Load TLS certificates
	tlsConfig, err := conf.TLS.Load()
	if err != nil {
		return err
	}

	// Setup Paho options
	options := MQTT.NewClientOptions().
		AddBroker(conf.URL()).
		SetClientID(clientID).
		SetUsername(conf.Username).
		SetPassword(conf.Password).
		SetTLSConfig(tlsConfig).
		SetCleanSession(conf.CleanSession).
		SetKeepAlive(conf.KeepAlive).
		SetOrderMatters(conf.OrderMatters).
		SetConnectTimeout(conf.CommunicationTimeout).
		SetPingTimeout(conf.CommunicationTimeout).
		SetWriteTimeout(conf.CommunicationTimeout).
		SetAutoReconnect(false).
		SetConnectionLostHandler(func(client MQTT.Client, e error) {
			if c.disconnectionCb != nil {
				c.disconnectionCb(e)
			}
		}).
		SetOnConnectHandler(func(_ MQTT.Client) {
			if c.connectionCb != nil {
				c.connectionCb()
			}
		}).
		SetStore(store)

	if conf.LastWillEnabled {
		willTopic, err := c.baseEnv.EscapeStringVariables(conf.DisconnectedTopic)
		if err != nil {
			return eris.Wrap(err, "failed to build MQTT will topic")
		}
		willMsg := &messages.Disconnected{
			DeviceID:  c.deviceID(),
			Timestamp: nil,
		}
		msgJson, err := c.serializer.Serialize(willMsg)
		if err != nil {
			return eris.Wrap(err, "failed to marshal MQTT will message to JSON")
		}
		options = options.SetWill(
			willTopic,
			string(msgJson),
			conf.PubQoS,
			conf.PubRetained,
		)
	}

	c.client = MQTT.NewClient(options)

	log.Debug("Paho MQTT client created")

	return nil
}

//=============================================================================
// Client interface implementation
//=============================================================================

func (c *MQTTClient) Connect() error {
	err := c.waitToken(c.client.Connect())
	if err != nil {
		return err
	}

	// TODO: Move this up to sync worker
	ts := util.TimestampMs()
	connectedMsg := &messages.Connected{
		DeviceID:  c.deviceID(),
		Timestamp: &ts,
	}

	if c.syncConf.TelemetryEnabled {
		telData, err := telemetry.Get()
		if err != nil {
			return eris.Wrap(err, "failed to get telemetry data to publish")
		}
		connectedMsg.Telemetry = telData
	}

	return c.publishConnected(connectedMsg)
}

func (c *MQTTClient) Disconnect() error {
	// TODO: move this up to sync worker
	if c.syncConf.NotifyGracefulDisconnect {
		// Try to send disconnect message
		ts := util.TimestampMs()
		msg := &messages.Disconnected{
			DeviceID:  c.deviceID(),
			Timestamp: &ts,
		}
		err := c.publishDisconnect(msg)
		if err != nil {
			logging.Error(err, "Failed to publish disconnection message")
		} else {
			log.Info("Disconnection message published")
		}
	}

	c.client.Disconnect(uint(c.conf.CommunicationTimeout.Milliseconds()))

	return nil
}

func (c *MQTTClient) SetSyncCallback(cb SyncCallback) {
	c.syncCb = cb
}

func (c *MQTTClient) SetCommandCallback(cb CommandCallback) {
	c.commandCb = cb
}

func (c *MQTTClient) SetConnectionCallback(cb ConnectionCallback) {
	c.connectionCb = cb
}

func (c *MQTTClient) SetDisconnectionCallback(cb DisconnectionCallback) {
	c.disconnectionCb = cb
}

func (c *MQTTClient) Subscribe() error {
	syncTopicGlobal, err := c.baseEnv.EscapeStringVariables(c.conf.SyncTopicGlobal)
	if err != nil {
		return err
	}

	syncTopicSpecific, err := c.baseEnv.EscapeStringVariables(c.conf.SyncTopicSpecific)
	if err != nil {
		return err
	}

	commandsTopic, err := c.baseEnv.EscapeStringVariables(c.conf.CommandsTopic)
	if err != nil {
		return err
	}

	// Create a list to subscribe in parallel and wait for
	// all tokens at once
	tokens := make([]MQTT.Token, 0, 3)

	syncMessagesFn := func(_ MQTT.Client, message MQTT.Message) {
		if c.syncCb != nil {
			var syncMessage messages.Sync
			err := c.deserializer.Deserialize(message.Payload(), &syncMessage)
			if err != nil {
				logging.Error(err, "Failed to deserialize sync message")
			} else if syncMessage == nil {
				log.Error("Received empty sync message")
			} else {
				c.syncCb(syncMessage)
			}
		}
	}

	token := c.client.Subscribe(syncTopicGlobal, c.conf.SubQoS, syncMessagesFn)
	tokens = append(tokens, token)

	token = c.client.Subscribe(syncTopicSpecific, c.conf.SubQoS, syncMessagesFn)
	tokens = append(tokens, token)

	token = c.client.Subscribe(commandsTopic, c.conf.SubQoS, func(_ MQTT.Client, message MQTT.Message) {
		if c.commandCb != nil {
			var commandMessage messages.ServerCommand
			err := c.deserializer.Deserialize(message.Payload(), &commandMessage)
			if err != nil {
				logging.Error(err, "Failed to deserialize server command message")
			} else {
				c.commandCb(&commandMessage)
			}
		}
	})

	tokens = append(tokens, token)

	for _, t := range tokens {
		if err := c.waitToken(t); err != nil {
			return mapErr(err)
		}
	}

	return nil
}

func (c *MQTTClient) PublishVersions() error {
	connectedTopic, err := c.baseEnv.EscapeStringVariables(c.conf.ConnectedTopic)
	if err != nil {
		return err
	}

	itemsCount, err := services.GetItemsCount()
	if err != nil {
		return eris.Wrap(err, "failed to get items count")
	}
	attributesCount, err := services.GetAttributesCount()
	if err != nil {
		return eris.Wrap(err, "failed to get attributes count")
	}

	pageSize := c.conf.MaxEntitiesPerMessage

	itemsPages := util.CeilDiv(int(itemsCount), pageSize)
	attributesPages := util.CeilDiv(int(attributesCount), pageSize)

	maxPages := util.MaxInt(itemsPages, attributesPages)

	tokens := make([]MQTT.Token, 0, maxPages)

	for page := 1; page <= maxPages; page++ {
		versions := map[types.EntityType]messages.EntityVersions{}

		if page <= itemsPages {
			itemsVersion, err := services.GetItemsVersion(page, pageSize)
			if err != nil {
				return eris.Wrap(err, "failed to get items version")
			}
			versions[types.EntityTypeItem] = itemsVersion
		}

		if page <= attributesPages {
			attributesVersion, err := services.GetAttributesVersion(page, pageSize)
			if err != nil {
				return eris.Wrap(err, "failed to get attributes version")
			}
			versions[types.EntityTypeAttribute] = attributesVersion
		}

		msg := &messages.Versions{
			Timestamp: util.TimestampMs(),
			Versions:  versions,
		}

		msgBytes, err := c.serializer.Serialize(msg)
		if err != nil {
			return eris.Wrap(err, "failed to serialize connection message")
		}

		token := c.client.Publish(connectedTopic, c.conf.PubQoS, c.conf.PubRetained, msgBytes)
		tokens = append(tokens, token)
	}

	// Wait all tokens
	for _, token := range tokens {
		if err := c.waitToken(token); err != nil {
			err = mapErr(err)
			return eris.Wrap(err, "failed to publish versions message")
		}
	}

	return nil
}

func (c *MQTTClient) PublishEvents(events []messages.Event) error {
	topic, err := c.buildEventTopic()
	if err != nil {
		return eris.Wrap(err, "failed to build events topic")
	}

	err = c.publish(topic, events)
	return mapErr(err)
}

func (c *MQTTClient) PublishCommandResponse(message *messages.CommandResponse) error {
	topic, err := c.buildCommandResponseTopic(message)
	if err != nil {
		return eris.Wrap(err, "failed to build command response topic")
	}

	return c.retryPublish(topic, message)
}

func (c *MQTTClient) publishConnected(message *messages.Connected) error {
	topic, err := c.baseEnv.EscapeStringVariables(c.conf.ConnectedTopic)
	if err != nil {
		return eris.Wrap(err, "failed to build connected topic")
	}
	err = c.publish(topic, message)
	if err != nil {
		return eris.Wrap(err, "failed to publish connected message")
	}
	return nil
}

func (c *MQTTClient) publishDisconnect(message *messages.Disconnected) error {
	topic, err := c.baseEnv.EscapeStringVariables(c.conf.DisconnectedTopic)
	if err != nil {
		return eris.Wrap(err, "failed to build disconnect topic")
	}
	err = c.publish(topic, message)
	if err != nil {
		return eris.Wrap(err, "failed to publish disconnect message")
	}
	return nil
}

//=============================================================================
// Utilities
//=============================================================================

// waitToken waits a Paho token, returning ErrTimeout when a timeout occurs
func (c *MQTTClient) waitToken(token MQTT.Token) error {
	if !token.WaitTimeout(c.conf.CommunicationTimeout) {
		return ErrTimeout
	}
	return token.Error()
}

func (c *MQTTClient) publish(topic string, message interface{}) error {
	payload, err := c.serializer.Serialize(message)
	if err != nil {
		return eris.Wrap(err, "failed to serialize message")
	}
	token := c.client.Publish(topic, c.conf.PubQoS, c.conf.PubRetained, payload)
	return c.waitToken(token)
}

// retryPublish publishes a message, retrying on failure.
// Message publishing will be retried until configured max retries or
// max backoff time is reached.
func (c *MQTTClient) retryPublish(topic string, message interface{}) error {
	backOff := c.syncConf.Backoff.NewBackoff()
	interval := backOff.InitialInterval
	retry := 0
	for {
		err := c.publish(topic, message)
		if err == nil {
			return nil
		}

		if c.conf.MaxRetries > 0 && retry >= c.conf.MaxRetries {
			return err
		}

		if interval == backOff.Stop {
			return err
		}

		time.Sleep(interval)

		interval = backOff.NextBackOff()
		retry++
	}
}

func (c *MQTTClient) buildEventTopic() (string, error) {
	env := util.NewEnvironment(c.baseEnv)

	return env.EscapeStringVariables(c.conf.EventsTopic)
}

func (c *MQTTClient) buildCommandResponseTopic(message *messages.CommandResponse) (string, error) {
	env := util.NewEnvironment(c.baseEnv)
	env.Set("uuid", message.UUID)

	return env.EscapeStringVariables(c.conf.CommandsResponseTopic)
}

// mapErr converts from Paho MQTT errors to worker errors
func mapErr(err error) error {
	if eris.Is(err, MQTT.ErrNotConnected) {
		return ErrNotConnected
	}
	return err
}
