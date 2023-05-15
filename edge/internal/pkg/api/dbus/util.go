package dbus

import (
	"devais.it/kronos/internal/pkg/config"
	"github.com/godbus/dbus/v5/introspect"
)

func GetXMLIntro(conf *config.DBusConfig) string {
	allMethods := buildAllMethods(conf, nil, nil)

	node := &introspect.Node{}
	node.Name = conf.InterfaceName

	for _, m := range allMethods {
		iface := &introspect.Interface{}
		iface.Name = m.getInterface()
		iface.Signals = m.getSignals()
		iface.Methods = append(iface.Methods, introspect.Methods(m)...)

		node.Interfaces = append(node.Interfaces, *iface)
	}

	xmlIntro := introspect.NewIntrospectable(node)

	return string(xmlIntro)
}
