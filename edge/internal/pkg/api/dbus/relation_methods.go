package dbus

import (
	"devais.it/kronos/internal/pkg/constants"
	"devais.it/kronos/internal/pkg/db/models"
	"devais.it/kronos/internal/pkg/serialization"
	"devais.it/kronos/internal/pkg/services"
	"github.com/godbus/dbus/v5"
)

type relationMethods struct {
	methodsBase
}

func newRelationMethods(
	interfaceName string,
	serializer serialization.Serializer,
	deserializer serialization.Deserializer) *relationMethods {
	return &relationMethods{
		methodsBase{
			InterfaceName: interfaceName,
			Serializer:    serializer,
			Deserializer:  deserializer,
		},
	}
}

func (m *relationMethods) Create(msg messageType) (messageType, *dbus.Error) {
	relation := &models.Relation{}
	return m.withSerializer(msg, relation, func() (interface{}, *dbus.Error) {
		err := services.CreateRelation(relation, constants.ModifiedByDBusAPIName)
		if err != nil {
			return nilMessage, m.makeDbError(err)
		}
		if m.replyCreatedData() {
			relation, err = services.GetRelation(relation.ParentID, relation.ChildID)
			if err != nil {
				return nil, m.makeDbError(err)
			}
			return relation, nil
		}

		return relation.CompositeID(), nil
	})
}

func (m *relationMethods) GetAll(page, pageSize int) (messageType, *dbus.Error) {
	relations, err := services.GetAllRelations(page, pageSize)
	if err != nil {
		return nilMessage, m.makeDbError(err)
	}
	return m.serialize(relations)
}

func (m *relationMethods) Get(parentID, childID string) (messageType, *dbus.Error) {
	relation, err := services.GetRelation(parentID, childID)
	if err != nil {
		return nilMessage, m.makeDbError(err)
	}
	return m.serialize(relation)
}

func (m *relationMethods) Delete(parentID, childID string) (messageType, *dbus.Error) {
	err := services.DeleteRelation(parentID, childID, constants.ModifiedByDBusAPIName)
	if err != nil {
		return nilMessage, m.makeDbError(err)
	}
	return m.serialize(&models.Relation{ParentID: parentID, ChildID: childID})
}

func (m *relationMethods) HardDelete(parentID, childID string) (messageType, *dbus.Error) {
	err := services.HardDeleteRelation(parentID, childID, constants.ModifiedByDBusAPIName)
	if err != nil {
		return nilMessage, m.makeDbError(err)
	}
	return m.serialize(&models.Relation{ParentID: parentID, ChildID: childID})
}

func (m *relationMethods) Count() (int64, *dbus.Error) {
	count, err := services.GetRelationsCount()
	if err != nil {
		return 0, m.makeDbError(err)
	}
	return count, nil
}
