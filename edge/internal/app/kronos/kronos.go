package kronos

import (
	"devais.it/kronos/internal/pkg/constants"
	"devais.it/kronos/internal/pkg/version"
	"github.com/alecthomas/kong"
	"strings"
)

var (
	// CLI Defines the Kong CLI structure
	CLI struct {
		ConfigFile         string                `kong:"short=c,help='Set configuration file path',xor:config,type:existingfile"`
		Verbose            bool                  `kong:"short=d,help='Enable verbose logging output'"`
		Version            kong.VersionFlag      `kong:"short=v,help='Print program version',xor=config"`
		Run                runCmd                `kong:"cmd,help='Run the service'"`
		PrintDefaultConfig printDefaultConfig    `kong:"cmd,help='Print default configuration'"`
		SaveDefaultConfig  saveDefaultConfigCmd  `kong:"cmd,help='Save default configuration to file'"`
		PrintMessageSchema printMessageSchemaCmd `kong:"cmd,help='Print synchronization messages as JSON schema'"`
		SaveMessageSchema  saveMessageSchemaCmd  `kong:"cmd,help='Save synchronization messages to JSON schema'"`
		DumpDbusIntro      dumpDbusIntroCmd      `kong:"cmd,help='Print DBus introspectable XML file'"`
	}
)

// Execute parses command line commands and arguments and executes them
func Execute() {
	kongCtx := kong.Parse(&CLI,
		kong.Name(strings.ToLower(constants.AppName)),
		kong.Description("The definitive client/server synchronization tool"),
		kong.UsageOnError(),
		kong.ConfigureHelp(kong.HelpOptions{
			Compact: true,
		}),
		kong.Vars{
			"version": constants.AppName + " " + version.GetVersionString(true),
		})

	err := kongCtx.Run(&Context{})
	kongCtx.FatalIfErrorf(err)
}
