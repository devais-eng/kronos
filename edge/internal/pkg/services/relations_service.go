package services

import (
	"devais.it/kronos/internal/pkg/db"
	"devais.it/kronos/internal/pkg/db/models"
	"devais.it/kronos/internal/pkg/types"
	"github.com/google/uuid"
	"github.com/rotisserie/eris"
	"gopkg.in/guregu/null.v4"
	"gorm.io/gorm"
)

func BatchCreateRelationsTx(ctx *db.TxContext, relations []models.Relation, modifiedBy string) error {
	if len(relations) == 0 {
		return ErrEmptySlice
	}

	tx := ctx.Tx

	for _, relation := range relations {
		if relation.CreatedBy == "" {
			relation.CreatedBy = modifiedBy
		}
		if relation.ModifiedBy == "" {
			relation.ModifiedBy = modifiedBy
		}
		err := tx.Create(&relation).Error
		if err != nil {
			return err
		}
		err = PublishEvent(
			ctx,
			types.EventEntityCreated,
			types.EntityTypeRelation,
			relation.CompositeID(),
			modifiedBy,
			relation,
		)
		if err != nil {
			return err
		}
	}

	return nil
}

func BatchCreateRelations(relations []models.Relation, modifiedBy string) error {
	err := db.DB().Transaction(func(tx *gorm.DB) error {
		ctx := &db.TxContext{Tx: tx}

		if len(relations) > 1 {
			ctx.TxUUID = uuid.NewString()
			ctx.TxLen = len(relations)
		}

		return BatchCreateRelationsTx(ctx, relations, modifiedBy)
	})

	if err != nil {
		if len(relations) == 1 {
			return eris.Wrapf(err, "failed to create relation '%s'", relations[0].CompositeID())
		}

		return eris.Wrap(err, "failed to create relations")
	}

	return nil
}

func CreateRelation(relation *models.Relation, modifiedBy string) error {
	return BatchCreateRelations([]models.Relation{*relation}, modifiedBy)
}

func GetAllRelations(page, pageSize int) ([]models.Relation, error) {
	var relations []models.Relation

	err := db.GetAll(&relations, page, pageSize)
	if err != nil {
		return nil, eris.Wrapf(err, "failed to get all relations")
	}

	return relations, nil
}

func GetRelation(parentID, childID string) (*models.Relation, error) {
	relation := &models.Relation{}

	tx := db.DB().
		Where("parent_id = ? AND child_id = ?", parentID, childID).
		First(relation)
	if tx.Error != nil {
		return nil, eris.Wrapf(
			tx.Error,
			"failed to get relation between parent '%s' and child '%s'",
			parentID,
			childID,
		)
	}

	return relation, nil
}

func HardDeleteRelation(parentID, childID, modifiedBy string) error {
	if parentID == "" || childID == "" {
		return db.ErrMissingID
	}
	rel := &models.Relation{ParentID: parentID, ChildID: childID}
	err := db.GetHardDeleteTx(db.DB()).Transaction(func(tx *gorm.DB) error {
		ctx := &db.TxContext{Tx: tx}
		tx = tx.
			Where("parent_id = ? AND child_id = ?", parentID, childID).
			Delete(&models.Relation{})
		if tx.Error != nil {
			return tx.Error
		}
		if tx.RowsAffected == 0 {
			return gorm.ErrRecordNotFound
		}
		return PublishEvent(
			ctx,
			types.EventEntityDeleted,
			types.EntityTypeRelation,
			rel.CompositeID(),
			modifiedBy,
			rel,
		)
	})

	if err != nil {
		return eris.Wrapf(
			err,
			"failed to delete relation with parent '%s' and child '%s'",
			parentID,
			childID,
		)
	}

	return nil
}

func DeleteRelationTx(ctx *db.TxContext, parentID, childID, modifiedBy string) error {
	if parentID == "" || childID == "" {
		return db.ErrMissingID
	}
	tx := ctx.Tx.
		Where("parent_id = ? AND child_id = ?", parentID, childID).
		Delete(&models.Relation{})
	if tx.Error != nil {
		return tx.Error
	}
	if tx.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	rel := &models.Relation{ParentID: parentID, ChildID: childID}
	return PublishEvent(
		ctx,
		types.EventEntityDeleted,
		types.EntityTypeRelation,
		rel.CompositeID(),
		modifiedBy,
		rel,
	)
}

func DeleteRelation(parentID, childID, modifiedBy string) error {
	err := db.DB().Transaction(func(tx *gorm.DB) error {
		return DeleteRelationTx(&db.TxContext{Tx: tx}, parentID, childID, modifiedBy)
	})

	if err != nil {
		return eris.Wrapf(
			err,
			"failed to delete relation with parent '%s' and child '%s'",
			parentID,
			childID,
		)
	}

	return nil
}

func MoveItem(parentID, childID, newParentID, modifiedBy string) error {
	tx := db.DB().Exec(
		"UPDATE "+models.RelationsTableName+" "+
			"SET parent_id = ?, modified_by = ? "+
			"WHERE parent_id = ? AND child_id = ?",
		newParentID,
		modifiedBy,
		parentID,
		childID,
	)
	if tx.Error != nil {
		return eris.Wrapf(tx.Error, "failed to move item from '%s' to '%s'", parentID, newParentID)
	}
	if tx.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}

	return nil
}

func GetRelationsCount() (count int64, err error) {
	count, err = db.Count(&models.Relation{})
	if err != nil {
		err = eris.Wrap(err, "failed to get relations count")
	}
	return
}

func GetRelationSyncPolicy(parentID, childID string) (null.String, error) {
	relation := &models.Relation{ParentID: parentID, ChildID: childID}
	tx := db.DB().
		Model(relation).
		Select("sync_policy").
		Where("parent_id = ? AND child_id = ?", parentID, childID).
		First(relation)
	if tx.Error != nil {
		return null.String{}, eris.Wrapf(
			tx.Error,
			"failed to get sync policy of relation '%s'",
			relation.CompositeID(),
		)
	}
	return relation.SyncPolicy, nil
}
