package kronos

import (
	"os"
)

type runCmd struct{}

func (c *runCmd) Run(*Context) error {
	os.Exit(run(CLI.ConfigFile, CLI.Verbose))
	return nil
}
