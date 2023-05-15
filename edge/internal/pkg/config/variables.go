package config

import (
	"devais.it/kronos/internal/pkg/util"
	"github.com/rotisserie/eris"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"os"
	"strings"
)

var (
	ErrGlobalEnvUninitialized = eris.New("Global variables environment is not initialized")
)

var globalEnv *util.Environment

func InitGlobalEnvironment() error {
	if globalEnv != nil {
		return nil
	}

	var env *util.Environment

	if viper.GetBool("UseCaseSensitiveVariables") {
		env = util.NewEnvironmentCaseSensitive(nil)
	} else {
		env = util.NewEnvironment(nil)
	}

	hostname, err := os.Hostname()
	if err != nil {
		return eris.Wrap(err, "failed to read hostname")
	}

	env.Set("hostname", hostname)

	// Set config variables

	// Predefined variables
	predefinedVarNames := []string{
		"deviceID",
		"customerID",
		"customerName",
		"tenantID",
		"tenantName",
	}

	for _, varName := range predefinedVarNames {
		env.Set(varName, viper.GetString(varName))
	}

	customVariables := viper.GetStringMapString("Variables")
	if customVariables != nil {
		env.SetFromMap(customVariables)
	}

	deviceIDSource := strings.ToLower(viper.GetString("deviceIDSource"))
	// Set default
	if len(deviceIDSource) == 0 {
		deviceIDSource = DeviceIDSourceMAC
	}

	switch deviceIDSource {
	case DeviceIDSourceConfig:
		break
	case DeviceIDSourceHostname:
		env.Set("deviceID", hostname)
	case DeviceIDSourceMAC:
		macAddr, err := util.ReadMACAddress()
		if err != nil {
			return eris.Wrap(err, "failed to read device MAC address")
		}

		// Remove all ':'
		macAddrStr := strings.ReplaceAll(macAddr.String(), ":", "")

		env.Set("deviceID", macAddrStr)
	default:
		return eris.Errorf("Invalid deviceID source: '%s'", deviceIDSource)
	}

	log.Debugf("DeviceID: %s", env.Get("deviceID"))

	globalEnv = env

	log.Debug("Global variables environment initialized")

	return nil
}

func GetGlobalEnvironment() (*util.Environment, error) {
	if globalEnv == nil {
		return nil, ErrGlobalEnvUninitialized
	}
	return globalEnv, nil
}

// EscapeStringVariables will escape variables inside a given string
// using the global environment
func EscapeStringVariables(str string) (string, error) {
	env, err := GetGlobalEnvironment()
	if err != nil {
		return "", err
	}
	return env.EscapeStringVariables(str)
}
