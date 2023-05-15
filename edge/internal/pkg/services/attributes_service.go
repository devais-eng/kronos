package services

import (
	"devais.it/kronos/internal/pkg/constants"
	"devais.it/kronos/internal/pkg/db"
	"devais.it/kronos/internal/pkg/db/models"
	"devais.it/kronos/internal/pkg/types"
	"devais.it/kronos/internal/pkg/util"
	"github.com/google/uuid"
	"github.com/rotisserie/eris"
	"gopkg.in/guregu/null.v4"
	"gorm.io/gorm"
)

func BatchCreateAttributesTx(ctx *db.TxContext, attributes []models.Attribute, modifiedBy string) error {
	if len(attributes) == 0 {
		return ErrEmptySlice
	}

	var err error

	tx := ctx.Tx

	for _, attribute := range attributes {
		if attribute.CreatedBy == "" {
			attribute.CreatedBy = modifiedBy
		}

		if attribute.ModifiedBy == "" {
			attribute.ModifiedBy = modifiedBy
		}

		err = tx.Create(&attribute).Error
		if err != nil {
			return err
		}

		err = PublishEvent(
			ctx,
			types.EventEntityCreated,
			types.EntityTypeAttribute,
			attribute.ID,
			modifiedBy,
			attribute,
		)
		if err != nil {
			return err
		}
	}

	return nil
}

func BatchCreateAttributes(attributes []models.Attribute, modifiedBy string) error {
	err := db.DB().Transaction(func(tx *gorm.DB) error {
		ctx := &db.TxContext{Tx: tx}
		txLen := db.CalcTxLength(attributes)

		if txLen > 1 {
			ctx.TxUUID = uuid.NewString()
			ctx.TxLen += txLen
		}

		return BatchCreateAttributesTx(ctx, attributes, modifiedBy)
	})

	if err != nil {
		if len(attributes) == 1 {
			return eris.Wrapf(err, "failed to create attribute '%s'", attributes[0].ID)
		}

		return eris.Wrap(err, "failed to create attributes")
	}

	return nil
}

func CreateAttribute(attribute *models.Attribute, modifiedBy string) error {
	return BatchCreateAttributes([]models.Attribute{*attribute}, modifiedBy)
}

func GetAttributeByID(attrID string) (*models.Attribute, error) {
	attr := &models.Attribute{}
	err := db.GetByID(attrID, attr)
	if err != nil {
		return nil, eris.Wrapf(err, "failed to get attribute '%s'", attrID)
	}
	return attr, nil
}

func GetAttributesByIDs(ids []string) ([]models.Attribute, error) {
	var attributes []models.Attribute
	err := GetByIDs(ids, &attributes)
	if err != nil {
		return nil, eris.Wrapf(err, "failed to get attributes")
	}
	return attributes, nil
}

func GetAttributesByType(attributeType string, page, pageSize int) ([]models.Attribute, error) {
	var attributes []models.Attribute
	err := GetByType(attributeType, &attributes, page, pageSize)
	if err != nil {
		return nil, err
	}
	return attributes, nil
}

func FindAttributesByName(name string, page, pageSize int) ([]models.Attribute, error) {
	var attributes []models.Attribute
	err := FindByName(name, &attributes, page, pageSize)
	if err != nil {
		return nil, err
	}
	return attributes, nil
}

func GetAllAttributes(page, pageSize int) (attributes []models.Attribute, err error) {
	err = db.GetAll(&attributes, page, pageSize)
	if err != nil {
		err = eris.Wrap(err, "failed to get all attributes")
	}
	return
}

func UpdateAttributeTx(ctx *db.TxContext, patch map[string]interface{}, modifiedBy string) error {
	err := UpdateTx(ctx, &models.Attribute{}, patch, modifiedBy)
	if err != nil {
		return err
	}

	return nil
}

