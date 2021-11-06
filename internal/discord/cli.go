package discord

import (
	"bytes"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/urfave/cli/v2"
)

var buf = &bytes.Buffer{}
var CommandHelpTemplate = fmt.Sprintf("```%s```", cli.CommandHelpTemplate)
var SubCommandHelpTemplate = fmt.Sprintf("```%s```", cli.SubcommandHelpTemplate)

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
		Commands: []*cli.Command{
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
				Usage:              "",
				HelpName:           "killright",
				CustomHelpTemplate: SubCommandHelpTemplate,
				Aliases:            []string{"kr"},
				Action:             cli.ShowAppHelp,
				Subcommands: []*cli.Command{
					{
						Name:               "attacker",
						Usage:              "Search for Kill Rights by Killmail Attacker",
						UsageText:          "kr a <characterID>",
						HelpName:           "attacker",
						Description:        "Search for Kill Rights by Killmail Attacker",
						Aliases:            []string{"a"},
						Action:             s.killrightAttackerCommand,
						CustomHelpTemplate: CommandHelpTemplate,
					},
					{
						Name:               "victim",
						Description:        "Search for Kill Rights by Killmail Victim",
						Usage:              "Search for Kill Rights by Killmail Victim",
						UsageText:          "kr v <characterID>",
						Aliases:            []string{"v"},
						Action:             s.killrightVictimCommand,
						CustomHelpTemplate: CommandHelpTemplate,
					},
					{
						Name:               "ship",
						Description:        "Search for Kill Rights by Ship Group and Type. If you don't know the ship group id, please use the search command with the invgroup command to find a group by name. Type ID is optional, so you can ommit it, but the output will be a count of kills by Group, rather than charaters you can prosue. Pass ship type id to get the summary for that ship. They are printed when in the group summary",
						Usage:              "Search for Kill Rights by Ship Group and Type",
						UsageText:          "kr s <shipGroupID> <shipTypeID>",
						CustomHelpTemplate: CommandHelpTemplate,
					},
				},
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:    "format",
						Aliases: []string{"f"},
						Usage:   "Format of the list outputted. Options include simple, details, evelinks",
						Value:   "simple",
					},
				},
			},
			{
				Name:      "mail",
				Usage:     "Generates a link to the killmail on zkillboard.com",
				HelpName:  "mail",
				UsageText: "mail <killmailID>",
				Action: func(c *cli.Context) error {

					msg, err := messageFromCLIContext(c)
					if err != nil {
						return err
					}

					args := c.Args()
					if args.Len() > 1 {
						return errors.Errorf("expected 1 arg, got %d", args.Len())
					}
					id, err := strconv.ParseUint(args.Get(0), 10, 32)
					if err != nil {
						return errors.Wrap(err, "failed to parse killmail id to integer")
					}

					path := fmt.Sprintf("/kill/%d", id)

					uri := url.URL{
						Scheme: "https",
						Host:   "zkillboard.com",
						Path:   path,
					}

					_, err = s.session.ChannelMessageSend(msg.ChannelID, uri.String())
					if err != nil {
						return errors.Wrap(err, "failed to send message")
					}

					return nil
				},
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
		if len(command.Aliases) > 0 {
			for _, alias := range command.Aliases {
				if alias == words[0] {
					return true
				}
			}
		}
	}

	return false

}
