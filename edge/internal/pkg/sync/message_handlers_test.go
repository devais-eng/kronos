package sync

import (
	"devais.it/kronos/internal/pkg/constants"
	"devais.it/kronos/internal/pkg/db"
	"devais.it/kronos/internal/pkg/db/models"
	"devais.it/kronos/internal/pkg/services"
	"devais.it/kronos/internal/pkg/sync/messages"
	"devais.it/kronos/internal/pkg/types"
	"devais.it/kronos/internal/pkg/util"
	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"
	"testing"
)

const (
	mbTest = "TEST"
)

type MessageHandlersSuite struct {
	db.SuiteBase
}

func (s *MessageHandlersSuite) Marshal(model interface{}) map[string]interface{} {
	assert := s.Require()
	payload, err := util.StructToJSONMap(model)
	assert.NoError(err)
	return payload
}

func (s *MessageHandlersSuite) newItem() *models.Item {
	id := uuid.NewString()
	return &models.Item{
		ID:   id,
		Name: "TestItem-" + id,
		Type: "TestItem",
	}
}

func (s *MessageHandlersSuite) newAttribute(itemID string) *models.Attribute {
	id := uuid.NewString()
	return &models.Attribute{
		ID:     id,
		Name:   "TestAttribute-" + id,
		Type:   "TestAttribute",
		ItemID: itemID,
	}
}

func (s *MessageHandlersSuite) TestSyncEntry() {
	batchSize := 20

	assert := s.Require()

	// Test create
	item := s.newItem()
	for i := 0; i < batchSize; i++ {
		item.Attributes = append(item.Attributes, *s.newAttribute(item.ID))
	}

	ctx := &db.TxContext{Tx: db.DB(), TxLen: 1, TxUUID: uuid.NewString()}

	entry := &messages.SyncEntry{
		EntityType: types.EntityTypeItem,
		EntityID:   item.ID,
		Payload:    s.Marshal(item),
	}

	// Test error
	entry.Action = "FAKE"
	err := syncEntry(ctx, entry)
	assert.ErrorIs(err, ErrInvalidAction)

	entry.Action = messages.SyncActionCreate
	entry.EntityType = "FAKE"

	err = syncEntry(ctx, entry)
	assert.ErrorIs(err, ErrInvalidEntityType)

	entry.EntityType = types.EntityTypeItem

	err = syncEntry(ctx, entry)
	assert.NoError(err)

	count, err := services.GetItemsCount()
	assert.NoError(err)
	assert.Equal(int64(1), count)

	count, err = services.GetAttributesCount()
	assert.NoError(err)
	assert.Equal(int64(batchSize), count)

	createdItem, err := services.GetItemByID(item.ID)
	assert.NoError(err)
	assert.Equal(item.Name, createdItem.Name)
	assert.Equal(item.Type, createdItem.Type)
	assert.Equal(constants.ModifiedBySyncName, createdItem.ModifiedBy)

	attribute := s.newAttribute(item.ID)

	entry.EntityID = attribute.ID
	entry.EntityType = types.EntityTypeAttribute
	entry.Payload = s.Marshal(attribute)

	err = syncEntry(ctx, entry)
	assert.NoError(err)

	count, err = services.GetAttributesCount()
	assert.NoError(err)
	assert.Equal(int64(batchSize+1), count)

	parent := s.newItem()
	child := s.newItem()

	assert.NoError(services.CreateItem(parent, mbTest))
	assert.NoError(services.CreateItem(child, mbTest))

	// Test relation composite ID error
	entry = &messages.SyncEntry{
		EntityID:   parent.ID + "@" + child.ID,
		EntityType: types.EntityTypeRelation,
		Action:     messages.SyncActionCreate,
	}

	err = syncEntry(ctx, entry)
	assert.Error(err)

	entry.EntityID = (&models.Relation{ParentID: parent.ID, ChildID: child.ID}).CompositeID()

	err = syncEntry(ctx, entry)
	assert.NoError(err)

	count, err = services.GetRelationsCount()
	assert.NoError(err)
	assert.Equal(int64(1), count)

	// Test update
	newName := "NewItemName"

	entry = &messages.SyncEntry{
		EntityID:   item.ID,
		EntityType: types.EntityTypeItem,
		Action:     messages.SyncActionUpdate,
		Payload:    map[string]interface{}{"name": newName},
	}

	err = syncEntry(ctx, entry)
	assert.NoError(err)

	updatedItem, err := services.GetItemByID(item.ID)
	assert.NoError(err)
	assert.Equal(newName, updatedItem.Name)
	assert.Equal(constants.ModifiedBySyncName, updatedItem.ModifiedBy)

	newName = "NewAttributeName"

	entry = &messages.SyncEntry{
		EntityID:   attribute.ID,
		EntityType: types.EntityTypeAttribute,
		Action:     messages.SyncActionUpdate,
		Payload:    map[string]interface{}{"name": newName},
	}

	err = syncEntry(ctx, entry)
	assert.NoError(err)

	updatedAttribute, err := services.GetAttributeByID(attribute.ID)
	assert.NoError(err)
	assert.Equal(newName, updatedAttribute.Name)
	assert.Equal(constants.ModifiedBySyncName, updatedAttribute.ModifiedBy)

	// Test upsert
	item2 := s.newItem()

	entry = &messages.SyncEntry{
		EntityID:   item2.ID,
		EntityType: types.EntityTypeItem,
		Action:     messages.SyncActionUpdate,
		Payload:    s.Marshal(item2),
	}

	err = syncEntry(ctx, entry)
	assert.NoError(err)

	createdItem, err = services.GetItemByID(item2.ID)
	assert.NoError(err)
	assert.Equal(item2.ID, createdItem.ID)
	assert.Equal(item2.Name, createdItem.Name)
	assert.Equal(constants.ModifiedBySyncName, createdItem.CreatedBy)

	assert.NoError(services.DeleteItem(item2, "TEST"))

	// Test delete
	entry = &messages.SyncEntry{
		EntityID:   attribute.ID,
		EntityType: types.EntityTypeAttribute,
		Action:     messages.SyncActionDelete,
	}

	err = syncEntry(ctx, entry)
	assert.NoError(err)

	_, err = services.GetAttributeByID(attribute.ID)
	assert.ErrorIs(err, gorm.ErrRecordNotFound)

	entry = &messages.SyncEntry{
		EntityID:   "FakeUUID",
		EntityType: types.EntityTypeItem,
		Action:     messages.SyncActionDelete,
	}

	err = syncEntry(ctx, entry)
	// Should not give any error when the entity to delete
	// doesn't exist
	assert.NoError(err)

	entry.EntityID = item.ID

	err = syncEntry(ctx, entry)
	assert.NoError(err)

	entry = &messages.SyncEntry{
		EntityID:   parent.ID + " -> " + child.ID,
		EntityType: types.EntityTypeRelation,
		Action:     messages.SyncActionDelete,
	}
	err = syncEntry(ctx, entry)
	assert.NoError(err)
	assert.NoError(services.DeleteItem(parent, mbTest))
	assert.NoError(services.DeleteItem(child, mbTest))

	count, err = services.GetItemsCount()
	assert.NoError(err)
	assert.Equal(int64(0), count)

	count, err = services.GetAttributesCount()
	assert.NoError(err)
	assert.Equal(int64(0), count)

	count, err = services.GetRelationsCount()
	assert.NoError(err)
	assert.Equal(int64(0), count)
}

func TestMessageHandlers(t *testing.T) {
	suite.Run(t, new(MessageHandlersSuite))
}
