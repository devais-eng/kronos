package services

import (
	"devais.it/kronos/internal/pkg/db"
	"devais.it/kronos/internal/pkg/db/models"
	"devais.it/kronos/internal/pkg/types"
	"devais.it/kronos/internal/pkg/util"
	"github.com/google/uuid"
	"github.com/rotisserie/eris"
	"gopkg.in/guregu/null.v4"
	"gorm.io/gorm"
	"reflect"
)

func CreateItem(item *models.Item, modifiedBy string) error {
	return BatchCreateItems([]models.Item{*item}, modifiedBy)
}

func BatchCreateItemsTx(ctx *db.TxContext, items []models.Item, modifiedBy string) error {
	if len(items) == 0 {
		return ErrEmptySlice
	}

	var err error

	tx := ctx.Tx

	for _, item := range items {
		if item.CreatedBy == "" {
			item.CreatedBy = modifiedBy
		}

		if item.ModifiedBy == "" {
			item.ModifiedBy = modifiedBy
		}

		err = tx.Create(&item).Error
		if err != nil {
			return err
		}

		// Copy item to avoid serialization of attributes inside event body
		itemBody := item
		itemBody.Attributes = nil

		err = PublishEvent(
			ctx,
			types.EventEntityCreated,
			types.EntityTypeItem,
			item.ID,
			modifiedBy,
			itemBody,
		)
		if err != nil {
			return err
		}

		// Create attributes if present
		for _, attribute := range item.Attributes {
			if attribute.ItemID == "" {
				attribute.ItemID = item.ID
			} else if attribute.ItemID != item.ID {
				return eris.Errorf(
					"Can't create a attribute on item '%s' during creation of item '%s'",
					attribute.ItemID,
					item.ID,
				)
			}

			err = BatchCreateAttributesTx(ctx, []models.Attribute{attribute}, modifiedBy)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func BatchCreateItems(items []models.Item, modifiedBy string) error {
	err := db.DB().Transaction(func(tx *gorm.DB) error {
		ctx := &db.TxContext{Tx: tx}
		txLen := db.CalcTxLength(items)

		if txLen > 1 {
			ctx.TxUUID = uuid.NewString()
			ctx.TxLen += txLen
		}

		return BatchCreateItemsTx(ctx, items, modifiedBy)
	})

	if err != nil {
		if len(items) == 1 {
			return eris.Wrapf(err, "failed to create item '%s'", items[0].ID)
		}
		return eris.Wrap(err, "failed to create items")
	}

	return nil
}

func ItemExists(itemID string) (bool, error) {
	exists := false
	err := db.
		DB().
		Model(&models.Item{}).
		Select("1").
		First(&exists, "id = ?", itemID).
		Error
	if err != nil {
		return false, err
	}
	return exists, nil
}

func ItemExistsErr(itemID string) error {
	exists, err := ItemExists(itemID)
	if err != nil {
		return err
	}
	if !exists {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func GetItemByID(itemID string) (*models.Item, error) {
	item := &models.Item{}
	err := db.GetByID(itemID, item)
	if err != nil {
		return nil, eris.Wrapf(err, "failed to get item '%s'", itemID)
	}
	return item, nil
}

func GetItemsByIDs(ids []string) ([]models.Item, error) {
	var items []models.Item
	err := GetByIDs(ids, &items)
	if err != nil {
		return nil, eris.Wrap(err, "failed to get items")
	}
	return items, nil
}

func GetItemByName(itemName string) (*models.Item, error) {
	item := &models.Item{}
	err := GetByName(itemName, item)
	if err != nil {
		return nil, eris.Wrapf(err, "failed to get item '%s'", itemName)
	}
	return item, nil
}

func GetItemsByType(itemType string, page, pageSize int) ([]models.Item, error) {
	var items []models.Item
	err := GetByType(itemType, &items, page, pageSize)
	if err != nil {
		return nil, eris.Wrapf(err, "failed to get items by type '%s'", itemType)
	}
	return items, nil
}

func FindItemsByName(name string, page, pageSize int) ([]models.Item, error) {
	var items []models.Item
	err := FindByName(name, &items, page, pageSize)
	if err != nil {
		return nil, err
	}
	return items, nil
}

func FindItemsByType(typeStr string, page, pageSize int) ([]models.Item, error) {
	var items []models.Item
	err := FindByType(typeStr, &items, page, pageSize)
	if err != nil {
		return nil, err
	}
	return items, nil
}

func GetItemVersion(itemID string) (string, error) {
	version, err := GetVersion(itemID, models.ItemsTableName)
	if err != nil {
		return "", eris.Wrapf(err, "failed to get version of item '%s'", itemID)
	}
	return version, nil
}

func GetItemMac(itemID string) (null.String, error) {
	item := &models.Item{}

	tx := db.DB().
		Select("edge_mac").
		First(item, "id = ?", itemID)
	if tx.Error != nil {
		return null.String{}, tx.Error
	}

	return item.EdgeMac, nil
}

func GetItemLastUpdateTime(itemID string) (uint64, error) {
	var updateTime uint64
	err := db.DB().
		Model(&models.Item{}).
		Select("modified_at").
		Where("id = ?", itemID).
		First(&updateTime).
		Error

	if err != nil {
		return 0, eris.Wrapf(err, "failed to get item update time '%s'", itemID)
	}
	return updateTime, nil
}

func GetItemCustomerID(itemID string) (null.String, error) {
	item := &models.Item{}
	tx := db.DB().
		Select("customer_id").
		First(item, "id = ?", itemID)
	if tx.Error != nil {
		return null.String{}, eris.Wrapf(tx.Error, "failed to get item customer id '%s'", itemID)
	}
	return item.CustomerID, nil
}

func GetItemCreatedBy(itemID string) (string, error) {
	var createdBy string
	err := db.DB().
		Model(&models.Item{}).
		Select("created_by").
		Where("id = ?", itemID).
		First(&createdBy).
		Error
	if err != nil {
		return "", eris.Wrapf(err, "failed to get item created by '%s'", itemID)
	}
	return createdBy, nil
}

func GetItemModifiedBy(itemID string) (string, error) {
	var modifiedBy string
	err := db.DB().
		Model(&models.Item{}).
		Select("modified_by").
		Where("id = ?", itemID).
		First(&modifiedBy).
		Error
	if err != nil {
		return "", eris.Wrapf(err, "failed to get item modified by '%s'", itemID)
	}

	return modifiedBy, nil
}

func GetAllItems(page, pageSize int) (items []models.Item, err error) {
	err = db.GetAll(&items, page, pageSize)
	if err != nil {
		err = eris.Wrap(err, "failed to get all items")
	}
	return
}

func UpdateItemTx(ctx *db.TxContext, patch map[string]interface{}, modifiedBy string) error {
	itemPatch := make(map[string]interface{})
	for key, value := range patch {
		itemPatch[key] = value
	}

	var attributes interface{}

	if attrs, ok := itemPatch["attributes"]; ok {
		// Extract attributes from item patch
		attributes = attrs
		delete(itemPatch, "attributes")

		// If transaction length is 0, set it to at least to 1
		if ctx.TxLen == 0 {
			ctx.TxLen = 1
		}
	}

	err := UpdateTx(ctx, &models.Item{}, itemPatch, modifiedBy)
	if err != nil {
		return err
	}

	if attributes != nil {
		itemId := itemPatch["id"]

		// Update all attributes, increasing transaction length accordingly
		switch v := attributes.(type) {
		case []map[string]interface{}:
			ctx.TxLen += len(v)

			for _, attrPatch := range v {
				attrPatch["item_id"] = itemId
				err = UpsertAttributeTx(ctx, attrPatch, modifiedBy)
				if err != nil {
					return err
				}
			}
		case []interface{}:
			ctx.TxLen += len(v)

			for _, attr := range v {
				attrPatch, ok := attr.(map[string]interface{})
				if !ok {
					return eris.New("attribute patches must be JSON objects")
				}
				attrPatch["item_id"] = itemId
				err = UpsertAttributeTx(ctx, attrPatch, modifiedBy)
				if err != nil {
					return err
				}
			}
		default:
			return eris.Errorf(
				"attributes field must be a list, go %v instead",
				reflect.TypeOf(attributes),
			)
		}
	}

	return nil
}

func UpdateItem(patch map[string]interface{}, modifiedBy string) error {
	err := db.DB().Transaction(func(tx *gorm.DB) error {
		return UpdateItemTx(&db.TxContext{Tx: tx}, patch, modifiedBy)
	})

	if err != nil {
		return eris.Wrapf(err, "failed to update item")
	}

	return nil
}

func UpsertItemTx(ctx *db.TxContext, patch map[string]interface{}, modifiedBy string) error {
	err := UpdateItemTx(ctx, patch, modifiedBy)
	if eris.Is(err, gorm.ErrRecordNotFound) {
		item := models.Item{}
		err = util.JSONToStruct(patch, &item)
		if err != nil {
			return eris.Wrapf(err, "failed to unmarshal item")
		}
		err = BatchCreateItemsTx(ctx, []models.Item{item}, modifiedBy)
	}

	return err
}

func UpsertItem(patch map[string]interface{}, modifiedBy string) error {
	err := db.DB().Transaction(func(tx *gorm.DB) error {
		return UpsertItemTx(&db.TxContext{Tx: tx}, patch, modifiedBy)
	})

	if err != nil {
		return eris.Wrapf(err, "failed to upsert item")
	}

	return nil
}

func DeleteItemByIDTx(ctx *db.TxContext, itemID, modifiedBy string) error {
	err := DeleteByIDTx(ctx, itemID, &models.Item{}, modifiedBy)
	if err != nil {
		return eris.Wrapf(err, "failed to delete item '%s'", itemID)
	}
	return nil
}

func DeleteItem(item *models.Item, modifiedBy string) error {
	err := Delete(item, modifiedBy)

	if err != nil {
		return eris.Wrapf(err, "failed to delete item '%s'", item.ID)
	}

	return nil
}

func HardDeleteItemByID(itemID, modifiedBy string) error {
	err := HardDeleteByID(itemID, &models.Item{}, modifiedBy)

	if err != nil {
		return eris.Wrapf(err, "failed to hard delete item '%s'", itemID)
	}

	return nil
}

func DeleteItemByID(itemID, modifiedBy string) error {
	err := DeleteByID(itemID, &models.Item{}, modifiedBy)

	if err != nil {
		return eris.Wrapf(err, "failed to delete item '%s'", itemID)
	}

	return nil
}

func GetItemAttributes(itemID string) ([]models.Attribute, error) {
	if err := ItemExistsErr(itemID); err != nil {
		return nil, err
	}

	var attributes []models.Attribute

	tx := db.DB().Where("item_id = ?", itemID).Find(&attributes)
	if tx.Error != nil {
		return nil, eris.Wrapf(tx.Error, "failed to get attributes of item '%s'", itemID)
	}

	return attributes, nil
}

func GetItemAttributeIDByName(itemID, attributeName string) (string, error) {
	var id string
	err := db.DB().
		Model(&models.Attribute{}).
		Select("id").
		First(&id, "item_id = ? AND name = ?", itemID, attributeName).
		Error
	if err != nil {
		return "", eris.Wrapf(
			err,
			"failed to get ID of attribute '%s' on item '%s'",
			attributeName,
			itemID,
		)
	}
	return id, nil
}

func GetItemAttributeByName(itemID, attributeName string) (*models.Attribute, error) {
	attribute := &models.Attribute{}
	err := db.DB().
		First(attribute, "item_id = ? AND name = ?", itemID, attributeName).
		Error
	if err != nil {
		return nil, eris.Wrapf(
			err,
			"failed to get attribute with name '%s' on item '%s'",
			attributeName,
			itemID,
		)
	}
	return attribute, nil
}

func GetItemAttributeValueByName(itemID, attributeName string) (*models.AttributeValue, error) {
	value := &models.AttributeValue{}
	err := db.DB().
		Model(&models.Attribute{}).
		Select("value, value_type").
		First(value, "item_id = ? AND name = ?", itemID, attributeName).
		Error
	if err != nil {
		return nil, eris.Wrapf(
			err,
			"failed to get value of attribute '%s' on item '%s'",
			attributeName,
			itemID,
		)
	}
	return value, nil
}

func GetItemAttributesByType(itemID, attributeType string) ([]models.Attribute, error) {
	var attributes []models.Attribute
	err := db.DB().
		Find(&attributes, "item_id = ? AND type = ?", itemID, attributeType).
		Error
	if err != nil {
		return nil, eris.Wrapf(
			err,
			"failed to get attributes of item '%s' with type '%s'",
			itemID,
			attributeType,
		)
	}
	return attributes, nil
}

func GetItemChildren(itemID string) ([]models.Item, error) {
	if err := ItemExistsErr(itemID); err != nil {
		return nil, err
	}

	var items []models.Item

	tx := db.DB().Raw(
		"SELECT * FROM "+models.ItemsTableName+" "+
			"WHERE id IN "+
			"(SELECT child_id FROM "+models.RelationsTableName+" "+
			"WHERE parent_id=?)",
		itemID,
	).Find(&items)

	if tx.Error != nil {
		return nil, eris.Wrapf(tx.Error, "failed to get children of item '%s'", itemID)
	}

	return items, nil
}

func GetItemParents(itemID string) ([]models.Item, error) {
	if err := ItemExistsErr(itemID); err != nil {
		return nil, err
	}

	var items []models.Item

	tx := db.DB().Raw(
		"SELECT * FROM "+models.ItemsTableName+" "+
			"WHERE items.id IN "+
			"(SELECT parent_id FROM "+models.RelationsTableName+" "+
			"WHERE child_id=?)",
		itemID,
	).Find(&items)

	if tx.Error != nil {
		return nil, eris.Wrapf(tx.Error, "failed to get parents of item '%s'", itemID)
	}

	return items, nil
}

func GetItemRelations(itemID string) ([]models.Relation, error) {
	if err := ItemExistsErr(itemID); err != nil {
		return nil, err
	}

	var relations []models.Relation

	tx := db.DB().Raw(
		"SELECT * FROM "+models.RelationsTableName+" "+
			"WHERE parent_id = ? OR child_id = ?",
		itemID,
		itemID,
	).Find(&relations)

	if tx.Error != nil {
		return nil, eris.Wrap(tx.Error, "failed to get item relations")
	}

	return relations, nil
}

func GetItemsCount() (count int64, err error) {
	count, err = db.Count(&models.Item{})
	if err != nil {
		err = eris.Wrap(err, "failed to get items count")
	}
	return
}

func GetItemsVersion(page, pageSize int) ([]models.EntityVersion, error) {
	versions, err := GetAllVersions(models.ItemsTableName, page, pageSize)
	if err != nil {
		return nil, eris.Wrap(err, "failed to get items version")
	}
	return versions, nil
}

func GetItemSyncPolicy(itemID string) (null.String, error) {
	item := &models.Item{}
	err := db.DB().
		Select("sync_policy").
		First(item, "id = ?", itemID).
		Error
	if err != nil {
		return null.String{}, eris.Wrapf(err, "failed to get sync policy of item '%s'", itemID)
	}
	return item.SyncPolicy, nil
}
