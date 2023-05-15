package services

import (
	"devais.it/kronos/internal/pkg/constants"
	"devais.it/kronos/internal/pkg/db"
	"devais.it/kronos/internal/pkg/db/models"
	"devais.it/kronos/internal/pkg/types"
	"github.com/google/uuid"
	jsoniter "github.com/json-iterator/go"
	"github.com/rotisserie/eris"
	"gorm.io/gorm"
	"strings"
)

var (
	json = jsoniter.ConfigCompatibleWithStandardLibrary

	ErrEmptySlice = eris.New("an empty slice was given")
)

//=============================================================================
// Common functions
//=============================================================================

// lcEntityType returns the lower case entity type of a given model.
// Used to add context to error traces
func lcEntityType(entity interface{}) string {
	return strings.ToLower(string(models.GetEntityType(entity)))
}

// GetByID gets an entity by its ID
func GetByID(id string, model interface{}) error {
	err := db.GetByID(id, model)
	if err != nil {
		return eris.Wrapf(
			err,
			"failed to get %s '%s'",
			lcEntityType(model),
			id,
		)
	}
	return nil
}

func GetByName(name string, model interface{}) error {
	err := db.DB().
		Where("name = ?", name).
		First(model).
		Error
	if err != nil {
		return eris.Wrapf(
			err,
			"failed to get %s '%s'",
			lcEntityType(model),
			name,
		)
	}
	return nil
}

