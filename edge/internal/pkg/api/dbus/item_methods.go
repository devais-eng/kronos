package dbus

import (
	"devais.it/kronos/internal/pkg/constants"
	"devais.it/kronos/internal/pkg/db/models"
	"devais.it/kronos/internal/pkg/serialization"
	"devais.it/kronos/internal/pkg/services"
	"github.com/godbus/dbus/v5"
)

type itemMethods struct {
	methodsBase
}

func newItemMethods(
	interfaceName string,
	serializer serialization.Serializer,
	deserializer serialization.Deserializer) *itemMethods {
	return &itemMethods{
		methodsBase{
			InterfaceName: interfaceName,
			Serializer:    serializer,
			Deserializer:  deserializer,
		},
	}
}

func (m *itemMethods) Create(msg messageType) (messageType, *dbus.Error) {
	item := &models.Item{}

	if dErr := m.deserialize(msg, item); dErr != nil {
		return nilMessage, dErr
	}

	err := services.CreateItem(item, constants.ModifiedByDBusAPIName)
	if err != nil {
		return nilMessage, m.makeDbError(err)
	}

	if m.replyCreatedData() {
		item, err = services.GetItemByID(item.ID)
		if err != nil {
			return nilMessage, m.makeDbError(err)
		}
		return m.serialize(item)
	}

	return m.serialize(item.ID)
}

func (m *itemMethods) CreateBatch(msg messageType) (messageType, *dbus.Error) {
	var items []models.Item
	return m.withSerializer(msg, &items, func() (interface{}, *dbus.Error) {
		err := services.BatchCreateItems(items, constants.ModifiedByDBusAPIName)
		if err != nil {
			return nil, m.makeDbError(err)
		}
		ids := make([]string, len(items))
		for i := 0; i < len(items); i++ {
			ids[i] = items[i].ID
		}

		if m.replyCreatedData() {
			items, err = services.GetItemsByIDs(ids)
			if err != nil {
				return nil, m.makeDbError(err)
			}
			return items, nil
		}

		return ids, nil
	})
}

func (m *itemMethods) GetByID(itemID string) (messageType, *dbus.Error) {
	item, err := services.GetItemByID(itemID)
	if err != nil {
		return nilMessage, m.makeDbError(err)
	}
	return m.serialize(item)
}

func (m *itemMethods) GetByName(itemName string) (messageType, *dbus.Error) {
	item, err := services.GetItemByName(itemName)
	if err != nil {
		return nilMessage, m.makeDbError(err)
	}
	return m.serialize(item)
}

func (m *itemMethods) GetByType(itemType string, page, pageSize int) (messageType, *dbus.Error) {
	items, err := services.GetItemsByType(itemType, page, pageSize)
	if err != nil {
		return nilMessage, m.makeDbError(err)
	}
	return m.serialize(items)
}

func (m *itemMethods) FindByName(name string, page, pageSize int) (messageType, *dbus.Error) {
	items, err := services.FindItemsByName(name, page, pageSize)
	if err != nil {
		return nilMessage, m.makeDbError(err)
	}
	return m.serialize(items)
}

func (m *itemMethods) FindByType(typeStr string, page, pageSize int) (messageType, *dbus.Error) {
	items, err := services.FindItemsByType(typeStr, page, pageSize)
	if err != nil {
		return nilMessage, m.makeDbError(err)
	}
	return m.serialize(items)
}

func (m *itemMethods) GetAll(page, pageSize int) (messageType, *dbus.Error) {
	items, err := services.GetAllItems(page, pageSize)
	if err != nil {
		return nilMessage, m.makeDbError(err)
	}
	return m.serialize(items)
}

func (m *itemMethods) Update(msg messageType) (messageType, *dbus.Error) {
	var patch map[string]interface{}
	return m.withSerializer(msg, &patch, func() (interface{}, *dbus.Error) {
		err := services.UpdateItem(patch, constants.ModifiedByDBusAPIName)
		if err != nil {
			return nil, m.makeDbError(err)
		}
		item, err := services.GetItemByID(patch["id"].(string))
		if err != nil {
			return nil, m.makeDbError(err)
		}
		return item, nil
	})
}

