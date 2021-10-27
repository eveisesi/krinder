package discord

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/eveisesi/krinder/internal/zkillboard"
	"github.com/pkg/errors"
)

func (s *Service) killrightResolver(msg string) bool {
	return strings.HasPrefix(msg, "killright")
}

func (s *Service) killrightExecutor(msg *discordgo.MessageCreate) error {

	stripped := strings.TrimPrefix(msg.Content, "killright ")

	id, err := strconv.ParseInt(stripped, 10, 64)
	if err != nil {
		return errors.Wrap(err, "failed to parse id to integer")
	}

	var killmails = make([]*zkillboard.Killmail, 0)
	i := int64(1)
	timeBoundary := time.Now().AddDate(0, 0, -14)
	timeBoundary = time.Date(timeBoundary.Year(), timeBoundary.Month(), timeBoundary.Day(), 0, 0, 0, 0, time.UTC)
	for {
		ctx := context.Background()

		killmailInteration, err := s.zkb.Killmails(ctx, "characterID", id, i)
		if err != nil {
			return errors.Wrap(err, "failed to fetch killmails")
		}

		if len(killmailInteration) == 0 {
			break
		}

		killmails = append(killmails, killmailInteration...)

		lastKill := killmails[len(killmailInteration)-1]

		killmail, err := s.esi.KillmailByIDHash(ctx, int64(lastKill.KillmailID), lastKill.Meta.Hash)
		if err != nil {
			return errors.Wrap(err, "failed to fetch killmail from esi")
		}

		if killmail.KillmailTime.Unix() < timeBoundary.Unix() {
			// We have at least 14 days worth of mails, maybe more
			// Additional filtering will be done after we fetch each mail
			fmt.Println(killmail.KillmailTime.Format("2006-01-02 15:04:05"), timeBoundary.Format("2006-01-02 15:04:05"))
			break
		}

		i++

	}

	_, err = s.session.ChannelMessageSend(msg.ChannelID, fmt.Sprintf("Found %d Killmails for This Character", len(killmails)))
	if err != nil {
		s.logger.WithError(err).Error("failed to send message")
		return err
	}

	return nil

}
