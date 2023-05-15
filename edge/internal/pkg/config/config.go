/*
Package config contains configuration structures and methods
to load configuration from files and environment.
*/
package config

import (
	"devais.it/kronos/internal/pkg/serialization"
	"devais.it/kronos/internal/pkg/types"
	"devais.it/kronos/internal/pkg/util"
	"fmt"
	"github.com/BurntSushi/toml"
	"github.com/mitchellh/mapstructure"
	"github.com/rotisserie/eris"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
	"io"
	"os"
	"reflect"
	"strconv"
	"strings"
)

const (
	// General
	envPrefix         = "kronos"
	defaultConfigName = "config"
	defaultDeviceID   = "default"
)

var (
	DeviceIDSourceMAC      = "mac"
	DeviceIDSourceHostname = "hostname"
	DeviceIDSourceConfig   = "config"
)

var (
	defaultConfigPaths = []string{
		"./res",
		"/etc/kronos",
	}
)

type Config struct {
	// Logging is the logging configuration
	Logging LoggingConfig

	// DB is the database configuration
	DB DBConfig

	// DBus is the DBus API configuration
	DBus DBusConfig

	// HTTP is the HTTP API configuration
	HTTP HTTPConfig

	// Sentry is the Sentry client configuration
	Sentry SentryConfig

	// Prometheus is the prometheus metrics server configuration
	Prometheus PrometheusConfig

	// Sync is the synchronization client configuration
	Sync SyncConfig

	// MaxProcs limits the number of operating system threads that
	// can run user-level Go code simultaneously.
	MaxProcs int

	// Variables
	//
	// You can refer to variables inside some string fields.
	// Fields which support variables are marked as such with a comment.
	// Variables can be used writing the name of a variable inside
	// brackets, i.g.: "test_string_{variableA}"
	// When the field string value is to be used, variable declarations will be
	// replaced with their values first.

	// Set to true if variable names should be case-sensitive
	UseCaseSensitiveVariables bool

	// Predefined variables
	// These are useful because setting their value through environment
	// variables is easier than defining a dictionary of custom variables

	// DeviceIDSource determines from where the device ID source should be loaded.
	// Valid values are:
	// "mac": The device ID is set to the MAC address of the first network interface
	// "hostname": The device ID is set to the device hostname
	// "config": The device ID is set from configuration with the "DeviceID" variable
	DeviceIDSource string

	DeviceID string

	CustomerID string

	CustomerName string

	TenantID string

	TenantName string

	// Custom variables dictionary
	Variables map[string]string
}

// DefaultConfig creates a new configuration structure
// filled with default options
func DefaultConfig() Config {
	return Config{
		Logging:    DefaultLoggingConfig(),
		DB:         DefaultDBConfig(),
		DBus:       DefaultDBusConfig(),
		HTTP:       DefaultHTTPConfig(),
		Sentry:     DefaultSentryConfig(),
		Prometheus: DefaultPrometheusConfig(),
		Sync:       DefaultSyncConfig(),
		MaxProcs:   0,
		// Variables
		UseCaseSensitiveVariables: false,
		DeviceIDSource:            DeviceIDSourceMAC,
		DeviceID:                  defaultDeviceID,
		CustomerID:                "",
		CustomerName:              "",
		TenantID:                  "",
		TenantName:                "",
		Variables:                 map[string]string{},
	}
}

// Parse loads the configuration
// using a pre initialized viper object
func Parse(v *viper.Viper, configFile string) (*Config, error) {
	var err error

	v.SetEnvPrefix(envPrefix)
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	if configFile != "" {
		v.SetConfigFile(configFile)
	} else {
		for i := 0; i < len(defaultConfigPaths); i++ {
			v.AddConfigPath(defaultConfigPaths[i])
		}

		v.SetConfigName(defaultConfigName)
	}

	err = v.ReadInConfig()
	if err != nil {
		switch err.(type) {
		case viper.ConfigFileNotFoundError:
			logrus.Info("Config file not found")
		default:
			return nil, eris.Wrap(err, "failed to read configuration file")
		}
	}

	config := DefaultConfig()

	// Create decode hooks to parse custom configuration types such as
	// logrus LogLevel or FileSize
	decodeHook := viper.DecodeHook(mapstructure.ComposeDecodeHookFunc(
		stringToLogLevelHookFunc(),
		stringToFileSizeHookFunc(),
		stringToVersionAlgoFunc(),
		stringToSerializationTypeFunc(),
		mapstructure.StringToTimeDurationHookFunc(),
		mapstructure.StringToIPHookFunc(),
		mapstructure.StringToSliceHookFunc(","),
	))

	// Workaround to make viper unmarshal from environment variables
	err = setViperDefaults(v, &config)
	if err != nil {
		return nil, eris.Wrap(err, "failed to set configuration defaults")
	}

	err = v.UnmarshalExact(&config, decodeHook)

	if err != nil {
		return nil, eris.Wrap(err, "failed to unmarshal configuration file")
	}

	// Set manual variables
	if maxProcs, ok := os.LookupEnv("GOMAXPROCS"); ok {
		cpus, err := strconv.ParseInt(maxProcs, 10, 32)
		if err != nil {
			return nil, eris.Wrap(err, "Failed to parse GOMAXPROCS variable")
		}

		config.MaxProcs = int(cpus)
	}

	return &config, nil
}

