package services

import (
	"testing"

	"devais.it/kronos/internal/pkg/constants"
	"devais.it/kronos/internal/pkg/db"
	"devais.it/kronos/internal/pkg/db/models"
	"devais.it/kronos/internal/pkg/util"
	"github.com/google/uuid"
	"github.com/rotisserie/eris"
	"github.com/stretchr/testify/suite"
	"gopkg.in/guregu/null.v4"
	"gorm.io/gorm"
)

const (
	mbItem         = "ITEMS_TEST"
	itemsBatchSize = 50
)

type ItemsSuite struct {
	db.SuiteBase
}

func newItem() *models.Item {
	id := uuid.NewString()
	return &models.Item{
		ID:   id,
		Name: "TestItem-" + id,
		Type: "TestItem",
	}
}

func (s *ItemsSuite) assertCount(expected int) {
	assert := s.Require()
	count, err := GetItemsCount()
	assert.NoError(err)
	assert.Equal(int64(expected), count)
}

func (s *ItemsSuite) assertContainsID(items []models.Item, id string) {
	assert := s.Require()
	contains := false
	for _, item := range items {
		if item.ID == id {
			contains = true
			break
		}
	}
	assert.Truef(contains, "List doesn't contain item '%s'", id)
}

func (s *ItemsSuite) TestGet() {
	assert := s.Require()

	s.assertCount(0)

	_, err := GetItemByID("")
	assert.ErrorIs(err, db.ErrMissingID)

	_, err = GetItemByName("")
	assert.Error(err)

	item, err := GetItemByID(uuid.NewString())
	assert.ErrorIs(err, gorm.ErrRecordNotFound)
	assert.Nil(item)

	item, err = GetItemByName("FakeItemName")
	assert.ErrorIs(err, gorm.ErrRecordNotFound)
	assert.Nil(item)

	item = newItem()
	err = CreateItem(item, mbItem)
	assert.NoError(err)

	createdItem, err := GetItemByID(item.ID)
	assert.NoError(err)
	assert.NotNil(createdItem)
	assert.Equal(item.ID, createdItem.ID)
	assert.Equal(item.Name, createdItem.Name)
	assert.Equal(item.Type, createdItem.Type)

	createdItem, err = GetItemByName(item.Name)
	assert.NoError(err)
	assert.NotNil(createdItem)
	assert.Equal(item.ID, createdItem.ID)
	assert.Equal(item.Name, createdItem.Name)
	assert.Equal(item.Type, createdItem.Type)

	assert.NoError(DeleteItemByID(item.ID, mbItem))

	items := make([]models.Item, itemsBatchSize)
	ids := make([]string, itemsBatchSize)
	for i := 0; i < itemsBatchSize; i++ {
		nItem := newItem()
		items[i] = *nItem
		ids[i] = nItem.ID
		assert.NoError(CreateItem(nItem, mbItem))
	}

	s.assertCount(itemsBatchSize)

	// Test pagination
	halfSize := itemsBatchSize / 2
	createdItems, err := GetAllItems(1, halfSize)
	assert.NoError(err)
	assert.Len(createdItems, halfSize)

	createdItems, err = GetAllItems(2, itemsBatchSize-halfSize)
	assert.NoError(err)
	assert.Len(createdItems, itemsBatchSize-halfSize)

	// Get all items
	createdItems, err = GetAllItems(-1, -1)
	assert.Error(err)
	assert.Empty(createdItems)

	createdItems, err = GetAllItems(1, itemsBatchSize)
	assert.NoError(err)
	assert.Len(createdItems, itemsBatchSize)

	createdItems, err = GetItemsByType(uuid.NewString(), 0, itemsBatchSize)
	assert.ErrorIs(err, gorm.ErrRecordNotFound)
	assert.Empty(createdItems)

	createdItems, err = GetItemsByType(item.Type, 0, itemsBatchSize)
	assert.NoError(err)
	assert.Len(createdItems, itemsBatchSize)

	createdItems, err = GetItemsByIDs(ids)
	assert.NoError(err)
	assert.Len(createdItems, itemsBatchSize)

	// Check GetItemsByIDs errors
	createdItems, err = GetItemsByIDs(nil)
	assert.ErrorIs(err, db.ErrMissingID)
	assert.Empty(createdItems)

	ids = append(ids, uuid.NewString())
	createdItems, err = GetItemsByIDs(ids)
	assert.ErrorIs(err, gorm.ErrRecordNotFound)
	assert.Empty(createdItems)
}

