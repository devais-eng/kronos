package dbus

import (
	"github.com/godbus/dbus/v5"
	"github.com/rotisserie/eris"
	"github.com/spf13/viper"
	"gorm.io/gorm"
)

func makeError(iface, name string, err error) *dbus.Error {
	withTrace := viper.GetBool("dbus.errorsWithTrace")

	return &dbus.Error{
		Name: iface + ".Error." + name,
		Body: []interface{}{eris.ToString(err, withTrace)},
	}
}

func makeDbError(iface string, err error) *dbus.Error {
	if eris.Is(err, gorm.ErrRecordNotFound) {
		return makeNotFoundError(iface, err)
	}
	if eris.Is(err, gorm.ErrInvalidData) {
		return makeError(iface, "InvalidData", err)
	}
	return makeError(iface, "DbError", err)
}

func makeNotFoundError(iface string, err error) *dbus.Error {
	return makeError(iface, "NotFound", err)
}

func makeSerializationError(iface string, err error) *dbus.Error {
	return makeError(iface, "SerializationFailed", err)
}

func makeDeserializationError(iface string, err error) *dbus.Error {
	return makeError(iface, "DeserializationFailed", err)
}
