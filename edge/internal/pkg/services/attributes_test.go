package services

import (
	"gopkg.in/guregu/null.v4"
	"testing"

	"devais.it/kronos/internal/pkg/constants"
	"devais.it/kronos/internal/pkg/db"
	"devais.it/kronos/internal/pkg/db/models"
	"devais.it/kronos/internal/pkg/util"
	"github.com/google/uuid"
	"github.com/rotisserie/eris"
	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"
)

const (
	mbAttribute         = "ATTRIBUTES_TEST"
	attributesBatchSize = 50
)

type AttributesSuite struct {
	db.SuiteBase
}

func newAttribute(itemID string) *models.Attribute {
	id := uuid.NewString()
	return &models.Attribute{
		ID:     id,
		Name:   "TestAttribute-" + id,
		Type:   uuid.NewString(),
		ItemID: itemID,
	}
}

func (s *AttributesSuite) assertCount(expected int) {
	assert := s.Require()
	count, err := GetAttributesCount()
	assert.NoError(err)
	assert.Equal(int64(expected), count)
}

func (s *AttributesSuite) assertContainsID(attributes []models.Attribute, id string) {
	assert := s.Require()
	contains := false
	for _, attribute := range attributes {
		if attribute.ID == id {
			contains = true
			break
		}
	}
	assert.True(contains, "List doesn't contain attribute '%s'", id)
}

func (s *AttributesSuite) TestGet() {
	assert := s.Require()

	s.assertCount(0)

	_, err := GetAttributeByID("")
	assert.ErrorIs(err, db.ErrMissingID)

	_, err = GetAttributeByID(uuid.NewString())
	assert.ErrorIs(err, gorm.ErrRecordNotFound)

	item := newItem()
	assert.NoError(CreateItem(item, mbAttribute))

	attribute := newAttribute(item.ID)
	assert.NoError(CreateAttribute(attribute, mbAttribute))

	s.assertCount(1)

	createdAttr, err := GetAttributeByID(attribute.ID)
	assert.NoError(err)
	assert.NotNil(createdAttr)
	assert.Equal(attribute.ID, createdAttr.ID)
	assert.Equal(attribute.Name, createdAttr.Name)
	assert.Equal(attribute.Type, createdAttr.Type)

	assert.NoError(DeleteAttribute(attribute, mbAttribute))

	s.assertCount(0)

	attributes := make([]models.Attribute, attributesBatchSize)
	ids := make([]string, attributesBatchSize)
	for i := 0; i < attributesBatchSize; i++ {
		nAttr := newAttribute(item.ID)
		attributes[i] = *nAttr
		ids[i] = nAttr.ID
		assert.NoError(CreateAttribute(nAttr, mbAttribute))
	}

	s.assertCount(attributesBatchSize)

	// Test pagination
	createdAttributes, err := GetAllAttributes(-1, -1)
	assert.Error(err)
	assert.Empty(createdAttributes)

	halfSize := attributesBatchSize / 2
	createdAttributes, err = GetAllAttributes(1, halfSize)
	assert.NoError(err)
	assert.Len(createdAttributes, halfSize)

	createdAttributes, err = GetAllAttributes(2, attributesBatchSize-halfSize)
	assert.NoError(err)
	assert.Len(createdAttributes, attributesBatchSize-halfSize)

	// Get all attributes
	createdAttributes, err = GetAllAttributes(1, attributesBatchSize)
	assert.NoError(err)
	assert.Len(createdAttributes, attributesBatchSize)

	// Get by ID
	createdAttributes, err = GetAttributesByIDs(nil)
	assert.ErrorIs(err, db.ErrMissingID)
	assert.Empty(createdAttributes)

	createdAttributes, err = GetAttributesByIDs(append(ids, uuid.NewString()))
	assert.ErrorIs(err, gorm.ErrRecordNotFound)
	assert.Empty(createdAttributes)

	createdAttributes, err = GetAttributesByIDs(ids)
	assert.NoError(err)
	assert.Len(createdAttributes, attributesBatchSize)
}

