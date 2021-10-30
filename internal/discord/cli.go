package discord

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/urfave/cli"
)

var buf = &bytes.Buffer{}
var CommandHelpTemplate = fmt.Sprintf("```%s```", cli.CommandHelpTemplate)

func (s *Service) initializeCLI() *cli.App {
	return &cli.App{
		Name:                  "KRinder Discord Commands",
		HelpName:              filepath.Base(os.Args[0]),
		Usage:                 "<command>",
		UsageText:             "command [command options] [arguments...]",
		Action:                cli.ShowAppHelp,
		Compiled:              time.Now(),
		Writer:                buf,
		CustomAppHelpTemplate: fmt.Sprintf("```%s```", cli.AppHelpTemplate),
		Commands: []cli.Command{
			{
				Name:               "ping",
				HelpName:           "ping",
				Usage:              "Play Ping Pong. Test Bot Connectivity and Latency",
				UsageText:          "ping",
				Action:             s.pingCommand,
				CustomHelpTemplate: CommandHelpTemplate,
			},
			{
				Name:               "help",
				HelpName:           "help",
				Usage:              "Display this help text",
				UsageText:          "help",
				Action:             cli.ShowAppHelp,
				CustomHelpTemplate: CommandHelpTemplate,
			},
			{
				Name:               "search",
				Usage:              fmt.Sprintf("Searches an entity by name. Valid categories are (%s)", strings.Join(validCategories, ",")),
				HelpName:           "search",
				UsageText:          "search <category> <term>",
				Action:             s.searchCommand,
				CustomHelpTemplate: CommandHelpTemplate,
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:  "strict",
						Usage: "enables strict searching. Should return a single result",
					},
				},
			},
			{
				Name:               "killright",
				Usage:              "Returns a list of victims that may have a killright against the supplied characters name based on kill mail history",
				HelpName:           "killright",
				UsageText:          "killright <characterID>",
				Action:             s.killrightExecutor,
				CustomHelpTemplate: CommandHelpTemplate,
			},
		},
		Metadata: make(map[string]interface{}),
	}
}

func (s *Service) shouldRunCLI(cli *cli.App, words []string) bool {

	if len(words) == 0 {
		return false
	}

	for _, command := range cli.Commands {
		if command.Name == words[0] {
			return true
		}
	}

	return false

}
