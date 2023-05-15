package services

import (
	"devais.it/kronos/internal/pkg/db"
	"devais.it/kronos/internal/pkg/db/models"
	"devais.it/kronos/internal/pkg/types"
	"github.com/google/uuid"
	"github.com/rotisserie/eris"
	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"
	"testing"
)

const (
	mbEvents        = "EVENTS_TEST"
	eventsBatchSize = 50
)

type EventsSuite struct {
	db.SuiteBase
}

func (s *EventsSuite) assertCount(expected int) {
	assert := s.Require()
	count, err := GetEventsCount()
	assert.NoError(err)
	assert.Equal(int64(expected), count)
}

func (s *EventsSuite) TestGet() {
	assert := s.Require()

	s.assertCount(0)

	// Test not found errors
	event, err := GetLastEvent()
	assert.ErrorIs(err, gorm.ErrRecordNotFound)
	assert.Nil(event)

	event, err = GetFirstEvent()
	assert.ErrorIs(err, gorm.ErrRecordNotFound)
	assert.Nil(event)

	events, err := GetLastEvents(db.DB(), eventsBatchSize)
	assert.NoError(err)
	assert.Empty(events)

	events, err = GetFirstEvents(db.DB(), eventsBatchSize)
	assert.NoError(err)
	assert.Empty(events)

	event, err = GetEvent(types.EventEntityCreated, types.EntityTypeItem, uuid.NewString())
	assert.ErrorIs(err, gorm.ErrRecordNotFound)
	assert.Nil(event)

	called := false
	err = TryDequeueEvent(db.DB(), func(event *models.Event) error {
		called = true
		return nil
	})
	assert.False(called)
	assert.ErrorIs(err, gorm.ErrRecordNotFound)

	count, err := TryDequeueEvents(db.DB(), eventsBatchSize, func(events []models.Event) error {
		called = true
		return nil
	})
	assert.False(called)
	assert.ErrorIs(err, gorm.ErrRecordNotFound)
	assert.Zero(count)

	// Create events
	firstItem := &models.Item{}
	lastItem := &models.Item{}

	for i := 0; i < eventsBatchSize; i++ {
		item := newItem()
		if i == 0 {
			firstItem = item
		} else if i == eventsBatchSize-1 {
			lastItem = item
		}
		assert.NoError(CreateItem(item, mbEvents))
	}

	s.assertCount(eventsBatchSize)

	// Check events content
	firstEvent, err := GetFirstEvent()
	assert.NoError(err)

	assert.Equal(types.EventEntityCreated, firstEvent.EventType)
	assert.Equal(types.EntityTypeItem, firstEvent.EntityType)
	assert.Equal(firstItem.ID, firstEvent.EntityID)
	assert.Equal(mbEvents, firstEvent.TriggeredBy)
	assert.Empty(firstEvent.TxUUID)
	assert.Zero(firstEvent.TxLen)

	lastEvent, err := GetLastEvent()
	assert.NoError(err)

	assert.Equal(types.EventEntityCreated, lastEvent.EventType)
	assert.Equal(types.EntityTypeItem, lastEvent.EntityType)
	assert.Equal(lastItem.ID, lastEvent.EntityID)
	assert.Equal(mbEvents, lastEvent.TriggeredBy)
	assert.Empty(lastEvent.TxUUID)
	assert.Zero(lastEvent.TxLen)

	err = TryDequeueEvent(db.DB(), func(event *models.Event) error {
		return eris.New("test err")
	})
	assert.Error(err)

	count, err = TryDequeueEvents(db.DB(), eventsBatchSize, func(events []models.Event) error {
		return eris.New("test err")
	})
	assert.Error(err)
	assert.Zero(count)

	err = TryDequeueEvent(db.DB(), func(event *models.Event) error {
		assert.Equal(types.EventEntityCreated, event.EventType)
		assert.Equal(types.EntityTypeItem, event.EntityType)
		assert.Equal(firstItem.ID, event.EntityID)

		return nil
	})
	assert.NoError(err)

	s.assertCount(eventsBatchSize - 1)

	count, err = TryDequeueEvents(db.DB(), eventsBatchSize, func(events []models.Event) error {
		assert.Len(events, eventsBatchSize-1)
		return nil
	})

	assert.NoError(err)
	assert.Equal(count, eventsBatchSize-1)

	s.assertCount(0)

	assert.NoError(CreateItem(newItem(), mbEvents))

	firstEvent, err = GetFirstEvent()
	assert.NoError(err)
	assert.Equal(firstEvent.ID, uint(1))

	items := make([]models.Item, eventsBatchSize)
	for i := 0; i < eventsBatchSize; i++ {
		items[i] = *newItem()
	}

	assert.NoError(BatchCreateItems(items, mbEvents))

	err = TryDequeueEvent(db.DB(), func(event *models.Event) error {
		return nil
	})
	assert.NoError(err)

	assert.NoError(CreateItem(newItem(), mbEvents))

	firstEvent, err = GetFirstEvent()
	assert.NoError(err)
	assert.Equal(uint(2), firstEvent.ID)
	assert.Equal(0, firstEvent.TxIndex)
	assert.Equal(eventsBatchSize, firstEvent.TxLen)

	lastEvent, err = GetLastEvent()
	assert.NoError(err)
	assert.Equal(uint(eventsBatchSize+2), lastEvent.ID)
	assert.Equal(0, lastEvent.TxIndex)
	assert.Equal(0, lastEvent.TxLen)
}