func (s *AttributesSuite) TestCreate() {
	assert := s.Require()

	// Create without item ID
	err := CreateAttribute(newAttribute(""), mbAttribute)
	assert.Error(err)

	item := newItem()
	attribute := newAttribute(item.ID)

	assert.NoError(CreateItem(item, mbAttribute))

	start := util.TimestampMs()
	assert.NoError(CreateAttribute(attribute, mbAttribute))
	end := util.TimestampMs()

	s.assertCount(1)

	createdAttr, err := GetAttributeByID(attribute.ID)
	assert.NoError(err)

	assert.Equal(attribute.ID, createdAttr.ID)
	assert.Equal(attribute.Name, createdAttr.Name)
	assert.Equal(attribute.Type, createdAttr.Type)
	assert.Equal(attribute.ItemID, createdAttr.ItemID)
	assert.GreaterOrEqual(createdAttr.CreatedAt, start)
	assert.GreaterOrEqual(createdAttr.ModifiedAt, start)
	assert.LessOrEqual(createdAttr.CreatedAt, end)
	assert.LessOrEqual(createdAttr.ModifiedAt, end)
	assert.Empty(createdAttr.SyncPolicy)
	assert.Empty(createdAttr.SyncVersion.ValueOrZero())
	assert.NotNil(createdAttr.Version)

	allAttributes, err := GetAllAttributes(1, 1)
	assert.NoError(err)
	assert.Len(allAttributes, 1)
	s.assertContainsID(allAttributes, attribute.ID)
}

func (s *AttributesSuite) TestGetAttributes() {
	assert := s.Require()

	item := newItem()

	itemAttributes, err := GetItemAttributes(item.ID)
	assert.ErrorIs(err, gorm.ErrRecordNotFound)
	assert.Empty(itemAttributes)

	attribute := newAttribute(item.ID)
	attribute.SyncPolicy = null.StringFrom(constants.SyncPolicyDontSync)

	assert.NoError(CreateItem(item, mbAttribute))
	assert.NoError(CreateAttribute(attribute, mbAttribute))

	created, err := GetAttributeByID(attribute.ID)
	assert.NoError(err)

	assert.Equal(attribute.ID, created.ID)
	assert.Equal(attribute.Name, created.Name)
	assert.Equal(attribute.Type, created.Type)
	assert.Equal(attribute.ItemID, created.ItemID)

	errID := "err" + uuid.NewString()

	// Check error
	version, err := GetAttributeVersion(errID)
	assert.ErrorIs(err, gorm.ErrRecordNotFound)
	assert.Empty(version)

	// Check version
	version, err = GetAttributeVersion(attribute.ID)
	assert.NoError(err)
	assert.Equal(created.Version, version)

	// Check error
	syncPolicy, err := GetAttributesSyncPolicy(errID)
	assert.ErrorIs(err, gorm.ErrRecordNotFound)
	assert.True(syncPolicy.IsZero())

	// Check sync policy
	syncPolicy, err = GetAttributesSyncPolicy(attribute.ID)
	assert.NoError(err)
	assert.Equal(created.SyncPolicy, syncPolicy)
}

func (s *AttributesSuite) TestBatchCreate() {
	assert := s.Require()

	err := BatchCreateAttributes([]models.Attribute{}, mbAttribute)
	assert.ErrorIs(err, ErrEmptySlice)

	err = BatchCreateAttributes(nil, mbAttribute)
	assert.ErrorIs(err, ErrEmptySlice)

	item := newItem()
	attributes := make([]models.Attribute, attributesBatchSize)

	for i := 0; i < attributesBatchSize; i++ {
		attributes[i] = *newAttribute(item.ID)
	}

	halfSize := attributesBatchSize / 2

	prevName := attributes[halfSize].Name
	attributes[halfSize].Name = ""

	assert.NoError(CreateItem(item, mbAttribute))

	// Test failure
	err = BatchCreateAttributes(attributes, mbAttribute)
	assert.Error(err)

	// No attributes should be created
	s.assertCount(0)

	attributes[halfSize].Name = prevName

	err = BatchCreateAttributes(attributes, mbAttribute)
	assert.NoError(err)

	s.assertCount(attributesBatchSize)

	createdAttributes, err := GetAllAttributes(1, attributesBatchSize)
	assert.NoError(err)
	assert.Len(createdAttributes, attributesBatchSize)

	createdMap := make(map[string]models.Attribute, attributesBatchSize)
	createdIds := make([]string, attributesBatchSize)
	for i, attribute := range createdAttributes {
		createdMap[attribute.ID] = attribute
		createdIds[i] = attribute.ID
	}

	for i := 0; i < attributesBatchSize; i++ {
		attribute := attributes[i]

		assert.Containsf(createdMap, attribute.ID, "Attribute '%s' not created", attribute.ID)

		createdAttribute := createdMap[attribute.ID]

		assert.Equal(attribute.ID, createdAttribute.ID)
		assert.Equal(attribute.Name, createdAttribute.Name)
		assert.Equal(attribute.ItemID, createdAttribute.ItemID)
		assert.Equal(attribute.Type, createdAttribute.Type)

		// Delete
		assert.NoError(DeleteAttributeByID(attribute.ID, mbAttribute))
	}

	s.assertCount(0)
}

