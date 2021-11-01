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
				Usage:              "Returns a list of victims that may have a killright against the supplied characters name based on kill mail history",
				HelpName:           "killright",
				UsageText:          "killright <characterID>",
				Action:             s.killrightExecutor,
				CustomHelpTemplate: CommandHelpTemplate,
				Aliases:            []string{"kr"},

				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:    "format",
						Aliases: []string{"f"},
						Usage:   "Format of the list outputted. Options include simple, details, evelinks",
						Value:   "simple",
					},
					&cli.StringFlag{
						Name:    "perspective",
						Aliases: []string{"p"},
						Usage:   "Perspective to analyze killmails from. Valid values are agressor, victim, ship",
						Value:   "aggressor",
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
