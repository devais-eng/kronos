package sync

import (
	"devais.it/kronos/internal/pkg/sync/messages"
	"sync"
	"testing"
	"time"

	"devais.it/kronos/internal/pkg/config"
	"devais.it/kronos/internal/pkg/constants"
	"devais.it/kronos/internal/pkg/db"
	"devais.it/kronos/internal/pkg/db/models"
	"devais.it/kronos/internal/pkg/services"
	"devais.it/kronos/internal/pkg/types"
	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"
)

const (
	timeout     = 30 * time.Second
	tick        = 200 * time.Millisecond
	stopTimeout = 1 * time.Second

	modifiedByTest = "TEST"
)

func testSyncConfig() *config.SyncConfig {
	conf := config.DefaultSyncConfig()
	conf.Backoff.InitialInterval = tick
	conf.Backoff.MaxInterval = timeout
	conf.StopTimeout = stopTimeout
	conf.PublishVersions = true

	conf.ClientType = "MQTT"
	conf.MQTT.CommunicationTimeout = timeout
	conf.MQTT.MaxRetries = 1

	return &conf
}

type testClient struct {
	sync.Mutex
	connected         bool
	subscribed        bool
	versionsPublished bool
	connectionCb      ConnectionCallback
	disconnectionCb   DisconnectionCallback
	syncCb            SyncCallback
	commandCb         CommandCallback
	events            []messages.Event
	commandResponses  []messages.CommandResponse
}

func (c *testClient) Connect() error {
	c.Lock()
	defer c.Unlock()
	c.connected = true
	if c.connectionCb != nil {
		c.connectionCb()
	}
	return nil
}

func (c *testClient) Disconnect() error {
	c.Lock()
	defer c.Unlock()
	c.connected = false
	c.subscribed = false
	c.versionsPublished = false
	if c.disconnectionCb != nil {
		c.disconnectionCb(nil)
	}
	return nil
}

func (c *testClient) SetConnectionCallback(cb ConnectionCallback) {
	c.Lock()
	defer c.Unlock()
	c.connectionCb = cb
}

func (c *testClient) SetDisconnectionCallback(cb DisconnectionCallback) {
	c.Lock()
	defer c.Unlock()
	c.disconnectionCb = cb
}

func (c *testClient) SetSyncCallback(cb SyncCallback) {
	c.Lock()
	defer c.Unlock()
	c.syncCb = cb
}

func (c *testClient) SetCommandCallback(cb CommandCallback) {
	c.Lock()
	defer c.Unlock()
	c.commandCb = cb
}

func (c *testClient) Subscribe() error {
	c.Lock()
	defer c.Unlock()
	c.subscribed = true
	return nil
}

func (c *testClient) PublishVersions() error {
	c.Lock()
	defer c.Unlock()
	c.versionsPublished = true
	return nil
}

func (c *testClient) PublishEvents(events []messages.Event) error {
	c.Lock()
	defer c.Unlock()
	c.events = append(c.events, events...)

	return nil
}

func (c *testClient) PublishCommandResponse(message *messages.CommandResponse) error {
	c.Lock()
	defer c.Unlock()
	c.commandResponses = append(c.commandResponses, *message)
	return nil
}

type WorkerTestSuite struct {
	db.SuiteBase
}

