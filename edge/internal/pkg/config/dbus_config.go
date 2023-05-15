package config

const (
	defaultDBusPathName                = "/it/devais/kronos"
	defaultDBusInterfaceName           = "it.devais.kronos"
	defaultDBusItemsInterfaceName      = "it.devais.kronos.Items"
	defaultDBusRelationsInterfaceName  = "it.devais.kronos.Relations"
	defaultDBusAttributesInterfaceName = "it.devais.kronos.Attributes"
	defaultDBusEventsInterfaceName     = "it.devais.kronos.Events"
	defaultDBusConfigInterfaceName     = "it.devais.kronos.Config"
)

type DBusConfig struct {
	// Enabled determines if DBus integration should be enabled
	Enabled bool

	// UseSystemBus sets whether to use the system bus or not.
	// If false, session bus is used instead
	UseSystemBus bool

	// Serialization is the configuration of messages serialization
	Serialization SerializationConfig

	// ErrorsWithTrace determines if errors that occurs during DBus
	// methods calls should be reported with full trace
	ErrorsWithTrace bool

	// ReplyCreatedData sets whether created entities should be replied
	// in its entirety.
	// If false, when an entity is successfully created, only its ID is replied
	ReplyCreatedData bool

	// PathName is the DBus service root path name
	PathName string

	// InterfaceName is the DBus service root interface name
	InterfaceName string

	// ItemsInterfaceName is the DBus interface name for items
	ItemsInterfaceName string

	// RelationsInterfaceName is the DBus interface name for relations
	RelationsInterfaceName string

	// AttributesInterfaceName is the DBus interface name for attributes
	AttributesInterfaceName string

	// EventsInterfaceName is the DBus interface name for events
	EventsInterfaceName string

	// ConfigInterfaceName is the DBus interface name for configuration
	ConfigInterfaceName string
}

// DefaultDBusConfig creates a new DBus configuration structure
// filled with default options
func DefaultDBusConfig() DBusConfig {
	return DBusConfig{
		Enabled:                 false,
		UseSystemBus:            false,
		Serialization:           DefaultSerializationConfig(),
		ErrorsWithTrace:         false,
		ReplyCreatedData:        false,
		PathName:                defaultDBusPathName,
		InterfaceName:           defaultDBusInterfaceName,
		ItemsInterfaceName:      defaultDBusItemsInterfaceName,
		RelationsInterfaceName:  defaultDBusRelationsInterfaceName,
		AttributesInterfaceName: defaultDBusAttributesInterfaceName,
		EventsInterfaceName:     defaultDBusEventsInterfaceName,
		ConfigInterfaceName:     defaultDBusConfigInterfaceName,
	}
}