func (m *itemMethods) DeleteByID(itemID string) (messageType, *dbus.Error) {
	err := services.DeleteItemByID(itemID, constants.ModifiedByDBusAPIName)
	if err != nil {
		return nilMessage, m.makeDbError(err)
	}
	return m.serialize(itemID)
}

func (m *itemMethods) HardDeleteByID(itemID string) (messageType, *dbus.Error) {
	err := services.HardDeleteItemByID(itemID, constants.ModifiedByDBusAPIName)
	if err != nil {
		return nilMessage, m.makeDbError(err)
	}
	return m.serialize(itemID)
}

func (m *itemMethods) Count() (int64, *dbus.Error) {
	count, err := services.GetItemsCount()
	if err != nil {
		return 0, m.makeDbError(err)
	}
	return count, nil
}

func (m *itemMethods) GetMacByID(itemID string) (string, *dbus.Error) {
	mac, err := services.GetItemMac(itemID)
	if err != nil {
		return "", m.makeDbError(err)
	}
	return mac.ValueOrZero(), nil
}

func (m *itemMethods) GetVersionByID(itemID string) (string, *dbus.Error) {
	version, err := services.GetItemVersion(itemID)
	if err != nil {
		return "", m.makeDbError(err)
	}
	return version, nil
}

func (m *itemMethods) GetCustomerIDByID(itemID string) (string, *dbus.Error) {
	customerID, err := services.GetItemCustomerID(itemID)
	if err != nil {
		return "", m.makeDbError(err)
	}
	return customerID.ValueOrZero(), nil
}

func (m *itemMethods) GetModifiedByByID(itemID string) (string, *dbus.Error) {
	modifiedBy, err := services.GetItemModifiedBy(itemID)
	if err != nil {
		return "", m.makeDbError(err)
	}
	return modifiedBy, nil
}

func (m *itemMethods) GetChildrenByID(itemID string) (messageType, *dbus.Error) {
	children, err := services.GetItemChildren(itemID)
	if err != nil {
		return nilMessage, m.makeDbError(err)
	}
	return m.serialize(children)
}

func (m *itemMethods) GetParentsByID(itemID string) (messageType, *dbus.Error) {
	parents, err := services.GetItemParents(itemID)
	if err != nil {
		return nilMessage, m.makeDbError(err)
	}
	return m.serialize(parents)
}

func (m *itemMethods) GetRelationsByID(itemID string) (messageType, *dbus.Error) {
	relations, err := services.GetItemRelations(itemID)
	if err != nil {
		return nilMessage, m.makeDbError(err)
	}
	return m.serialize(relations)
}

func (m *itemMethods) GetAttributesByID(itemID string) (messageType, *dbus.Error) {
	attributes, err := services.GetItemAttributes(itemID)
	if err != nil {
		return nilMessage, m.makeDbError(err)
	}
	return m.serialize(attributes)
}

func (m *itemMethods) GetAttributeIDByName(itemID, attributeName string) (messageType, *dbus.Error) {
	attributeId, err := services.GetItemAttributeIDByName(itemID, attributeName)
	if err != nil {
		return nilMessage, m.makeDbError(err)
	}
	return m.serialize(attributeId)
}

func (m *itemMethods) GetAttributeByName(itemID, attributeName string) (messageType, *dbus.Error) {
	attribute, err := services.GetItemAttributeByName(itemID, attributeName)
	if err != nil {
		return nilMessage, m.makeDbError(err)
	}
	return m.serialize(attribute)
}

func (m *itemMethods) GetAttributesByType(itemID, attributeType string) (messageType, *dbus.Error) {
	attributes, err := services.GetItemAttributesByType(itemID, attributeType)
	if err != nil {
		return nilMessage, m.makeDbError(err)
	}
	return m.serialize(attributes)
}
