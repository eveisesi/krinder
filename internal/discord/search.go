package discord

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/eveisesi/krinder/internal/esi"
	"github.com/pkg/errors"
	"github.com/urfave/cli/v2"
)

var validCategories = []string{"character"}

func (s *Service) searchCommand(c *cli.Context) error {

	msg, err := messageFromCLIContext(c)
	if err != nil {
		return err
	}

	args := c.Args()
	if args.Len() > 2 {
		return errors.Errorf("expected 2 args, got %d. Surround name in double quotes \"<name>\"", args.Len())
	}

	category := args.Get(0)
	if !isValidCategory(category) {
		return errors.Errorf("%s is an invalid character, expected one of %s", category, strings.Join(validCategories, ", "))
	}

	term := args.Get(1)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	ids, err := s.esi.Search(ctx, category, term, c.Bool("strict"))
	if err != nil {
		return errors.Wrapf(err, "failed to search for %s", category)
	}

	var names []*esi.NamesOk

	switch category {
	case "character":
		characterNames, err := s.handleSearchCharacterIDs(ctx, ids.Character)
		if err != nil {
			return errors.Wrap(err, "search for characters failed")
		}

		names = characterNames
	}

	if len(names) == 0 {
		_, err := s.session.ChannelMessageSend(msg.ChannelID, "search return 0 results")
		if err != nil {
			s.logger.WithError(err).Error("failed to send message")
			return err
		}
	}

	var contentSlc = make([]string, 0, len(names))
	for _, name := range names {
		contentSlc = append(contentSlc, fmt.Sprintf("%d: %s", name.ID, name.Name))
	}

	_, err = s.session.ChannelMessageSend(msg.ChannelID, appendLatencyToMessageCreate(msg, fmt.Sprintf("Search Return %d Results: ```%s```", len(contentSlc), strings.Join(contentSlc, "\n")), false))
	if err != nil {
		s.logger.WithError(err).Error("failed to send message")
		return err
	}

	return nil

}

func appendLatencyToMessageCreate(msg *discordgo.MessageCreate, out string, useNewline bool) string {

	ts, err := msg.Timestamp.Parse()
	if err != nil {
		return out
	}
	latency := time.Until(ts) * -1
	format := "%s_latency_: %v"
	if useNewline {
		format = "%s\n_latency_: %v"
	}
	return fmt.Sprintf(format, out, latency.String())

}

func isValidCategory(category string) bool {
	for _, c := range validCategories {
		if c == category {
			return true
		}
	}

	return false
}

func (s *Service) handleSearchCharacterIDs(ctx context.Context, ids []int) ([]*esi.NamesOk, error) {
	if len(ids) == 0 {
		return nil, nil
	}

	names, err := s.esi.Names(ctx, ids)
	if err != nil {
		return nil, errors.Wrap(err, "failed to search characters name from ids")
	}

	return names, nil

}