func (s *AttributesSuite) TestUpdate() {
	assert := s.Require()

	mbUpdate := "TEST_UPDATE"

	item := newItem()
	attribute := newAttribute(item.ID)

	newName := "AttributeNewName"

	patch := map[string]interface{}{
		"name": newName,
	}

	assert.ErrorIs(UpdateAttribute(patch, mbUpdate), db.ErrMissingID)

	patch["id"] = attribute.ID
	patch["fake_field"] = 20

	err := UpdateAttribute(patch, mbUpdate)
	assert.Error(err)
	assert.NotErrorIs(err, db.ErrMissingID)
	assert.Contains(eris.ToString(err, true), "no such column")

	delete(patch, "fake_field")

	err = UpdateAttribute(patch, mbUpdate)
	assert.ErrorIs(err, gorm.ErrRecordNotFound)

	assert.NoError(CreateItem(item, mbAttribute))
	assert.NoError(CreateAttribute(attribute, mbAttribute))

	err = UpdateAttribute(patch, mbUpdate)
	assert.NoError(err)

	updatedAttr, err := GetAttributeByID(attribute.ID)
	assert.NoError(err)
	assert.Equal(newName, updatedAttr.Name)

	newModifiedBy := "TEST_UPDATE2"
	patch[constants.ModifiedByField] = newModifiedBy

	err = UpdateAttribute(patch, mbAttribute)
	assert.NoError(err)

	updatedAttr, err = GetAttributeByID(attribute.ID)
	assert.NoError(err)
	assert.Equal(newModifiedBy, updatedAttr.ModifiedBy)

	assert.NoError(DeleteItem(item, mbAttribute))
}

func (s *AttributesSuite) TestNestedUpdate() {
	assert := s.Require()

	mbUpdate := "TEST_UPDATE"

	item := newItem()
	attr := newAttribute(item.ID)

	patch := map[string]interface{}{
		"id": attr.ID,
	}

	err := CreateItem(item, mbAttribute)
	assert.NoError(err)

	err = CreateAttribute(attr, mbAttribute)
	assert.NoError(err)

	newAttributeName := "NewAttributeName"

	patch["name"] = newAttributeName

	err = UpdateAttribute(patch, mbUpdate)
	assert.NoError(err)

	updatedAttr, err := GetAttributeByID(attr.ID)
	assert.NoError(err)
	assert.Equal(newAttributeName, updatedAttr.Name)

	err = UpdateAttribute(patch, mbUpdate)
	assert.NoError(err)
}

func (s *AttributesSuite) TestUpsert() {
	assert := s.Require()

	item := newItem()
	attr := newAttribute(item.ID)

	patch := map[string]interface{}{
		"name":    attr.Name,
		"item_id": attr.ItemID,
	}

	err := CreateItem(item, mbAttribute)
	assert.NoError(err)

	err = UpsertAttribute(patch, mbAttribute)
	assert.ErrorIs(err, db.ErrMissingID)

	patch["id"] = attr.ID

	err = UpsertAttribute(patch, mbAttribute)
	assert.Error(err)
	assert.NotErrorIs(err, gorm.ErrRecordNotFound)

	patch["type"] = attr.Type

	err = UpsertAttribute(patch, mbAttribute)
	assert.NoError(err)

	createdAttr, err := GetAttributeByID(attr.ID)
	assert.NoError(err)
	assert.Equal(attr.Name, createdAttr.Name)

	newAttrName := "NewAttrName"
	patch["name"] = newAttrName
	err = UpsertAttribute(patch, mbAttribute)
	assert.NoError(err)

	createdAttr, err = GetAttributeByID(attr.ID)
	assert.NoError(err)
	assert.Equal(newAttrName, createdAttr.Name)
}