func (s *ItemsSuite) TestCreate() {
	assert := s.Require()

	item := newItem()

	start := util.TimestampMs()
	assert.NoError(CreateItem(item, mbItem))
	end := util.TimestampMs()

	s.assertCount(1)

	createdItem, err := GetItemByID(item.ID)
	assert.NoError(err)

	assert.Equal(item.ID, createdItem.ID)
	assert.Equal(item.Name, createdItem.Name)
	assert.Equal(item.Type, createdItem.Type)
	assert.GreaterOrEqual(createdItem.CreatedAt, start)
	assert.GreaterOrEqual(createdItem.ModifiedAt, start)
	assert.LessOrEqual(createdItem.CreatedAt, end)
	assert.LessOrEqual(createdItem.ModifiedAt, end)
	assert.Equal(mbItem, createdItem.CreatedBy)
	assert.Equal(mbItem, createdItem.ModifiedBy)

	assert.Zero(createdItem.SourceTimestamp)
	assert.Nil(createdItem.CustomerID.Ptr())
	assert.True(createdItem.EdgeMac.IsZero())
	assert.NotNil(createdItem.Version)
	assert.Nil(createdItem.SyncVersion.Ptr())
	assert.Empty(createdItem.Attributes)

	// Test get by name
	createdItem2, err := GetItemByName(item.Name)
	assert.NoError(err)
	assert.NotNil(createdItem2)
	assert.Equal(createdItem, createdItem2)

	// Test get all
	allItems, err := GetAllItems(0, 0)
	assert.NoError(err)
	s.assertContainsID(allItems, item.ID)
}

func (s *ItemsSuite) TestCreateAttributes() {
	assert := s.Require()

	item := newItem()

	attributes := make([]models.Attribute, itemsBatchSize)
	attributesMap := make(map[string]models.Attribute, itemsBatchSize)

	for i := 0; i < itemsBatchSize; i++ {
		id := uuid.NewString()
		attributes[i].ID = id
		attributes[i].Name = "Attribute-" + id
		attributes[i].Type = uuid.NewString()

		attributesMap[id] = attributes[i]
	}

	item.Attributes = attributes

	attributes[0].ItemID = uuid.NewString()
	assert.Error(CreateItem(item, mbItem))
	s.assertCount(0)

	attributes[0].ItemID = ""
	assert.NoError(CreateItem(item, mbItem))

	createdAttributes, err := GetItemAttributes(item.ID)
	assert.NoError(err)
	assert.Len(createdAttributes, len(attributes))

	for _, createdAttribute := range createdAttributes {
		attribute := attributesMap[createdAttribute.ID]

		assert.Equal(attribute.ID, createdAttribute.ID)
		assert.Equal(attribute.Name, createdAttribute.Name)
		assert.Equal(attribute.Type, createdAttribute.Type)
	}
}

