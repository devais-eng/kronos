package services

import (
	"testing"

	"devais.it/kronos/internal/pkg/db"
	"devais.it/kronos/internal/pkg/db/models"
	"devais.it/kronos/internal/pkg/util"
	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"
)

const (
	mbRelation         = "RELATIONS_TEST"
	relationsBatchSize = 50
)

type RelationsSuite struct {
	db.SuiteBase
}

func newRelation(parentID, childID string) *models.Relation {
	return &models.Relation{ParentID: parentID, ChildID: childID}
}

func (s *RelationsSuite) assertCount(expected int) {
	assert := s.Require()
	count, err := GetRelationsCount()
	assert.NoError(err)
	assert.Equal(int64(expected), count)
}

func (s *RelationsSuite) assertContains(relations []models.Relation, parentID, childID string) {
	assert := s.Require()
	contains := false
	for _, relation := range relations {
		if relation.ParentID == parentID && relation.ChildID == childID {
			contains = true
			break
		}
	}
	assert.True(contains, "List doesn't contain relation '%s->%s'", parentID, childID)
}

func (s *RelationsSuite) assertContainsItem(items []models.Item, id string) {
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

func (s *RelationsSuite) TestGet() {
	assert := s.Require()

	s.assertCount(0)

	parent := newItem()
	child := newItem()
	rel := newRelation(parent.ID, child.ID)

	_, err := GetRelation(parent.ID, child.ID)
	assert.ErrorIs(err, gorm.ErrRecordNotFound)

	assert.NoError(CreateItem(parent, mbRelation))
	assert.NoError(CreateItem(child, mbRelation))

	assert.NoError(CreateRelation(rel, mbRelation))

	s.assertCount(1)

	createdRel, err := GetRelation(parent.ID, child.ID)
	assert.NoError(err)
	assert.NotNil(createdRel)
	assert.Equal(parent.ID, createdRel.ParentID)
	assert.Equal(child.ID, createdRel.ChildID)

	allRelations, err := GetAllRelations(-1, -1)
	assert.Error(err)
	assert.Empty(allRelations)

	allRelations, err = GetAllRelations(1, 1)
	assert.NoError(err)
	s.assertContains(allRelations, parent.ID, child.ID)

	nonexistentItemID := "ThisItemDoesNotExist"

	children, err := GetItemChildren(nonexistentItemID)
	assert.ErrorIs(err, gorm.ErrRecordNotFound)
	assert.Empty(children)

	parents, err := GetItemParents(nonexistentItemID)
	assert.ErrorIs(err, gorm.ErrRecordNotFound)
	assert.Empty(parents)

	itemRelations, err := GetItemRelations(nonexistentItemID)
	assert.ErrorIs(err, gorm.ErrRecordNotFound)
	assert.Empty(itemRelations)
}

func (s *RelationsSuite) TestCreate() {
	assert := s.Require()

	parent := newItem()
	children := make([]models.Item, relationsBatchSize)
	childrenIds := util.NewSet()

	for i := 0; i < relationsBatchSize; i++ {
		nItem := newItem()
		childrenIds.Add(nItem.ID)
		children[i] = *nItem
	}

	assert.NoError(CreateItem(parent, mbRelation))

	for _, child := range children {
		assert.NoError(CreateItem(&child, mbRelation))
	}

	assert.Error(CreateRelation(&models.Relation{
		ParentID: parent.ID,
		ChildID:  uuid.NewString(),
	}, mbRelation))

	start := util.TimestampMs()
	for _, child := range children {
		relation := &models.Relation{
			ParentID: parent.ID,
			ChildID:  child.ID,
		}
		assert.NoError(CreateRelation(relation, mbRelation))
	}
	end := util.TimestampMs()

	s.assertCount(relationsBatchSize)

	allRelations, err := GetAllRelations(0, relationsBatchSize)
	assert.NoError(err)
	assert.Len(allRelations, relationsBatchSize)

	for _, relation := range allRelations {
		assert.GreaterOrEqual(relation.CreatedAt, start)
		assert.GreaterOrEqual(relation.ModifiedAt, start)
		assert.LessOrEqual(relation.CreatedAt, end)
		assert.LessOrEqual(relation.ModifiedAt, end)
	}

	resChildren, err := GetItemChildren(parent.ID)
	assert.NoError(err)
	assert.Len(resChildren, relationsBatchSize)

	for _, child := range resChildren {
		assert.True(childrenIds.Has(child.ID))
		parents, err := GetItemParents(child.ID)
		assert.NoError(err)
		assert.Len(parents, 1)
		assert.Equal(parent.ID, parents[0].ID)
	}

	assert.NoError(DeleteItem(parent, mbRelation))
	s.assertCount(0)
}

func (s *RelationsSuite) TestBatchCreate() {
	assert := s.Require()

	s.assertCount(0)

	err := BatchCreateRelations([]models.Relation{}, mbRelation)
	assert.ErrorIs(err, ErrEmptySlice)

	err = BatchCreateRelations(nil, mbRelation)
	assert.ErrorIs(err, ErrEmptySlice)

	items := make([]models.Item, relationsBatchSize*2)
	relations := make([]models.Relation, relationsBatchSize)

	for i := 0; i < relationsBatchSize; i++ {
		parent := newItem()
		child := newItem()

		relation := newRelation(parent.ID, child.ID)

		items[i*2+0] = *parent
		items[i*2+1] = *child
		relations[i] = *relation
	}

	assert.NoError(BatchCreateItems(items, mbRelation))

	// Create all relations
	start := util.TimestampMs()
	assert.NoError(BatchCreateRelations(relations, mbRelation))
	end := util.TimestampMs()

	allRelations, err := GetAllRelations(0, relationsBatchSize)
	assert.NoError(err)
	assert.Len(allRelations, relationsBatchSize)

	for _, relation := range relations {
		s.assertContains(relations, relation.ParentID, relation.ChildID)

		createdRelation, err := GetRelation(relation.ParentID, relation.ChildID)
		assert.NoError(err)
		assert.GreaterOrEqual(createdRelation.CreatedAt, start)
		assert.GreaterOrEqual(createdRelation.ModifiedAt, start)
		assert.LessOrEqual(createdRelation.CreatedAt, end)
		assert.LessOrEqual(createdRelation.ModifiedAt, end)
	}
}

func (s *RelationsSuite) TestMove() {
	assert := s.Require()

	itemA := newItem()
	itemB := newItem()
	itemC := newItem()

	items := []models.Item{
		*itemA, *itemB, *itemC,
	}

	assert.NoError(BatchCreateItems(items, mbRelation))

	assert.NoError(CreateRelation(&models.Relation{ParentID: itemA.ID, ChildID: itemB.ID}, mbRelation))

	children, err := GetItemChildren(itemA.ID)
	assert.NoError(err)
	assert.Len(children, 1)
	s.assertContainsItem(children, itemB.ID)

	parents, err := GetItemParents(itemB.ID)
	assert.NoError(err)
	assert.Len(parents, 1)
	s.assertContainsItem(parents, itemA.ID)

	// Test move error
	assert.ErrorIs(MoveItem(itemA.ID, uuid.NewString(), itemC.ID, mbRelation), gorm.ErrRecordNotFound)
	assert.Error(MoveItem(itemA.ID, itemB.ID, uuid.NewString(), mbRelation))

	// Move B from A to C
	assert.NoError(MoveItem(itemA.ID, itemB.ID, itemC.ID, mbRelation))

	children, err = GetItemChildren(itemA.ID)
	assert.NoError(err)
	assert.Empty(children)

	children, err = GetItemChildren(itemC.ID)
	assert.NoError(err)
	assert.Len(children, 1)
	s.assertContainsItem(children, itemB.ID)

	parents, err = GetItemParents(itemB.ID)
	assert.NoError(err)
	assert.Len(parents, 1)
	s.assertContainsItem(parents, itemC.ID)

	relation, err := GetRelation(itemC.ID, itemB.ID)
	assert.NoError(err)
	assert.Equal(mbRelation, relation.CreatedBy)
	assert.Equal(mbRelation, relation.ModifiedBy)
}

func (s *RelationsSuite) TestDelete() {
	assert := s.Require()

	s.assertCount(0)

	parent := newItem()
	child := newItem()

	assert.ErrorIs(DeleteRelation("", "", mbRelation), db.ErrMissingID)
	assert.ErrorIs(HardDeleteRelation("", "", mbRelation), db.ErrMissingID)

	assert.ErrorIs(DeleteRelation(parent.ID, child.ID, mbRelation), gorm.ErrRecordNotFound)
	assert.ErrorIs(HardDeleteRelation(parent.ID, child.ID, mbRelation), gorm.ErrRecordNotFound)

	assert.NoError(CreateItem(parent, mbRelation))
	assert.NoError(CreateItem(child, mbRelation))

	relation := &models.Relation{ParentID: parent.ID, ChildID: child.ID}

	// Test delete
	assert.NoError(CreateRelation(relation, mbRelation))
	s.assertCount(1)

	assert.NoError(DeleteRelation(parent.ID, child.ID, mbRelation))
	s.assertCount(0)

	// Test hard delete
	assert.NoError(CreateRelation(relation, mbRelation))
	s.assertCount(1)

	assert.NoError(HardDeleteRelation(parent.ID, child.ID, mbRelation))
	s.assertCount(0)

	// Test delete parent cascade
	assert.NoError(CreateRelation(relation, mbRelation))
	s.assertCount(1)

	assert.NoError(DeleteItem(parent, mbRelation))
	s.assertCount(0)

	// Test delete child cascade
	assert.NoError(CreateItem(parent, mbRelation))
	assert.NoError(CreateRelation(relation, mbRelation))
	s.assertCount(1)
	assert.NoError(DeleteItem(child, mbRelation))
	s.assertCount(0)
}

func TestRelationsService(t *testing.T) {
	suite.Run(t, new(RelationsSuite))
}