// PrintDebug logs the current viper configuration entries with trace level
func PrintDebug(v *viper.Viper) {
	redactPasswords := v.GetBool("logging.redactPasswords")

	for _, k := range v.AllKeys() {
		val := v.Get(k)

		if redactPasswords {
			// Redact passwords
			if strings.Contains(strings.ToLower(k), "password") {
				if reflect.TypeOf(val) == reflect.TypeOf("") {
					val = strings.Repeat("*", len(val.(string)))
				}
			}
		}

		logrus.Tracef("%s = %v", k, val)
	}
}

// SaveDefault outputs the default configuration through the given writer, as TOML.
func SaveDefault(w io.Writer, format string) error {
	defaultConf := DefaultConfig()

	var err error

	if format == "toml" {
		encoder := toml.NewEncoder(w)
		err = encoder.Encode(defaultConf)
	} else if format == "yaml" {
		encoder := yaml.NewEncoder(w)
		err = encoder.Encode(defaultConf)
	} else if format == "json" {
		json := serialization.DefaultJSONSerializer()
		json.Indent = "  "
		var jsonBytes []byte
		jsonBytes, err = json.Serialize(defaultConf)
		if err == nil {
			_, err = w.Write(jsonBytes)
		}
	} else if format == "env" {
		var configMap map[string]interface{}

		err = mapstructure.Decode(defaultConf, &configMap)
		if err != nil {
			return err
		}

		err = iterConfigMap(
			configMap,
			strings.ToUpper(envPrefix),
			"_",
			func(key string, value interface{}) error {
				line := fmt.Sprintf("%s=%v\n", strings.ToUpper(key), value)
				_, err := w.Write([]byte(line))
				return err
			},
		)
	} else {
		return eris.Errorf("Invalid configuration format: '%s'", format)
	}

	if err != nil {
		return eris.Wrap(err, "failed to encode default configuration")
	}

	return nil
}

// iterConfigMap performs an iteration over a map and calls the given
// callback for each element along with a generated key.
// The element key is generated using prefix and
// separator parameters.
func iterConfigMap(
	configMap map[string]interface{},
	prefix, separator string,
	fn func(string, interface{}) error) error {
	var err error

	for key, val := range configMap {
		viperKey := strings.ToLower(key)
		if prefix != "" {
			viperKey = prefix + separator + viperKey
		}

		switch v := val.(type) {
		case map[string]interface{}:
			err = iterConfigMap(v, viperKey, separator, fn)
		case types.FileSize:
			err = fn(viperKey, v.String())
		case logrus.Level:
			err = fn(viperKey, v.String())
		case util.VersionAlgorithm:
			err = fn(viperKey, v.String())
		case serialization.Type:
			err = fn(viperKey, v.String())
		default:
			err = fn(viperKey, val)
		}

		if err != nil {
			return err
		}
	}

	return nil
}

// setDefaultsFromMap sets Viper default values from a map
func setDefaultsFromMap(v *viper.Viper, configMap map[string]interface{}) error {
	return iterConfigMap(
		configMap,
		"",
		".",
		func(key string, value interface{}) error {
			v.SetDefault(key, value)
			return nil
		},
	)
}

// setViperDefaults sets defaults values to Viper reading them from
// a configuration structure.
func setViperDefaults(v *viper.Viper, configStruct interface{}) error {
	var configMap map[string]interface{}

	err := mapstructure.Decode(configStruct, &configMap)
	if err != nil {
		return err
	}

	return setDefaultsFromMap(v, configMap)
}

//=============================================================================
// mapstructure hooks
//=============================================================================

// stringToFileSizeHookFunc is a mapstructure decode hook
// which decodes strings to file sizes
func stringToFileSizeHookFunc() mapstructure.DecodeHookFunc {
	return func(f reflect.Type, t reflect.Type, data interface{}) (interface{}, error) {
		if f.Kind() != reflect.String || t != reflect.TypeOf(types.FileSize(0)) {
			return data, nil
		}
		return types.FileSizeFromString(data.(string))
	}
}

// stringToLogLevelHookFunc is a mapstructure decode hook
// which decodes strings to log levels
func stringToLogLevelHookFunc() mapstructure.DecodeHookFunc {
	return func(f reflect.Type, t reflect.Type, data interface{}) (interface{}, error) {
		if f.Kind() != reflect.String || t != reflect.TypeOf(logrus.DebugLevel) {
			return data, nil
		}
		return logrus.ParseLevel(data.(string))
	}
}

// stringToVersionAlgoFunc is a mapstructure decode hook
// which decodes strings to version algorithms
func stringToVersionAlgoFunc() mapstructure.DecodeHookFunc {
	return func(f reflect.Type, t reflect.Type, data interface{}) (interface{}, error) {
		if f.Kind() != reflect.String || t != reflect.TypeOf(util.VersionAlgorithmSha1) {
			return data, nil
		}
		return util.VersionAlgorithmFromString(data.(string))
	}
}

// stringToSerializationTypeFunc is a mapstructure decode hook
// which decodes strings to serialization type
func stringToSerializationTypeFunc() mapstructure.DecodeHookFunc {
	return func(f reflect.Type, t reflect.Type, data interface{}) (interface{}, error) {
		if f.Kind() != reflect.String || t != reflect.TypeOf(serialization.TypeJSON) {
			return data, nil
		}
		return serialization.TypeFromString(data.(string))
	}
}