func (s *ItemsSuite) TestGetAttributes() {
	assert := s.Require()

	item := newItem()
	item.CustomerID = null.StringFrom("FakeCustomer")
	item.EdgeMac = null.StringFrom("01:02:03:04:05:06")
	item.SyncPolicy = null.StringFrom(constants.SyncPolicyDontSync)

	err := CreateItem(item, mbItem)
	assert.NoError(err)

	created, err := GetItemByID(item.ID)
	assert.NoError(err)

	assert.Equal(item.ID, created.ID)
	assert.Equal(item.Name, created.Name)
	assert.Equal(item.Type, created.Type)

	errID := "err" + uuid.NewString()

	// Check error
	itemMac, err := GetItemMac(errID)
	assert.ErrorIs(err, gorm.ErrRecordNotFound)
	assert.Empty(itemMac.ValueOrZero())

	// Check edge_mac
	itemMac, err = GetItemMac(item.ID)
	assert.NoError(err)
	assert.Equal(itemMac, item.EdgeMac)

	// Check error
	createdBy, err := GetItemCreatedBy(errID)
	assert.ErrorIs(err, gorm.ErrRecordNotFound)
	assert.Empty(createdBy)

	// Check created_by
	createdBy, err = GetItemCreatedBy(item.ID)
	assert.NoError(err)
	assert.Equal(mbItem, createdBy)

	// Check error
	modifiedBy, err := GetItemModifiedBy(errID)
	assert.ErrorIs(err, gorm.ErrRecordNotFound)
	assert.Empty(modifiedBy)

	// Check modified_by
	modifiedBy, err = GetItemModifiedBy(item.ID)
	assert.NoError(err)
	assert.Equal(mbItem, modifiedBy)

	// Check error
	version, err := GetItemVersion(errID)
	assert.ErrorIs(err, gorm.ErrRecordNotFound)
	assert.Empty(version)

	// Check version
	version, err = GetItemVersion(item.ID)
	assert.NoError(err)
	assert.Equal(created.Version, version)

	// Check error
	syncPolicy, err := GetItemSyncPolicy(errID)
	assert.ErrorIs(err, gorm.ErrRecordNotFound)
	assert.True(syncPolicy.IsZero())

	// Check sync policy
	syncPolicy, err = GetItemSyncPolicy(item.ID)
	assert.NoError(err)
	assert.Equal(created.SyncPolicy, syncPolicy)

	// Check error
	lastUpdate, err := GetItemLastUpdateTime(errID)
	assert.ErrorIs(err, gorm.ErrRecordNotFound)
	assert.Zero(lastUpdate)

	// Check last update
	lastUpdate, err = GetItemLastUpdateTime(item.ID)
	assert.NoError(err)
	assert.NotZero(lastUpdate)
	assert.Equal(created.ModifiedAt, lastUpdate)

	// Check error
	customerID, err := GetItemCustomerID(errID)
	assert.ErrorIs(err, gorm.ErrRecordNotFound)
	assert.False(customerID.Valid)

	customerID, err = GetItemCustomerID(item.ID)
	assert.NoError(err)
	assert.True(customerID.Valid)
	assert.Equal(item.CustomerID, customerID)
}

func (s *ItemsSuite) TestBatchCreate() {
	assert := s.Require()

	err := BatchCreateItems([]models.Item{}, mbItem)
	assert.ErrorIs(err, ErrEmptySlice)

	err = BatchCreateItems(nil, mbItem)
	assert.ErrorIs(err, ErrEmptySlice)

	items := make([]models.Item, itemsBatchSize)

	for i := 0; i < itemsBatchSize; i++ {
		items[i] = *newItem()
	}

	items[5].Type = ""

	// Test failure
	err = BatchCreateItems(items, mbItem)
	assert.Error(err)

	// No items should be created
	s.assertCount(0)

	items[5].Type = "TestItem"

	err = BatchCreateItems(items, mbItem)
	assert.NoError(err)

	s.assertCount(itemsBatchSize)

	createdItems, err := GetAllItems(1, itemsBatchSize)
	assert.NoError(err)
	assert.Len(createdItems, itemsBatchSize)

	createdMap := make(map[string]models.Item, itemsBatchSize)
	createdIds := make([]string, itemsBatchSize)
	for i, item := range createdItems {
		createdMap[item.ID] = item
		createdIds[i] = item.ID
	}

	for i := 0; i < itemsBatchSize; i++ {
		item := items[i]

		assert.Containsf(createdMap, item.ID, "Item '%'s not created", item.ID)

		createdItem := createdMap[item.ID]

		assert.Equal(item.ID, createdItem.ID)
		assert.Equal(item.Name, createdItem.Name)
		assert.Equal(item.Type, createdItem.Type)
	}

	itemsByID, err := GetItemsByIDs(createdIds)
	assert.NoError(err)
	assert.Len(itemsByID, itemsBatchSize)

	for i := 0; i < itemsBatchSize; i++ {
		item := itemsByID[i]
		createdItem := createdMap[item.ID]

		assert.Equal(item.ID, createdItem.ID)
		assert.Equal(item.Name, createdItem.Name)
		assert.Equal(item.Type, createdItem.Type)

		assert.NoError(DeleteItemByID(item.ID, mbItem))
	}

	s.assertCount(0)
}