func GetByType(typeStr string, model interface{}, page, pageSize int) error {
	tx, err := db.Paginate(db.DB(), page, pageSize)
	if err != nil {
		return err
	}
	err = tx.
		Find(model, "type = ?", typeStr).
		Error
	if err != nil {
		return eris.Wrapf(
			err,
			"failed to get %ss by type %s",
			lcEntityType(model),
			typeStr,
		)
	}
	if tx.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func findByQuery(value, field string, model interface{}, page, pageSize int) error {
	tx, err := db.Paginate(db.DB(), page, pageSize)
	if err != nil {
		return err
	}
	err = tx.
		Find(model, field+" LIKE ?", "%"+value+"%").
		Error
	if err != nil {
		return eris.Wrapf(
			err,
			"failed to find %ss with %s %s",
			lcEntityType(model),
			field,
			value,
		)
	}
	return nil
}

func FindByName(name string, model interface{}, page, pageSize int) error {
	return findByQuery(
		name,
		"name",
		model,
		page,
		pageSize,
	)
}

func FindByType(typeStr string, model interface{}, page, pageSize int) error {
	return findByQuery(
		typeStr,
		"type",
		model,
		page,
		pageSize,
	)
}

func GetByIDs(ids []string, models interface{}) error {
	if len(ids) == 0 {
		return db.ErrMissingID
	}
	tx := db.DB().
		Model(models).
		Where("id IN ?", ids).
		Find(models)
	if tx.Error != nil {
		return tx.Error
	}
	if tx.RowsAffected != int64(len(ids)) {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func UpdateTx(ctx *db.TxContext, model interface{}, patch map[string]interface{}, modifiedBy string) error {
	if m := patch[constants.ModifiedByField]; m == nil || m == "" {
		patch[constants.ModifiedByField] = modifiedBy
	}

	err := db.Update(ctx.Tx, model, patch)
	if err != nil {
		return err
	}

	id := patch["id"].(string)

	// Fetch version
	var version string

	err = ctx.Tx.
		Model(model).
		Select("version").
		First(&version, "id = ?", id).
		Error
	if err != nil {
		return err
	}

	// Add version to patch
	patch["version"] = version

	return PublishEvent(
		ctx,
		types.EventEntityUpdated,
		models.GetEntityType(model),
		id,
		modifiedBy,
		patch,
	)
}

func Update(model interface{}, patch map[string]interface{}, modifiedBy string) error {
	return db.DB().Transaction(func(tx *gorm.DB) error {
		return UpdateTx(&db.TxContext{Tx: tx}, model, patch, modifiedBy)
	})
}

func Delete(model interface{}, modifiedBy string) error {
	return db.DB().Transaction(func(tx *gorm.DB) error {
		err := db.Delete(tx, model)
		if err != nil {
			return err
		}

		return PublishEvent(
			&db.TxContext{Tx: tx},
			types.EventEntityDeleted,
			models.GetEntityType(model),
			models.GetEntityID(model),
			modifiedBy,
			nil,
		)
	})
}

func DeleteByIDTx(ctx *db.TxContext, id string, model interface{}, modifiedBy string) error {
	err := db.DeleteByID(ctx.Tx, id, model)
	if err != nil {
		return err
	}

	return PublishEvent(
		ctx,
		types.EventEntityDeleted,
		models.GetEntityType(model),
		id,
		modifiedBy,
		nil,
	)
}

func DeleteByID(id string, model interface{}, modifiedBy string) error {
	return db.DB().Transaction(func(tx *gorm.DB) error {
		return DeleteByIDTx(&db.TxContext{Tx: tx}, id, model, modifiedBy)
	})
}

func HardDeleteByID(id string, model interface{}, modifiedBy string) error {
	return db.GetHardDeleteTx(db.DB()).Transaction(func(tx *gorm.DB) error {
		return DeleteByIDTx(&db.TxContext{Tx: tx}, id, model, modifiedBy)
	})
}

func GetVersion(id, modelTableName string) (string, error) {
	err := db.CheckID(id)
	if err != nil {
		return "", err
	}
	var version string
	tx := db.DB().
		Raw("SELECT version FROM "+modelTableName+" WHERE id = ?", id).
		First(&version)
	if tx.Error != nil {
		return "", tx.Error
	}
	return version, nil
}

func GetAllVersions(modelTableName string, page, pageSize int) ([]models.EntityVersion, error) {
	var versions []models.EntityVersion
	tx, err := db.Paginate(db.DB(), page, pageSize)
	if err != nil {
		return nil, eris.Wrap(err, "failed to paginate versions")
	}
	err = tx.
		Raw("SELECT id, version, sync_version, modified_at, modified_by FROM " + modelTableName).
		Find(&versions).Error
	if err != nil {
		return nil, err
	}
	return versions, nil
}

//=============================================================================
// Batch functions
//=============================================================================

func BatchCreateAllTx(
	ctx *db.TxContext,
	items []models.Item,
	attributes []models.Attribute,
	relations []models.Relation,
	modifiedBy string) error {
	err := BatchCreateItemsTx(ctx, items, modifiedBy)
	if err != nil {
		return err
	}
	err = BatchCreateAttributesTx(ctx, attributes, modifiedBy)
	if err != nil {
		return err
	}
	err = BatchCreateRelationsTx(ctx, relations, modifiedBy)
	if err != nil {
		return err
	}

	return nil
}

func BatchUpdateAllTx(
	ctx *db.TxContext,
	items []map[string]interface{},
	attributes []map[string]interface{},
	relations []map[string]interface{},
	modifiedBy string,
) error {
	for _, itemPatch := range items {
		err := UpdateItemTx(ctx, itemPatch, modifiedBy)
		if err != nil {
			return err
		}
	}

	for _, attributePatch := range attributes {
		err := UpdateAttributeTx(ctx, attributePatch, modifiedBy)
		if err != nil {
			return err
		}
	}

	/*
		for _, relationPatch := range relations {
		}
	*/

	return nil
}

func BatchDeleteAllTx(
	ctx *db.TxContext,
	items []string,
	attributes []string,
	relations []string,
	modifiedBy string,
) error {
	for _, relationID := range relations {
		relation := &models.Relation{}
		err := relation.SetCompositeID(relationID)
		if err != nil {
			return err
		}
		err = DeleteRelationTx(ctx, relation.ParentID, relation.ChildID, modifiedBy)
		if err != nil {
			return err
		}
	}

	for _, attributeID := range attributes {
		err := DeleteAttributeByIDTx(ctx, attributeID, modifiedBy)
		if err != nil {
			return err
		}
	}

	for _, itemID := range items {
		err := DeleteItemByIDTx(ctx, itemID, modifiedBy)
		if err != nil {
			return err
		}
	}

	return nil
}

func BatchCreateAll(
	items []models.Item,
	attributes []models.Attribute,
	relations []models.Relation,
	modifiedBy string) error {
	return db.DB().Transaction(func(tx *gorm.DB) error {
		ctx := &db.TxContext{
			Tx:     tx,
			TxUUID: uuid.NewString(),
			TxLen:  len(items) + len(attributes) + len(relations),
		}
		return BatchCreateAllTx(
			ctx,
			items,
			attributes,
			relations,
			modifiedBy,
		)
	})
}

func BatchDeleteAll(
	items []string,
	attributes []string,
	relations []string,
	modifiedBy string) error {
	return db.DB().Transaction(func(tx *gorm.DB) error {
		ctx := &db.TxContext{
			Tx:     tx,
			TxUUID: uuid.NewString(),
			TxLen:  len(items) + len(attributes) + len(relations),
		}
		return BatchDeleteAllTx(ctx, items, attributes, relations, modifiedBy)
	})
}
