package kronos

import (
	"devais.it/kronos/internal/pkg/config"
	"fmt"
	"github.com/rotisserie/eris"
	"github.com/sirupsen/logrus"
	"os"
	"path/filepath"
	"strings"
)

type saveDefaultConfigCmd struct {
	Filename string `kong:"arg,name=file,help='Configuration file',type=file"`
	Format   string `kong:"arg,optional,name=format,enum=',toml,yaml,env,json',help='Configuration format (toml,yaml,env,json)'"`
}

func configFormatFromFilename(filename string) string {
	ext := strings.Replace(filepath.Ext(filename), ".", "", 1)

	switch ext {
	case "":
		return "toml"
	case "yml":
		return "yaml"
	default:
		return ext
	}
}

func (c *saveDefaultConfigCmd) Run(*Context) error {
	file, err := os.Create(c.Filename)
	if err != nil {
		return eris.Wrapf(err, "failed to open '%s' config file for creation", c.Filename)
	}
	defer func() {
		if err := file.Close(); err != nil {
			logrus.Errorf("failed to close file '%s': %s", c.Filename, eris.ToString(err, false))
		}
	}()

	var format string

	if c.Format == "" {
		format = configFormatFromFilename(c.Filename)
	} else {
		format = c.Format
	}

	err = config.SaveDefault(file, format)
	if err != nil {
		return err
	}

	fmt.Println("Default configuration saved to ", c.Filename)

	return nil
}

type printDefaultConfig struct {
	Format string `kong:"arg,optional,name=format,help='Configuration format (toml,yaml,json)',enum='toml,yaml,env,json',default=toml"`
}

func (c *printDefaultConfig) Run(*Context) error {
	return config.SaveDefault(os.Stdout, c.Format)
}