func (s *ItemsSuite) TestUpdate() {
	assert := s.Require()

	mbUpdate := "TEST_UPDATE"

	item := newItem()

	newName := "ItemNewName"
	newCustomer := "FakeCustomer"
	newSourceTs := uint64(12)

	patch := map[string]interface{}{
		"name":             newName,
		"customer_id":      newCustomer,
		"source_timestamp": newSourceTs,
	}

	err := UpdateItem(patch, mbUpdate)
	assert.Error(err)
	assert.ErrorIs(err, db.ErrMissingID)

	patch["id"] = item.ID
	patch["fake_field"] = 20

	err = UpdateItem(patch, mbUpdate)
	assert.Error(err)
	assert.NotErrorIs(err, db.ErrMissingID)
	assert.Contains(eris.ToString(err, true), "no such column")

	delete(patch, "fake_field")

	err = UpdateItem(patch, mbUpdate)
	assert.ErrorIs(err, gorm.ErrRecordNotFound)

	err = CreateItem(item, mbItem)
	assert.NoError(err)

	err = UpdateItem(patch, mbUpdate)
	assert.NoError(err)

	updatedItem, err := GetItemByID(item.ID)
	assert.NoError(err)
	assert.Equal(newName, updatedItem.Name)
	assert.Equal(mbUpdate, updatedItem.ModifiedBy)
	assert.Equal(newCustomer, updatedItem.CustomerID.ValueOrZero())
	assert.Equal(newSourceTs, updatedItem.SourceTimestamp)

	newModifiedBy := "TEST_UPDATE2"
	patch[constants.ModifiedByField] = newModifiedBy

	err = UpdateItem(patch, mbItem)
	assert.NoError(err)

	mb, err := GetItemModifiedBy(item.ID)
	assert.NoError(err)
	assert.Equal(newModifiedBy, mb)

	assert.NoError(DeleteItem(item, mbItem))
}

func (s *ItemsSuite) TestNestedUpdate() {
	assert := s.Require()

	mbUpdate := "TEST_UPDATE"

	item := newItem()
	attr := newAttribute(item.ID)

	patch := map[string]interface{}{
		"id": item.ID,
	}

	err := CreateItem(item, mbItem)
	assert.NoError(err)

	err = CreateAttribute(attr, mbItem)
	assert.NoError(err)

	patch["attributes"] = 10
	err = UpdateItem(patch, mbUpdate)
	assert.Error(err)

	patch["attributes"] = []string{"a", "b"}
	err = UpdateItem(patch, mbUpdate)
	assert.Error(err)

	newItemName := "NewItemName"
	newAttrName := "NewNestedAttrName"

	patch["name"] = newItemName
	patch["attributes"] = []map[string]interface{}{
		{"id": attr.ID, "name": newAttrName},
	}

	err = UpdateItem(patch, mbUpdate)
	assert.NoError(err)

	updatedItem, err := GetItemByID(item.ID)
	assert.NoError(err)
	assert.Equal(newItemName, updatedItem.Name)

	updatedAttr, err := GetAttributeByID(attr.ID)
	assert.NoError(err)
	assert.Equal(newAttrName, updatedAttr.Name)

	// Test creating attributes
	newAttr := newAttribute(item.ID)
	patch["attributes"] = []map[string]interface{}{
		{
			"id":      newAttr.ID,
			"name":    newAttr.Name,
			"item_id": newAttr.ItemID,
			"type":    newAttr.Type,
		},
	}

	err = UpdateItem(patch, mbUpdate)
	assert.NoError(err)

	createdAttr, err := GetAttributeByID(newAttr.ID)
	assert.NoError(err)
	assert.Equal(newAttr.Name, createdAttr.Name)
	assert.Equal(newAttr.Type, createdAttr.Type)
	assert.Equal(newAttr.ItemID, createdAttr.ItemID)
}

