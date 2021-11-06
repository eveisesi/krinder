package discord

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/davecgh/go-spew/spew"
	"github.com/eveisesi/krinder/internal/esi"
	"github.com/eveisesi/krinder/internal/wars"
	"github.com/eveisesi/krinder/internal/zkillboard"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
)

var initialBoundary = time.Now().AddDate(0, 0, -14)
var timeBoundary = time.Date(initialBoundary.Year(), initialBoundary.Month(), initialBoundary.Day(), 0, 0, 0, 0, time.UTC)

func (s *Service) killrightAttackerCommand(c *cli.Context) error {

	msg, err := messageFromCLIContext(c)
	if err != nil {
		return err
	}

	args := c.Args()

	ctx := context.Background()
	id, err := strconv.ParseUint(args.Get(0), 10, 64)
	if err != nil {
		return errors.Wrap(err, "failed to parse id to integer")
	}

	var zmails = make([]*zkillboard.Killmail, 0)
	i := uint(1)
	for {

		killmailInteration, err := s.zkb.Killmails(ctx, zkillboard.CharacterEntityType, id, zkillboard.KillsFetchType, i)
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

	_, err = s.session.ChannelMessageSend(msg.ChannelID, fmt.Sprintf("found %d killmails, normalizing with ESI....", len(zmails)))
	if err != nil {
		s.logger.WithError(err).Errorln("failed to send message")
	}

	normalizedKillmails := make([]*esi.KillmailOk, 0, len(zmails))
	for _, zmail := range zmails {
		killmail, err := s.esi.KillmailByIDHash(ctx, int64(zmail.KillmailID), zmail.Meta.Hash)
		if err != nil {
			return errors.Wrap(err, "failed to fetch killmail from ESI")
		}
		if killmail.KillmailTime.Unix() < timeBoundary.Unix() {
			s.logger.WithFields(logrus.Fields{
				"killTime":     killmail.KillmailTime.Format("2006-01-02 15:04:05"),
				"timeBoundary": timeBoundary.Format("2006-01-02 15:04:05"),
			}).Debugln("outside time boundary, skipping")
			break
		}
		normalizedKillmails = append(normalizedKillmails, killmail)
	}

	s.logger.WithFields(logrus.Fields{
		"normalizedKillmails": len(normalizedKillmails),
	}).Debugln()

	return s.analyzeAgressorKillmail(ctx, c, msg, id, normalizedKillmails)
}

func (s *Service) killrightVictimCommand(c *cli.Context) error {

	msg, err := messageFromCLIContext(c)
	if err != nil {
		return err
	}

	args := c.Args()

	ctx := context.Background()
	id, err := strconv.ParseUint(args.Get(0), 10, 64)
	if err != nil {
		return errors.Wrap(err, "failed to parse id to integer")
	}

	var zmails = make([]*zkillboard.Killmail, 0)
	i := uint(1)
	for {

		killmailInteration, err := s.zkb.Killmails(ctx, zkillboard.CharacterEntityType, id, zkillboard.LossesFetchType, i)
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

	if len(zmails) == 0 {
		_, err = s.session.ChannelMessageSend(msg.ChannelID, "found 0 killmails, halting search")
		if err != nil {
			s.logger.WithError(err).Errorln("failed to send message")
		}
	}

	_, err = s.session.ChannelMessageSend(msg.ChannelID, fmt.Sprintf("found %d killmails, normalizing with ESI....", len(zmails)))
	if err != nil {
		s.logger.WithError(err).Errorln("failed to send message")
	}

	normalizedKillmails := make([]*esi.KillmailOk, 0, len(zmails))
	for _, zmail := range zmails {
		killmail, err := s.esi.KillmailByIDHash(ctx, int64(zmail.KillmailID), zmail.Meta.Hash)
		if err != nil {
			return errors.Wrap(err, "failed to fetch killmail from ESI")
		}
		if killmail.KillmailTime.Unix() < timeBoundary.Unix() {
			s.logger.WithFields(logrus.Fields{
				"killTime":     killmail.KillmailTime.Format("2006-01-02 15:04:05"),
				"timeBoundary": timeBoundary.Format("2006-01-02 15:04:05"),
			}).Debugln("outside time boundary, skipping")
			break
		}
		normalizedKillmails = append(normalizedKillmails, killmail)
	}

	s.logger.WithFields(logrus.Fields{
		"normalizedKillmails": len(normalizedKillmails),
	}).Debugln()

	return s.analyzeVictimKillmail(ctx, c, msg, id, normalizedKillmails)

}

func (s *Service) killrightShipCommand(c *cli.Context) error {

	// msg, err := messageFromCLIContext(c)
	// if err != nil {
	// 	return err
	// }

	args := c.Args()

	spew.Dump(args)

	return nil

}

func (s *Service) formatMessage(c *cli.Context, searchedCharacter *esi.CharacterOk, killmail *esi.KillmailOk) string {
	switch c.String("format") {
	case "evelink":
		return fmt.Sprintf(
			"<url=showinfo:1373//%d>%s</url>",
			killmail.Victim.CharacterID,
			killmail.Victim.Character.Name,
		)
	case "detailed":
		return fmt.Sprintf(
			"%s killed %s (%d) on %s in %s (%.2f)",
			searchedCharacter.Name,
			killmail.Victim.Character.Name,
			killmail.KillmailID,
			killmail.KillmailTime.Format("2006-01-02"),
			killmail.SolarSystem.Name,
			killmail.SolarSystem.SecurityStatus,
		)
	default:
		return killmail.Victim.Character.Name
	}
}

func (s *Service) analyzeAgressorKillmail(ctx context.Context, c *cli.Context, msg *discordgo.MessageCreate, targetID uint64, killmails []*esi.KillmailOk) error {

	searchedCharacter, err := s.esi.Character(ctx, targetID)
	if err != nil {
		return errors.Wrap(err, "failed to fetch character from ESI")
	}

	_, err = s.session.ChannelMessageSend(msg.ChannelID, "character found, checking for killmails")
	if err != nil {
		s.logger.WithError(err).Errorln("failed to send message")
	}

	filteredKillmails := make([]*esi.KillmailOk, 0, len(killmails))
	for _, killmail := range killmails {

		var aggressor *esi.KillmailAttacker
		for _, attacker := range killmail.Attackers {
			if uint64(attacker.CharacterID) == targetID {
				aggressor = attacker
				break
			}
		}

		if aggressor == nil {
			continue
		}

		system, err := s.esi.System(ctx, uint(killmail.SolarSystemID))
		if err != nil {
			s.logger.WithError(err).Error("failed to fetch killmail solar system from ESI")
			// return errors.Wrap(err, "failed to fetch killmail solar system from ESI")
			continue
		}
		killmail.SolarSystem = system

		if system.SecurityStatus < 0 {
			continue
		}

		// 670 is a pod/capsule
		if system.SecurityStatus < .5 && killmail.Victim.ShipTypeID != 670 {
			continue
		}

		// Structures don't have a character ID
		if killmail.Victim.CharacterID == 0 {
			continue
		}

		// Add Checking for Wars
		victimCharacter, err := s.esi.Character(ctx, uint64(killmail.Victim.CharacterID))
		if err != nil {
			s.logger.WithError(err).Error("failed to fetch character from ESI")
			// return false, errors.Wrap(err, "failed to fetch character from ESI")
			continue
		}

		killmail.Victim.Character = victimCharacter

		warEntityMatrix := [][]wars.Entity{}
		if aggressor.CorporationID > 0 && killmail.Victim.CorporationID > 0 {
			warEntityMatrix = append(warEntityMatrix, []wars.Entity{
				{T: "corporation", ID: uint(aggressor.CorporationID)},
				{T: "corporation", ID: uint(killmail.Victim.CorporationID)},
			})
		}
		if aggressor.AllianceID > 0 && killmail.Victim.AllianceID > 0 {
			warEntityMatrix = append(warEntityMatrix, []wars.Entity{
				{T: "alliance", ID: aggressor.AllianceID},
				{T: "alliance", ID: killmail.Victim.AllianceID},
			})
		}
		if aggressor.CorporationID > 0 && killmail.Victim.AllianceID > 0 {
			warEntityMatrix = append(warEntityMatrix, []wars.Entity{
				{T: "corporation", ID: aggressor.CorporationID},
				{T: "alliance", ID: killmail.Victim.AllianceID},
			})
		}
		if aggressor.AllianceID > 0 && killmail.Victim.CorporationID > 0 {
			warEntityMatrix = append(warEntityMatrix, []wars.Entity{
				{T: "alliance", ID: aggressor.AllianceID},
				{T: "corporation", ID: killmail.Victim.CorporationID},
			})
		}

		var warSkip = false
		for _, pair := range warEntityMatrix {
			atWar, err := s.wars.EntitiesAtWar(ctx, pair[0], pair[1], killmail.KillmailTime)
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

	if len(filteredKillmails) == 0 {
		_, err := s.session.ChannelMessageSend(msg.ChannelID, appendLatencyToMessageCreate(msg, "0 killmails remained after filtering....", true))
		if err != nil {
			s.logger.WithError(err).Errorln("failed to send message")
		}

		return nil

	}

	_, err = s.session.ChannelMessageSend(msg.ChannelID, fmt.Sprintf("mails filtered down to %d killmails, filtering out duplicate victims....", len(filteredKillmails)))
	if err != nil {
		s.logger.WithError(err).Errorln("failed to send message")
	}

	messages := make([]string, 0, len(filteredKillmails))
	seen := make(map[uint64]bool)
	for _, killmail := range filteredKillmails {
		_, ok := seen[killmail.Victim.CharacterID]
		if ok {
			continue
		}
		messages = append(messages, s.formatMessage(c, searchedCharacter, killmail))
		seen[killmail.Victim.CharacterID] = true
	}

	j := 0
	l := 20
	end := j + l

	var extra string
	switch c.String("format") {
	case "evelink":
		extra = "Copy and Paste this text into the in game notepad. The text will link to the characters showinfo window"
	case "detailed":
		extra = "```<Attacker> killed <Victim> (<Killmail ID>) on <Date> in <System> (<System Sec>)```"
	}

	_, err = s.session.ChannelMessageSend(msg.ChannelID, appendLatencyToMessageCreate(msg, fmt.Sprintf("Found %d potential killrights (Batches of %d):\n%s", len(messages), l, extra), false))
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

func (s *Service) analyzeVictimKillmail(ctx context.Context, c *cli.Context, msg *discordgo.MessageCreate, targetID uint64, killmails []*esi.KillmailOk) error {

	// validAggressors := make([]*esi.KillmailAttacker, 0, len(killmails)*5)
	type character struct {
		seen      int
		character *esi.CharacterOk
	}
	mapAggressors := make(map[uint64]*character)

	entry := s.logger.WithFields(logrus.Fields{
		"perspective": "victim",
		"victimID":    targetID,
	})

	for _, killmail := range killmails {

		entry := entry.WithFields(logrus.Fields{
			"killmailID": killmail.KillmailID,
		})

		// Fetch system killmail took place in
		system, err := s.esi.System(ctx, uint(killmail.SolarSystemID))
		if err != nil {
			return errors.Wrap(err, "failed to fetch killmail solar system from ESI")
		}
		killmail.SolarSystem = system

		// If killmail took place in null sec, ignore it
		if system.SecurityStatus < 0 {
			entry.Debug("skipping due to negative sec status")
			continue
		}

		// If killmail took place in low sec and the victim ship is not a capsule, ignore it
		if system.SecurityStatus < .5 && killmail.Victim.ShipTypeID != 670 {
			entry.Debug("skipping due to non pod kill in low sec")
			continue
		}

		// Since a corporation cannot belong to multiple alliances at once
		// loop over the attackers assumbling a map keyed by the attackers corporation ID with a value of the
		// attacker. This means if a killmail has multiple attackers from the same corporation,
		// we will have one corporationID - allianceID pair to look for wars against,
		// instead of just looking at all aggressors and querying for the same pair over and over.
		var mapUniqueAttackersByCorporationID = make(map[uint]*esi.KillmailAttacker)
		for _, attacker := range killmail.Attackers {
			if _, ok := mapUniqueAttackersByCorporationID[attacker.CorporationID]; !ok {
				mapUniqueAttackersByCorporationID[attacker.CorporationID] = attacker
			}
		}

		victim := killmail.Victim

		// For each of the unique pairs we gathered above, build out a matrix of or statments to assemble
		// a query to mongo with.
		for _, attacker := range mapUniqueAttackersByCorporationID {
			warEntityMatrix := make([][]wars.Entity, 0, 4)
			if victim.CorporationID > 0 && attacker.CorporationID > 0 {
				warEntityMatrix = append(warEntityMatrix, []wars.Entity{
					{T: "corporation", ID: victim.CorporationID},
					{T: "corporation", ID: attacker.CorporationID},
				})
			}
			if victim.AllianceID > 0 && attacker.AllianceID > 0 {
				warEntityMatrix = append(warEntityMatrix, []wars.Entity{
					{T: "alliance", ID: victim.AllianceID},
					{T: "alliance", ID: attacker.AllianceID},
				})
			}
			if victim.CorporationID > 0 && attacker.AllianceID > 0 {
				warEntityMatrix = append(warEntityMatrix, []wars.Entity{
					{T: "corporation", ID: victim.CorporationID},
					{T: "alliance", ID: attacker.AllianceID},
				})
			}
			if victim.AllianceID > 0 && attacker.CorporationID > 0 {
				warEntityMatrix = append(warEntityMatrix, []wars.Entity{
					{T: "alliance", ID: victim.AllianceID},
					{T: "corporation", ID: attacker.CorporationID},
				})
			}

			var warSkip = false
			for _, pair := range warEntityMatrix {
				atWar, err := s.wars.EntitiesAtWar(ctx, pair[0], pair[1], killmail.KillmailTime)
				if err != nil {
					return errors.Wrap(err, "failed to determine if entities are at war")
				}

				warSkip = atWar
				if warSkip {
					break
				}
			}

			if warSkip {
				entry.Debug("skipping due to active war")
				continue
			}

			// If we did not get any wars back, that means that this corporationID/allianceID pair is not at war
			// with the victim, which means the victim has a kill right for each of the attackers that belong to that
			// corporationID/allianceID pair, find those agressors and add them to a basket of *esi.KillmailAttackers
			// Our calling function should convert these to *esi.CharacterOk so that they get a name out of them.
			if !warSkip {
				for _, a := range killmail.Attackers {
					if a.CorporationID == attacker.CorporationID {
						if a.CharacterID == 0 {
							continue
						}
						if _, ok := mapAggressors[a.CharacterID]; !ok {
							mapAggressors[a.CharacterID] = &character{}
						}

						mapAggressors[a.CharacterID].seen++
					}
				}
			}
		}
	}

	if len(mapAggressors) == 0 {
		_, err := s.session.ChannelMessageSend(msg.ChannelID, appendLatencyToMessageCreate(msg, "0 agressors remained after filtering....", true))
		if err != nil {
			s.logger.WithError(err).Errorln("failed to send message")
		}

		return nil
	}

	for characterID, character := range mapAggressors {
		esiChar, err := s.esi.Character(ctx, characterID)
		if err != nil {
			entry.WithError(err).Error("failed to fetch character from ESI")

		}

		character.character = esiChar
	}

	var extra string
	switch c.String("format") {
	case "evelink":
		extra = "Copy and Paste this text into the in game notepad. The text will link to the characters showinfo window\n"
	}

	messages := make([]string, 0, len(mapAggressors))

	for _, character := range mapAggressors {
		var message string
		switch c.String("format") {
		case "evelinks":
			message = fmt.Sprintf(
				"<url=showinfo:1373//%d>%s</url>",
				character.character.ID,
				character.character.Name,
			)
		default:
			message = fmt.Sprintf("%s (%d)", character.character.Name, character.character.ID)
		}

		messages = append(messages, message)

	}

	_, err := s.session.ChannelMessageSend(msg.ChannelID, appendLatencyToMessageCreate(msg, fmt.Sprintf("Found %d potential attacker(s) who this victim may have killrights for:\n%s```%s```", len(messages), extra, strings.Join(messages, "\n")), false))
	if err != nil {
		s.logger.WithError(err).Error("failed to send message")
		return err
	}

	return nil

}
