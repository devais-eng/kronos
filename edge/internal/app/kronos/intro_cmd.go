package kronos

import (
	"devais.it/kronos/internal/pkg/api/dbus"
	"devais.it/kronos/internal/pkg/config"
	"fmt"
	"github.com/go-xmlfmt/xmlfmt"
	"strings"
)

type dumpDbusIntroCmd struct {
	Ident int `kong:"arg,optional,name=ident,default=0,help=XML ident"`
}

func (c *dumpDbusIntroCmd) Run(*Context) error {
	conf := config.DefaultDBusConfig()

	intro := dbus.GetXMLIntro(&conf)

	if c.Ident > 0 {
		identStr := strings.Repeat(" ", c.Ident)

		intro = xmlfmt.FormatXML(intro, "", identStr)
	}

	fmt.Println(intro)

	return nil
}
