package dbus

import (
	"devais.it/kronos/internal/pkg/constants"
	"devais.it/kronos/internal/pkg/db/models"
	"devais.it/kronos/internal/pkg/serialization"
	"devais.it/kronos/internal/pkg/services"
	"github.com/godbus/dbus/v5"
)

type attributeMethods struct {
	methodsBase
}

func newAttributeMethods(
	interfaceName string,
	serializer serialization.Serializer,
	deserializer serialization.Deserializer) *attributeMethods {
	return &attributeMethods{
		methodsBase{
			InterfaceName: interfaceName,
			Serializer:    serializer,
			Deserializer:  deserializer,
		},
	}
}

func (m *attributeMethods) Create(msg messageType) (messageType, *dbus.Error) {
	attribute := &models.Attribute{}

	if dErr := m.deserialize(msg, attribute); dErr != nil {
		return nilMessage, dErr
	}

	err := services.CreateAttribute(attribute, constants.ModifiedByDBusAPIName)
	if err != nil {
		return nilMessage, m.makeDbError(err)
	}

	if m.replyCreatedData() {
		attribute, err = services.GetAttributeByID(attribute.ID)
		if err != nil {
			return nilMessage, m.makeDbError(err)
		}
		return m.serialize(attribute)
	}

	return m.serialize(attribute.ID)
}

func (m *attributeMethods) CreateBatch(msg messageType) (messageType, *dbus.Error) {
	var attributes []models.Attribute
	return m.withSerializer(msg, &attributes, func() (interface{}, *dbus.Error) {
		err := services.BatchCreateAttributes(attributes, constants.ModifiedByDBusAPIName)
		if err != nil {
			return nil, m.makeDbError(err)
		}
		ids := make([]string, len(attributes))
		for i := 0; i < len(attributes); i++ {
			ids[i] = attributes[i].ID
		}

		if m.replyCreatedData() {
			attributes, err = services.GetAttributesByIDs(ids)
			if err != nil {
				return nil, m.makeDbError(err)
			}
			return attributes, nil
		}

		return ids, nil
	})
}

func (m *attributeMethods) GetByID(attributeID string) (messageType, *dbus.Error) {
	attribute, err := services.GetAttributeByID(attributeID)
	if err != nil {
		return nilMessage, m.makeDbError(err)
	}
	return m.serialize(attribute)
}

func (m *attributeMethods) GetByType(attributeType string, page, pageSize int) (messageType, *dbus.Error) {
	attributes, err := services.GetAttributesByType(attributeType, page, pageSize)
	if err != nil {
		return nilMessage, m.makeDbError(err)
	}
	return m.serialize(attributes)
}

func (m *attributeMethods) FindByName(name string, page, pageSize int) (messageType, *dbus.Error) {
	attributes, err := services.FindAttributesByName(name, page, pageSize)
	if err != nil {
		return nilMessage, m.makeDbError(err)
	}
	return m.serialize(attributes)
}

func (m *attributeMethods) GetValue(attributeID string) (messageType, *dbus.Error) {
	value, err := services.GetAttributeValue(attributeID)
	if err != nil {
		return nilMessage, m.makeDbError(err)
	}
	return m.serialize(value)
}

func (m *attributeMethods) Update(msg messageType) (messageType, *dbus.Error) {
	var patch map[string]interface{}
	return m.withSerializer(msg, &patch, func() (interface{}, *dbus.Error) {
		err := services.UpdateAttribute(patch, constants.ModifiedByDBusAPIName)
		if err != nil {
			return nil, m.makeDbError(err)
		}
		attribute, err := services.GetAttributeByID(patch["id"].(string))
		if err != nil {
			return nil, m.makeDbError(err)
		}
		return attribute, nil
	})
}

func (m *attributeMethods) DeleteByID(attributeID string) (messageType, *dbus.Error) {
	err := services.DeleteAttributeByID(attributeID, constants.ModifiedByDBusAPIName)
	if err != nil {
		return nilMessage, m.makeDbError(err)
	}
	return m.serialize(attributeID)
}

func (m *attributeMethods) HardDeleteByID(attributeID string) (messageType, *dbus.Error) {
	err := services.HardDeleteAttributeByID(attributeID, constants.ModifiedByDBusAPIName)
	if err != nil {
		return nilMessage, m.makeDbError(err)
	}
	return m.serialize(attributeID)
}

func (m *attributeMethods) GetAll(page, pageSize int) (messageType, *dbus.Error) {
	attributes, err := services.GetAllAttributes(page, pageSize)
	if err != nil {
		return nilMessage, m.makeDbError(err)
	}
	return m.serialize(attributes)
}

func (m *attributeMethods) Count() (int64, *dbus.Error) {
	count, err := services.GetAttributesCount()
	if err != nil {
		return 0, m.makeDbError(err)
	}
	return count, nil
}
