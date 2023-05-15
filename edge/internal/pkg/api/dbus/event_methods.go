package dbus

import (
	"devais.it/kronos/internal/pkg/serialization"
	"devais.it/kronos/internal/pkg/services"
	"github.com/godbus/dbus/v5"
	"github.com/godbus/dbus/v5/introspect"
)

const (
	onEventSignalName = "OnEvent"
)

type eventMethods struct {
	methodsBase
}

func newEventMethods(
	interfaceName string,
	serializer serialization.Serializer,
	deserializer serialization.Deserializer) *eventMethods {
	return &eventMethods{
		methodsBase{
			InterfaceName: interfaceName,
			Serializer:    serializer,
			Deserializer:  deserializer,
		},
	}
}

func (eventMethods) getSignals() []introspect.Signal {
	return []introspect.Signal{
		{
			Name: onEventSignalName,
			Args: []introspect.Arg{
				{Name: "entity_id", Type: "s", Direction: "out"},
				{Name: "entity_type", Type: "s", Direction: "out"},
				{Name: "event_type", Type: "s", Direction: "out"},
			},
			Annotations: nil,
		},
	}
}

func (m *eventMethods) GetFirstEvent() (messageType, *dbus.Error) {
	event, err := services.GetFirstEvent()
	if err != nil {
		return nilMessage, m.makeDbError(err)
	}
	return m.serialize(event)
}

func (m *eventMethods) GetLastEvent() (messageType, *dbus.Error) {
	event, err := services.GetLastEvent()
	if err != nil {
		return nilMessage, m.makeDbError(err)
	}
	return m.serialize(event)
}

func (m *eventMethods) Count() (int64, *dbus.Error) {
	count, err := services.GetEventsCount()
	if err != nil {
		return 0, m.makeDbError(err)
	}
	return count, nil
}
