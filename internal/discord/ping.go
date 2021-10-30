package discord

import (
	"github.com/urfave/cli"
)

func (s *Service) pingCommand(c *cli.Context) error {

	msg, err := messageFromCLIContext(c)
	if err != nil {
		return err
	}

	_, err = s.session.ChannelMessageSend(msg.ChannelID, appendLatencyToMessageCreate(msg, "Pong!", true))

	return err

}
