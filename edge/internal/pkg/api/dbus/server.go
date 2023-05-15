package dbus

import (
	"devais.it/kronos/internal/pkg/config"
	"devais.it/kronos/internal/pkg/logging"
	"devais.it/kronos/internal/pkg/serialization"
	"devais.it/kronos/internal/pkg/sync/messages"
	"github.com/godbus/dbus/v5"
	"github.com/godbus/dbus/v5/introspect"
	"github.com/rotisserie/eris"
	log "github.com/sirupsen/logrus"
)

const (
	introspectableInterface = "org.freedesktop.DBus.Introspectable"
)

var (
	ErrServerDisabled   = eris.New("DBus server disabled by config")
	ErrNameAlreadyTaken = eris.New("DBus name already taken")
)

type Server struct {
	conn         *dbus.Conn
	conf         *config.DBusConfig
	serializer   serialization.Serializer
	deserializer serialization.Deserializer
}

// NewServer creates a new DBus server
func NewServer(conf *config.DBusConfig) (*Server, error) {
	if !conf.Enabled {
		return nil, ErrServerDisabled
	}

	var conn *dbus.Conn
	var err error

	if conf.UseSystemBus {
		conn, err = dbus.SystemBus()
	} else {
		conn, err = dbus.SessionBus()
	}

	if err != nil {
		return nil, eris.Wrap(err, "failed to connect to session bus")
	}

	server := &Server{
		conn: conn,
		conf: conf,
	}

	server.serializer, server.deserializer, err = conf.Serialization.NewSerializer()
	if err != nil {
		return nil, err
	}

	return server, nil
}

func (s *Server) Start() error {
	reply, err := s.conn.RequestName(s.conf.InterfaceName, dbus.NameFlagDoNotQueue)
	if err != nil {
		return eris.Wrap(err, "failed to request name")
	}

	if reply != dbus.RequestNameReplyPrimaryOwner {
		return ErrNameAlreadyTaken
	}

	err = s.exportMethods(buildAllMethods(s.conf, s.serializer, s.deserializer)...)
	if err != nil {
		return eris.Wrap(err, "failed to export interface methods")
	}

	log.Infof("DBus server listening on %s %s", s.conf.InterfaceName, s.conf.PathName)

	return nil
}

func (s *Server) Stop() error {
	err := s.conn.Close()
	if err == nil {
		s.conn = nil
		log.Infof("DBus server stopped")
	}
	return err
}

func (s *Server) SignalSyncEvent(sync messages.Sync) {
	for _, msg := range sync {
		err := s.conn.Emit(
			dbus.ObjectPath(s.conf.PathName),
			s.conf.EventsInterfaceName+"."+onEventSignalName,
			msg.EntityID,
			msg.EntityType,
			msg.Action,
		)
		if err != nil {
			logging.Error(err, "failed to emit signal")
		}
	}
}

// exportMethods is a wrapper around godbus export functions.
// It is used to export methods of a methods interface.
func (s *Server) exportMethods(methods ...methods) error {
	var err error

	node := &introspect.Node{}
	node.Name = s.conf.InterfaceName

	for _, m := range methods {
		iface := &introspect.Interface{}

		// Set interface name
		iface.Name = m.getInterface()

		// Set signals
		iface.Signals = m.getSignals()

		iface.Methods = append(iface.Methods, introspect.Methods(m)...)
		node.Interfaces = append(node.Interfaces, *iface)

		err = s.conn.Export(m, dbus.ObjectPath(s.conf.PathName), m.getInterface())
		if err != nil {
			return eris.Wrapf(err, "failed to export interface '%s'", m.getInterface())
		}

		for _, signal := range iface.Signals {
			err := s.conn.AddMatchSignal(dbus.WithMatchInterface(iface.Name + "." + signal.Name))
			if err != nil {
				return err
			}
		}
	}

	xmlIntro := introspect.NewIntrospectable(node)

	log.Tracef("DBus XML intro: %s", xmlIntro)

	err = s.conn.Export(xmlIntro, dbus.ObjectPath(s.conf.PathName), introspectableInterface)
	if err != nil {
		return eris.Wrap(err, "failed to export intro")
	}

	return nil
}

func buildAllMethods(
	conf *config.DBusConfig,
	serializer serialization.Serializer,
	deserializer serialization.Deserializer) []methods {
	return []methods{
		newItemMethods(conf.ItemsInterfaceName, serializer, deserializer),
		newRelationMethods(conf.RelationsInterfaceName, serializer, deserializer),
		newAttributeMethods(conf.AttributesInterfaceName, serializer, deserializer),
		newEventMethods(conf.EventsInterfaceName, serializer, deserializer),
		newConfigMethods(conf.ConfigInterfaceName, serializer, deserializer),
	}
}