func UpdateAttribute(patch map[string]interface{}, modifiedBy string) error {
	err := db.DB().Transaction(func(tx *gorm.DB) error {
		return UpdateAttributeTx(&db.TxContext{Tx: tx}, patch, modifiedBy)
	})

	if err != nil {
		return eris.Wrapf(err, "failed to update attribute")
	}

	return nil
}

func UpsertAttributeTx(ctx *db.TxContext, patch map[string]interface{}, modifiedBy string) error {
	err := UpdateAttributeTx(ctx, patch, modifiedBy)
	if eris.Is(err, gorm.ErrRecordNotFound) {
		attr := models.Attribute{}
		err = util.JSONToStruct(patch, &attr)
		if err != nil {
			return eris.Wrapf(err, "failed to unmarshal attribute")
		}
		err = BatchCreateAttributesTx(ctx, []models.Attribute{attr}, modifiedBy)
	}

	return err
}

func UpsertAttribute(patch map[string]interface{}, modifiedBy string) error {
	err := db.DB().Transaction(func(tx *gorm.DB) error {
		return UpsertAttributeTx(&db.TxContext{Tx: tx}, patch, modifiedBy)
	})

	if err != nil {
		return eris.Wrapf(err, "failed to upsert attribute")
	}

	return nil
}

func DeleteAttribute(attr *models.Attribute, modifiedBy string) error {
	err := Delete(attr, modifiedBy)

	if err != nil {
		return eris.Wrapf(err, "failed to delete attribute '%s'", attr.ID)
	}

	return nil
}

func HardDeleteAttributeByID(attributeID, modifiedBy string) error {
	err := HardDeleteByID(attributeID, &models.Attribute{}, modifiedBy)

	if err != nil {
		return eris.Wrapf(err, "failed to hard delete attribute '%s'", attributeID)
	}

	return nil
}

func DeleteAttributeByIDTx(ctx *db.TxContext, attributeID, modifiedBy string) error {
	return DeleteByIDTx(ctx, attributeID, &models.Attribute{}, modifiedBy)
}

func DeleteAttributeByID(attributeID, modifiedBy string) error {
	err := DeleteByID(attributeID, &models.Attribute{}, modifiedBy)

	if err != nil {
		return eris.Wrapf(err, "failed to delete attribute '%s'", attributeID)
	}

	return nil
}

func GetAttributesCount() (count int64, err error) {
	count, err = db.Count(&models.Attribute{})
	if err != nil {
		err = eris.Wrap(err, "failed to get attributes count")
	}
	return
}

func GetAttributeVersion(attributeID string) (string, error) {
	version, err := GetVersion(attributeID, models.AttributesTableName)
	if err != nil {
		return "", eris.Wrapf(err, "failed to get version of attribute '%s'", attributeID)
	}
	return version, nil
}

func GetAttributesVersion(page, pageSize int) ([]models.EntityVersion, error) {
	versions, err := GetAllVersions(models.AttributesTableName, page, pageSize)
	if err != nil {
		return nil, eris.Wrap(err, "failed to get attributes version")
	}
	return versions, nil
}

func GetAttributesSyncPolicy(attributeID string) (null.String, error) {
	attribute := &models.Attribute{}
	err := db.DB().
		Select("sync_policy").
		First(attribute, constants.IDField+" = ?", attributeID).
		Error
	if err != nil {
		return null.String{}, eris.Wrapf(err, "failed to get sync policy of attribute '%s'", attributeID)
	}
	return attribute.SyncPolicy, nil
}

func GetAttributeValue(attributeID string) (*models.AttributeValue, error) {
	value := &models.AttributeValue{}
	err := db.DB().
		Model(&models.Attribute{}).
		Select("value, value_type").
		First(value, constants.IDField+" = ?", attributeID).
		Error
	if err != nil {
		return nil, eris.Wrapf(err, "failed to get value of attribute '%s'", attributeID)
	}
	return value, err
}
