package db

import (
	"devais.it/kronos/internal/pkg/db/models"
	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"
	"testing"
)

type DbTestSuite struct {
	SuiteBase
}

func (s *DbTestSuite) TestCrud() {
	assert := s.Require()

	assertRowsNz := func(tx *gorm.DB) {
		assert.False(tx.RowsAffected == 0, "No rows affected")
	}

	fakeItem := &models.Item{
		ID:         "test-item-uuid",
		Name:       "TestItem00",
		Type:       "TestItem",
	}

	fakeItem.CreatedBy = "TEST"
	fakeItem.ModifiedBy = "TEST"

	// Create
	tx := db.Create(fakeItem)

	assert.NoError(tx.Error, "Failed to create a new item")
	assertRowsNz(tx)

	createdItem := &models.Item{}

	// Get
	tx = db.Where("id = ?", fakeItem.ID).First(createdItem)

	assert.NoError(tx.Error, "Failed to get created item")
	assertRowsNz(tx)

	assert.Equal(createdItem.ID, fakeItem.ID, "Created item ID doesn't match")
	assert.Equal(createdItem.Name, fakeItem.Name, "Created item Name doesn't match")
	assert.Equal(createdItem.Type, fakeItem.Type, "Created item Type doesn't match")

	// Update
	newItemName := "TestItemNewName00"
	fakeItem.Name = newItemName
	tx = db.Updates(fakeItem)

	assert.NoError(tx.Error, "Failed to update item")
	assertRowsNz(tx)

	tx = db.Where("id = ?", fakeItem.ID).First(createdItem)

	assert.NoError(tx.Error, "Failed to get updated item")
	assertRowsNz(tx)

	assert.Equal(createdItem.Name, fakeItem.Name, "Updated item name doesn't match")

	// Delete
	tx = db.Delete(createdItem)

	assert.NoError(tx.Error, "Failed to delete created item")
	assertRowsNz(tx)

	tx = db.Where("id = ?", fakeItem.ID).First(createdItem)
	assert.LessOrEqual(tx.RowsAffected, int64(0), "Failed to delete created item")
}

func TestCrud(t *testing.T) {
	suite.Run(t, new(DbTestSuite))
}