func (s *WorkerTestSuite) TestEvents() {
	assert := s.Require()

	//=========================================================================
	// Setup
	//=========================================================================

	conf := testSyncConfig()

	worker, err := NewWorker(conf)
	assert.NoError(err)

	client := &testClient{}
	worker.client = client

	//=========================================================================
	// Start
	//=========================================================================

	err = worker.Start()
	assert.NoError(err)

	assert.Eventually(func() bool {
		return worker.fsm.Current() == stateDequeueing
	}, timeout, tick)

	assert.True(client.connected)
	assert.True(client.subscribed)
	assert.True(client.versionsPublished)

	//=========================================================================
	// Test sync
	//=========================================================================

	// Send sync message
	itemID := "FakeItem00-ID"
	itemName := "FakeItem00"
	itemType := "FakeItem"

	itemEntry := messages.SyncEntry{
		EntityType: types.EntityTypeItem,
		EntityID:   itemID,
		Action:     messages.SyncActionCreate,
		Payload: map[string]interface{}{
			"id":   itemID,
			"name": itemName,
			"type": itemType,
		},
	}

	fakeSync := messages.Sync{
		itemEntry,
	}

	client.syncCb(fakeSync)

	assert.Eventually(func() bool {
		client.Lock()
		defer client.Unlock()
		return len(client.events) == 1
	}, timeout, tick)

	//=========================================================================
	// Test event
	//=========================================================================

	func() {
		client.Lock()
		defer client.Unlock()

		firstEvent := client.events[0]

		assert.Equal(types.EventEntityCreated, firstEvent.TxType)
		assert.Equal(types.EntityTypeItem, firstEvent.EntityType)
		assert.Equal(constants.ModifiedBySyncName, firstEvent.TriggeredBy)
		assert.Equal(itemID, firstEvent.EntityID)
		assert.Equal(itemName, firstEvent.Body["name"])
		assert.Equal(itemType, firstEvent.Body["type"])
	}()

	createdItem, err := services.GetItemByID(itemID)
	assert.NoError(err)
	assert.Equal(itemID, createdItem.ID)
	assert.Equal(itemName, createdItem.Name)
	assert.Equal(itemType, createdItem.Type)

	//=========================================================================
	// Stop
	//=========================================================================

	err = worker.Stop()
	assert.NoError(err)
}

func (s *WorkerTestSuite) TestCommands() {
	assert := s.Require()

	//=========================================================================
	// Setup
	//=========================================================================

	conf := testSyncConfig()
	conf.PublishVersions = false

	worker, err := NewWorker(conf)
	assert.NoError(err)

	client := &testClient{}
	worker.client = client

	//=========================================================================
	// Start
	//=========================================================================

	err = worker.Start()
	assert.NoError(err)

	assert.Eventually(func() bool {
		return worker.fsm.Current() == stateDequeueing
	}, timeout, tick)

	assert.True(client.connected)
	assert.True(client.subscribed)
	assert.False(client.versionsPublished)

	//=========================================================================
	// Create item
	//=========================================================================

	fakeItem := &models.Item{
		ID:   "FakeItem00-ID",
		Name: "FakeItem00",
		Type: "FakeItem",
	}

	err = services.CreateItem(fakeItem, modifiedByTest)
	assert.NoError(err)

	//=========================================================================
	// Get item with a command
	//=========================================================================

	fakeCommand := &messages.ServerCommand{
		UUID:        uuid.New().String(),
		CommandType: messages.CommandGetEntity,
		EntityType:  types.EntityTypeItem,
		EntityID:    fakeItem.ID,
		Body:        nil,
	}

	client.commandCb(fakeCommand)

	assert.Eventually(func() bool {
		return len(client.commandResponses) == 1
	}, timeout, tick)

	response := client.commandResponses[0]

	assert.True(response.Success)
	assert.Empty(response.Error)
	assert.Equal(fakeCommand.UUID, response.UUID)
	assert.Equal(fakeItem.ID, response.Body["id"])
	assert.Equal(fakeItem.Name, response.Body["name"])
	assert.Equal(fakeItem.Type, response.Body["type"])

	//=========================================================================
	// Test command error
	//=========================================================================

	fakeCommand.CommandType = "INVALID"
	client.commandCb(fakeCommand)

	assert.Eventually(func() bool {
		return len(client.commandResponses) == 2
	}, timeout, tick)

	response = client.commandResponses[1]
	assert.False(response.Success)
	assert.NotEmpty(response.Error)

	itemVersion, err := services.GetItemVersion(fakeItem.ID)
	assert.NoError(err)

	//=========================================================================
	// Test get version command
	//=========================================================================

	fakeCommand.CommandType = messages.CommandGetVersion

	client.commandCb(fakeCommand)

	assert.Eventually(func() bool {
		return len(client.commandResponses) == 3
	}, timeout, tick)

	response = client.commandResponses[2]
	assert.True(response.Success)
	assert.Empty(response.Error)
	assert.Equal(itemVersion, response.Body["version"])

	//=========================================================================
	// Stop
	//=========================================================================

	err = worker.Stop()
	assert.NoError(err)
}

func TestWorker(t *testing.T) {
	suite.Run(t, new(WorkerTestSuite))
}