func (s *EventsSuite) TestCreate() {
	assert := s.Require()

	s.assertCount(0)

	assert.Error(UpdateEvent(db.DB(), nil))
	assert.Error(DeleteEvent(db.DB(), &models.Event{}))

	err := DeleteEventsByType(
		db.DB(),
		types.EntityTypeItem,
		uuid.NewString(),
		mbEvents,
		[]types.EventType{types.EventEntityCreated},
	)
	assert.ErrorIs(err, gorm.ErrRecordNotFound)

	err = PublishEvent(
		&db.TxContext{Tx: db.DB()},
		types.EventEntityCreated,
		types.EntityTypeItem,
		uuid.NewString(),
		mbEvents,
		make(chan int),
	)
	assert.Error(err)
	s.assertCount(0)

	items := make([]models.Item, eventsBatchSize)
	firstItem := &models.Item{}
	lastItem := &models.Item{}

	for i := 0; i < eventsBatchSize; i++ {
		item := newItem()
		if i == 0 {
			firstItem = item
		} else if i == eventsBatchSize-1 {
			lastItem = item
		}
		items[i] = *item
	}

	assert.NoError(BatchCreateItems(items, mbEvents))

	s.assertCount(eventsBatchSize)

	var eventBody map[string]interface{}

	firstEvent, err := GetFirstEvent()
	assert.NoError(err)
	assert.NotEmpty(firstEvent.TxUUID)
	assert.NotZero(firstEvent.TxLen)
	assert.NoError(firstEvent.UnmarshalBody(&eventBody))
	assert.Equal(firstItem.ID, firstEvent.EntityID)
	assert.Equal(firstItem.ID, eventBody["id"])
	assert.Equal(firstItem.Name, eventBody["name"])
	assert.Equal(firstItem.Type, eventBody["type"])
	assert.NotContains(eventBody, "attributes")

	lastEvent, err := GetLastEvent()
	assert.NoError(err)
	assert.NotEmpty(lastEvent.TxUUID)
	assert.NotZero(lastEvent.TxLen)
	assert.NoError(lastEvent.UnmarshalBody(&eventBody))
	assert.Equal(lastItem.ID, lastEvent.EntityID)
	assert.Equal(lastItem.ID, eventBody["id"])
	assert.Equal(lastItem.Name, eventBody["name"])
	assert.Equal(lastItem.Type, eventBody["type"])

	assert.Equal(firstEvent.TxLen, lastEvent.TxLen)
	assert.Equal(firstEvent.TxUUID, lastEvent.TxUUID)

	// Test update
	newItemName := "NewFirstItemName"

	assert.NoError(UpdateItem(map[string]interface{}{
		"id":   firstItem.ID,
		"name": newItemName,
	}, mbEvents))

	s.assertCount(eventsBatchSize)

	firstEvent, err = GetFirstEvent()
	assert.NoError(err)
	assert.Equal(firstItem.ID, firstEvent.EntityID)
	assert.NoError(firstEvent.UnmarshalBody(&eventBody))
	assert.Equal(newItemName, eventBody["name"])

	assert.NoError(DeleteItem(firstItem, mbEvents))
	s.assertCount(eventsBatchSize - 1)

	for i := 1; i < eventsBatchSize; i++ {
		assert.NoError(DeleteItem(&items[i], mbEvents))
	}

	s.assertCount(0)

	// Push delete with no previous related event
	err = PublishEvent(
		&db.TxContext{Tx: db.DB()},
		types.EventEntityDeleted,
		types.EntityTypeItem,
		uuid.NewString(),
		mbEvents,
		firstItem,
	)
	assert.NoError(err)
	s.assertCount(1)

	// Create and update events with different 'modifiedBy' fields shouldn't merge
	assert.NoError(CreateItem(firstItem, mbEvents))
	assert.NoError(UpdateItem(map[string]interface{}{"id": firstItem.ID, "name": newItemName}, "TEST_OTHER"))
	s.assertCount(3)
}

func TestEventsService(t *testing.T) {
	suite.Run(t, new(EventsSuite))
}
