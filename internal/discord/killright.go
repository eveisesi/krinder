package discord

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/eveisesi/krinder/internal/esi"
	"github.com/eveisesi/krinder/internal/zkillboard"
	"github.com/pkg/errors"
	"github.com/urfave/cli"
)

func (s *Service) killrightExecutor(c *cli.Context) error {

	msg, err := messageFromCLIContext(c)
	if err != nil {
		return err
	}

	args := c.Args()
	if len(args) > 1 {
		return errors.Errorf("expected 2 args, got %d. Surround name in double quotes \"<name>\"", len(args))
	}

	ctx := context.Background()
	id, err := strconv.ParseUint(args[0], 10, 64)
	if err != nil {
		return errors.Wrap(err, "failed to parse id to integer")
	}

	searchedCharacter, err := s.esi.Character(ctx, id)
	if err != nil {
		return errors.Wrap(err, "failed to fetch character from ESI")
	}

	var zmails = make([]*zkillboard.Killmail, 0)
	i := uint(1)
	timeBoundary := time.Now().AddDate(0, 0, -14)
	timeBoundary = time.Date(timeBoundary.Year(), timeBoundary.Month(), timeBoundary.Day(), 0, 0, 0, 0, time.UTC)
	for {

		killmailInteration, err := s.zkb.Killmails(ctx, "characterID", id, i)
		if err != nil {
			return errors.Wrap(err, "failed to fetch killmails")
		}

		if len(killmailInteration) == 0 {
			break
		}

		zmails = append(zmails, killmailInteration...)

		lastKill := zmails[len(killmailInteration)-1]

		killmail, err := s.esi.KillmailByIDHash(ctx, int64(lastKill.KillmailID), lastKill.Meta.Hash)
		if err != nil {
			return errors.Wrap(err, "failed to fetch killmail from esi")
		}

		if killmail.KillmailTime.Unix() < timeBoundary.Unix() {
			// We have at least 14 days worth of mails, maybe more
			// Additional filtering will be done after we fetch each mail
			fmt.Println("filtered page", killmail.KillmailTime.Format("2006-01-02 15:04:05"), timeBoundary.Format("2006-01-02 15:04:05"))
			break
		}

		i++

	}

	filteredKillmails := make([]*esi.KillmailOk, 0, len(zmails))
	for _, zmail := range zmails {
		killmail, err := s.esi.KillmailByIDHash(ctx, int64(zmail.KillmailID), zmail.Meta.Hash)
		if err != nil {
			return errors.Wrap(err, "failed to fetch killmail from ESI")
		}
		if killmail.KillmailTime.Unix() < timeBoundary.Unix() {
			break
		}

		system, err := s.esi.System(ctx, uint(killmail.SolarSystemID))
		if err != nil {
			return errors.Wrap(err, "failed to fetch killmail solar system from ESI")
		}

		if system.SecurityStatus < 0 {
			continue
		}

		// 670 is a pod/capsule
		if system.SecurityStatus < .5 && killmail.Victim.ShipTypeID != 670 {
			continue
		}

		if killmail.Victim.CharacterID == 0 {
			continue
		}

		killmail.SolarSystem = system

		// Add Checking for Wars
		victimCharacter, err := s.esi.Character(ctx, uint64(killmail.Victim.CharacterID))
		if err != nil {
			return errors.Wrap(err, "failed to fetch character from ESI")
		}

		killmail.Victim.Character = victimCharacter

		filteredKillmails = append(filteredKillmails, killmail)

	}

	// // Dedup Victims
	// mapVictimCharacters := make(map[uint64]*esi.KillmailOk)
	// for _, killmail := range filteredKillmails {
	// 	known, ok := mapVictimCharacters[uint64(killmail.Victim.CharacterID)]
	// 	if !ok {
	// 		mapVictimCharacters[uint64(killmail.Victim.CharacterID)] = killmail
	// 		continue
	// 	}

	// 	if killmail.KillmailTime.Unix() > known.KillmailTime.Unix() {
	// 		mapVictimCharacters[uint64(killmail.Victim.CharacterID)] = killmail
	// 	}
	// }

	messages := make([]string, 0, len(filteredKillmails))
	for _, killmail := range filteredKillmails {
		messages = append(messages, fmt.Sprintf(
			"%s killed %s (%d) on %s in %s (%.2f)",
			searchedCharacter.Name,
			killmail.Victim.Character.Name,
			killmail.KillmailID,
			killmail.KillmailTime.Format("2006-01-02"),
			killmail.SolarSystem.Name,
			killmail.SolarSystem.SecurityStatus,
		))
	}

	go func(channelID string) {
		err := s.session.ChannelTyping(channelID)
		if err != nil {
			s.logger.WithError(err).Error("failed to set typing status")
		}
	}(msg.ChannelID)

	_, err = s.session.ChannelMessageSend(msg.ChannelID, appendLatencyToMessageCreate(msg, fmt.Sprintf("Found %d potential killrights (Batches of 25):\n```<Attacker> killed <Victim> (<Killmail ID>) on <Date> in <System> (<System Sec>)``` ", len(filteredKillmails)), false))
	if err != nil {
		s.logger.WithError(err).Error("failed to send message")
		return err
	}

	j := 0
	l := 25
	end := j + l
	for {
		if end > len(messages) {
			end = len(messages)
		}
		message := strings.Join(messages[j:end], "\n")

		if message == "" {
			break
		}

		_, err = s.session.ChannelMessageSend(msg.ChannelID, fmt.Sprintf("```%s```", message))
		if err != nil {
			s.logger.WithError(err).Error("failed to send message")
			return err
		}

		if j >= len(messages) {
			break
		}

		j = j + l
		if j > len(messages) {
			break
		}
		end = j + l
		time.Sleep(time.Second)
	}

	return nil

}
