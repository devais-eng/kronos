package dbus

import (
	"devais.it/kronos/internal/pkg/config"
	"devais.it/kronos/internal/pkg/serialization"
	"github.com/godbus/dbus/v5"
	"github.com/rotisserie/eris"
	"github.com/spf13/viper"
)

type configMethods struct {
	methodsBase
}

func newConfigMethods(
	interfaceName string,
	serializer serialization.Serializer,
	deserializer serialization.Deserializer) *configMethods {
	return &configMethods{
		methodsBase{
			InterfaceName: interfaceName,
			Serializer:    serializer,
			Deserializer:  deserializer,
		},
	}
}

func (m *configMethods) makeConfigKeyNotFound(key string) *dbus.Error {
	err := eris.Errorf("Config key '%s' not found", key)
	return makeNotFoundError(m.InterfaceName, err)
}

func (m *configMethods) makeGlobalEnvError(err error) *dbus.Error {
	return makeError(m.InterfaceName, "ConfigVariablesError", err)
}

func (m *configMethods) GetString(key string) (string, *dbus.Error) {
	if !viper.IsSet(key) {
		return "", m.makeConfigKeyNotFound(key)
	}
	return viper.GetString(key), nil
}

func (m *configMethods) GetInt(key string) (int, *dbus.Error) {
	if !viper.IsSet(key) {
		return 0, m.makeConfigKeyNotFound(key)
	}
	return viper.GetInt(key), nil
}

func (m *configMethods) GetBool(key string) (bool, *dbus.Error) {
	if !viper.IsSet(key) {
		return false, m.makeConfigKeyNotFound(key)
	}
	return viper.GetBool(key), nil
}

func (m *configMethods) GetFloat(key string) (float64, *dbus.Error) {
	if !viper.IsSet(key) {
		return 0.0, m.makeConfigKeyNotFound(key)
	}
	return viper.GetFloat64(key), nil
}

func (m *configMethods) GetDuration(key string) (string, *dbus.Error) {
	if !viper.IsSet(key) {
		return "", m.makeConfigKeyNotFound(key)
	}
	return viper.GetDuration(key).String(), nil
}

func (m *configMethods) GetAllVariables() (messageType, *dbus.Error) {
	globalEnv, err := config.GetGlobalEnvironment()
	if err != nil {
		return nilMessage, m.makeGlobalEnvError(err)
	}

	variables := globalEnv.ToMap()

	return m.serialize(variables)
}

func (m *configMethods) GetVariable(name string) (string, *dbus.Error) {
	globalEnv, err := config.GetGlobalEnvironment()
	if err != nil {
		return "", m.makeGlobalEnvError(err)
	}

	return globalEnv.Get(name), nil
}

func (m *configMethods) SetVariable(name, value string) (string, *dbus.Error) {
	globalEnv, err := config.GetGlobalEnvironment()
	if err != nil {
		return "", m.makeGlobalEnvError(err)
	}

	globalEnv.Set(name, value)

	return value, nil
}