func (s *AttributesSuite) TestDelete() {
	assert := s.Require()

	item := newItem()
	attribute := newAttribute(item.ID)

	// Test failure
	assert.ErrorIs(DeleteAttributeByID("", mbAttribute), db.ErrMissingID)
	assert.ErrorIs(HardDeleteAttributeByID("", mbAttribute), db.ErrMissingID)

	assert.ErrorIs(DeleteAttributeByID(attribute.ID, mbAttribute), gorm.ErrRecordNotFound)
	assert.ErrorIs(DeleteAttribute(attribute, mbAttribute), gorm.ErrRecordNotFound)
	assert.ErrorIs(HardDeleteAttributeByID(attribute.ID, mbAttribute), gorm.ErrRecordNotFound)

	// Create item and attribute
	assert.NoError(CreateItem(item, mbAttribute))

	assert.NoError(CreateAttribute(attribute, mbAttribute))
	s.assertCount(1)
	assert.NoError(DeleteAttributeByID(attribute.ID, mbAttribute))
	s.assertCount(0)

	deletedAttribute, err := GetAttributeByID(attribute.ID)
	assert.ErrorIs(err, gorm.ErrRecordNotFound)
	assert.Nil(deletedAttribute)

	assert.NoError(CreateAttribute(attribute, mbAttribute))
	s.assertCount(1)
	assert.NoError(DeleteAttribute(attribute, mbAttribute))
	s.assertCount(0)

	assert.NoError(CreateAttribute(attribute, mbAttribute))
	s.assertCount(1)
	assert.NoError(HardDeleteAttributeByID(attribute.ID, mbAttribute))
	s.assertCount(0)

	// Delete item should delete associate attributes
	assert.NoError(CreateAttribute(attribute, mbAttribute))
	s.assertCount(1)
	assert.NoError(DeleteItem(item, mbAttribute))
	s.assertCount(0)

	assert.NoError(CreateItem(item, mbAttribute))

	attributes := make([]models.Attribute, attributesBatchSize)
	for i := 0; i < attributesBatchSize; i++ {
		attributes[i] = *newAttribute(item.ID)
	}

	assert.NoError(BatchCreateAttributes(attributes, mbAttribute))

	s.assertCount(attributesBatchSize)
	assert.NoError(DeleteItem(item, mbAttribute))
	s.assertCount(0)
}

func (s *AttributesSuite) TestVersions() {
	assert := s.Require()

	item := newItem()
	assert.NoError(CreateItem(item, mbAttribute))

	attributes := make([]models.Attribute, attributesBatchSize)
	for i := 0; i < attributesBatchSize; i++ {
		nAttr := newAttribute(item.ID)
		attributes[i] = *nAttr
		assert.NoError(CreateAttribute(nAttr, mbAttribute))
	}

	s.assertCount(attributesBatchSize)

	createdAttributes, err := GetAllAttributes(1, attributesBatchSize)
	assert.NoError(err)
	assert.Len(createdAttributes, attributesBatchSize)

	versions, err := GetAttributesVersion(-1, -1)
	assert.Error(err)
	assert.Empty(versions)

	versions, err = GetAttributesVersion(1, attributesBatchSize)
	assert.NoError(err)
	assert.Len(versions, attributesBatchSize)
	versionsMap := make(map[string]models.EntityVersion, attributesBatchSize)

	for _, version := range versions {
		versionsMap[version.ID] = version
	}

	algo := util.VersionAlgorithm(0)

	for _, attr := range createdAttributes {
		fields := map[string]interface{}{
			"id":               attr.ID,
			"name":             attr.Name,
			"type":             attr.Type,
			"value":            attr.Value,
			"value_type":       attr.ValueType,
			"item_id":          attr.ItemID,
			"created_by":       attr.CreatedBy,
			"modified_by":      attr.ModifiedBy,
			"source_timestamp": 0,
			"sync_policy":      nil,
		}

		fieldsBytes, err := json.Marshal(fields)
		assert.NoError(err)

		computed, err := util.BytesChecksum(fieldsBytes, algo)
		assert.NoError(err)

		version, err := GetAttributeVersion(attr.ID)
		assert.NoError(err)

		assert.Equal(computed, attr.Version)
		assert.Equal(computed, version)
		assert.Equal(computed, versionsMap[attr.ID].Version)

		computed2, err := util.GenerateVersionChecksum(attr, algo)
		assert.NoError(err)

		assert.Equal(computed, computed2)
	}

	assert.NoError(DeleteItem(item, mbAttribute))
}

func TestAttributesService(t *testing.T) {
	suite.Run(t, new(AttributesSuite))
}
