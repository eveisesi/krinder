package discord

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/eveisesi/krinder/internal/esi"
	"github.com/eveisesi/krinder/internal/wars"
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
		killmail.SolarSystem = system

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

		// Add Checking for Wars
		victimCharacter, err := s.esi.Character(ctx, uint64(killmail.Victim.CharacterID))
		if err != nil {
			return errors.Wrap(err, "failed to fetch character from ESI")
		}

		killmail.Victim.Character = victimCharacter

		warEntityMatrix := [][]wars.Entity{}
		if searchedCharacter.CorporationID > 0 && victimCharacter.CorporationID > 0 {
			warEntityMatrix = append(warEntityMatrix, []wars.Entity{
				{T: "corporation", ID: searchedCharacter.CorporationID},
				{T: "corporation", ID: victimCharacter.CorporationID},
			})
		}
		if searchedCharacter.AllianceID > 0 && victimCharacter.AllianceID > 0 {
			warEntityMatrix = append(warEntityMatrix, []wars.Entity{
				{T: "alliance", ID: searchedCharacter.AllianceID},
				{T: "alliance", ID: victimCharacter.AllianceID},
			})
		}
		if searchedCharacter.CorporationID > 0 && victimCharacter.AllianceID > 0 {
			warEntityMatrix = append(warEntityMatrix, []wars.Entity{
				{T: "corporation", ID: searchedCharacter.CorporationID},
				{T: "alliance", ID: victimCharacter.AllianceID},
			})
		}
		if searchedCharacter.AllianceID > 0 && victimCharacter.CorporationID > 0 {
			warEntityMatrix = append(warEntityMatrix, []wars.Entity{
				{T: "alliance", ID: searchedCharacter.AllianceID},
				{T: "corporation", ID: victimCharacter.CorporationID},
			})
		}

		var warSkip = false
		for _, pair := range warEntityMatrix {
			atWar, err := s.wars.EntitiesAtWar(ctx, pair[0], pair[1])
			if err != nil {
				return errors.Wrap(err, "failed to determine if entities are at war")
			}

			warSkip = atWar
			if warSkip {
				break
			}
		}

		if warSkip {
			continue
		}

		filteredKillmails = append(filteredKillmails, killmail)

	}

	messages := make([]string, 0, len(filteredKillmails))
	seen := make(map[int]bool)
	for _, killmail := range filteredKillmails {
		_, ok := seen[killmail.Victim.CharacterID]
		if ok {
			continue
		}
		messages = append(messages, fmt.Sprintf(
			"%s killed %s (%d) on %s in %s (%.2f)",
			searchedCharacter.Name,
			killmail.Victim.Character.Name,
			killmail.KillmailID,
			killmail.KillmailTime.Format("2006-01-02"),
			killmail.SolarSystem.Name,
			killmail.SolarSystem.SecurityStatus,
		))
		seen[killmail.Victim.CharacterID] = true
	}

	j := 0
	l := 20
	end := j + l

	_, err = s.session.ChannelMessageSend(msg.ChannelID, appendLatencyToMessageCreate(msg, fmt.Sprintf("Found %d potential killrights (Batches of %d):\n```<Attacker> killed <Victim> (<Killmail ID>) on <Date> in <System> (<System Sec>)``` ", len(messages), l), false))
	if err != nil {
		s.logger.WithError(err).Error("failed to send message")
		return err
	}

	for {
		if end > len(messages) {
			end = len(messages)
		}
		message := strings.Join(messages[j:end], "\n")

		if message == "" {
			break
		}

		message = fmt.Sprintf("```%s```", message)
		_, err = s.session.ChannelMessageSend(msg.ChannelID, message)
		if err != nil {
			fmt.Println(message)
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