func (s *ItemsSuite) TestUpsert() {
	assert := s.Require()

	item := newItem()

	patch := map[string]interface{}{
		"name": item.Name,
	}

	err := UpsertItem(patch, mbItem)
	assert.ErrorIs(err, db.ErrMissingID)

	patch["id"] = item.ID

	err = UpsertItem(patch, mbItem)
	assert.Error(err)
	assert.NotErrorIs(err, gorm.ErrRecordNotFound)

	patch["type"] = item.Type

	err = UpsertItem(patch, mbItem)
	assert.NoError(err)

	createdItem, err := GetItemByID(item.ID)
	assert.NoError(err)
	assert.Equal(item.Name, createdItem.Name)

	newItemName := "NewItemName"
	patch["name"] = newItemName
	err = UpsertItem(patch, mbItem)
	assert.NoError(err)

	createdItem, err = GetItemByID(item.ID)
	assert.NoError(err)
	assert.Equal(newItemName, createdItem.Name)
}

func (s *ItemsSuite) TestDelete() {
	assert := s.Require()

	item := newItem()

	// Test failure
	assert.ErrorIs(DeleteItemByID("", mbItem), db.ErrMissingID)
	assert.ErrorIs(HardDeleteItemByID("", mbItem), db.ErrMissingID)

	assert.ErrorIs(DeleteItemByID(item.ID, mbItem), gorm.ErrRecordNotFound)

	assert.ErrorIs(DeleteItem(item, mbItem), gorm.ErrRecordNotFound)

	assert.ErrorIs(HardDeleteItemByID(item.ID, mbItem), gorm.ErrRecordNotFound)

	// Create item
	assert.NoError(CreateItem(item, mbItem))

	s.assertCount(1)

	assert.NoError(DeleteItemByID(item.ID, mbItem))

	s.assertCount(0)

	deletedItem, err := GetItemByID(item.ID)
	assert.Error(err)
	assert.ErrorIs(err, gorm.ErrRecordNotFound)
	assert.Nil(deletedItem)

	item = newItem()
	assert.NoError(CreateItem(item, mbItem))
	s.assertCount(1)
	assert.NoError(DeleteItem(item, mbItem))
	s.assertCount(0)

	item = newItem()
	assert.NoError(CreateItem(item, mbItem))
	s.assertCount(1)
	assert.NoError(HardDeleteItemByID(item.ID, mbItem))
	s.assertCount(0)
}

func (s *ItemsSuite) TestVersions() {
	assert := s.Require()

	items := make([]models.Item, itemsBatchSize)
	for i := 0; i < itemsBatchSize; i++ {
		nItem := newItem()
		items[i] = *nItem
		assert.NoError(CreateItem(nItem, mbItem))
	}

	s.assertCount(itemsBatchSize)

	createdItems, err := GetAllItems(1, itemsBatchSize)
	assert.NoError(err)
	assert.Len(createdItems, itemsBatchSize)

	versions, err := GetItemsVersion(-1, itemsBatchSize)
	assert.Error(err)
	assert.Empty(versions)

	versions, err = GetItemsVersion(1, itemsBatchSize)
	assert.NoError(err)
	assert.Len(versions, itemsBatchSize)
	versionsMap := make(map[string]models.EntityVersion, itemsBatchSize)

	for _, version := range versions {
		versionsMap[version.ID] = version
	}

	algo := util.VersionAlgorithm(0)

	for _, item := range createdItems {
		fields := map[string]interface{}{
			"id":          item.ID,
			"name":        item.Name,
			"type":        item.Type,
			"created_by":  mbItem,
			"modified_by": mbItem,
			// Empty fields
			"edge_mac":         nil,
			"source_timestamp": 0,
			"customer_id":      nil,
			"sync_policy":      nil,
		}

		fieldsBytes, err := json.Marshal(fields)
		assert.NoError(err)

		computed, err := util.BytesChecksum(fieldsBytes, algo)
		assert.NoError(err)

		version, err := GetItemVersion(item.ID)
		assert.NoError(err)

		assert.Equal(computed, item.Version)
		assert.Equal(computed, version)
		assert.Equal(computed, versionsMap[item.ID].Version)

		computed2, err := util.GenerateVersionChecksum(item, algo)
		assert.NoError(err)

		assert.Equal(computed, computed2)
	}
}

func TestItemsService(t *testing.T) {
	suite.Run(t, new(ItemsSuite))
}
