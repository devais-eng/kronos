package dbus

import (
	"devais.it/kronos/internal/pkg/serialization"
	"github.com/godbus/dbus/v5"
	"github.com/godbus/dbus/v5/introspect"
	"github.com/spf13/viper"
)

// Typedefs useful to change messages type
type messageType string

var (
	nilMessage messageType = ""
)

type methods interface {
	// getInterface should return the DBus interface name for this object
	getInterface() string

	// getSignals should return the signals that should be exported by this
	// object
	getSignals() []introspect.Signal
}

type methodsBase struct {
	InterfaceName string
	Serializer    serialization.Serializer
	Deserializer  serialization.Deserializer
}

func (m *methodsBase) getInterface() string {
	return m.InterfaceName
}

func (m *methodsBase) getSignals() []introspect.Signal {
	return nil
}

func (m *methodsBase) replyCreatedData() bool {
	return viper.GetBool("dbus.replyCreatedData")
}

func (m *methodsBase) deserialize(msg messageType, v interface{}) *dbus.Error {
	err := m.Deserializer.Deserialize([]byte(msg), v)
	if err != nil {
		return m.makeDeserializationError(err)
	}

	return nil
}

func (m *methodsBase) serialize(v interface{}) (messageType, *dbus.Error) {
	result, err := m.Serializer.Serialize(v)
	if err != nil {
		return nilMessage, m.makeSerializationError(err)
	}
	return messageType(result), nil
}

func (m *methodsBase) withSerializer(
	msg messageType,
	model interface{},
	f func() (interface{}, *dbus.Error)) (messageType, *dbus.Error) {
	dErr := m.deserialize(msg, model)
	if dErr != nil {
		return nilMessage, dErr
	}
	result, dErr := f()
	if dErr != nil {
		return nilMessage, dErr
	}
	return m.serialize(result)
}

func (m *methodsBase) makeSerializationError(err error) *dbus.Error {
	return makeSerializationError(m.InterfaceName, err)
}

func (m *methodsBase) makeDeserializationError(err error) *dbus.Error {
	return makeDeserializationError(m.InterfaceName, err)
}

func (m *methodsBase) makeDbError(err error) *dbus.Error {
	return makeDbError(m.InterfaceName, err)
}
