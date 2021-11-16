package discord

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/eveisesi/krinder"
	"github.com/eveisesi/krinder/internal/esi"
	"github.com/eveisesi/krinder/internal/store"
	"github.com/pkg/errors"
	"github.com/urfave/cli/v2"
)

var validCategories = []string{"character", "invgroup"}

type searchResult struct {
	id   uint
	name string
}

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

	var results = make([]*searchResult, 0)

	switch category {
	case "character":
		ids, err := s.esi.Search(ctx, category, term, c.Bool("strict"))
		if err != nil {
			return errors.Wrapf(err, "failed to search for %s", category)
		}

		characterNames, err := s.handleSearchCharacterIDs(ctx, ids.Character)
		if err != nil {
			return errors.Wrap(err, "search for characters failed")
		}

		for _, name := range characterNames {
			results = append(results, &searchResult{
				id:   uint(name.ID),
				name: name.Name,
			})
		}
	case "invgroup":
		groupNames, err := s.handleInvGroupSearch(ctx, term, c.Bool("strict"))
		if err != nil {
			return errors.Wrap(err, "search for characters failed")
		}

		for _, name := range groupNames {
			results = append(results, &searchResult{
				id:   uint(name.GroupID),
				name: name.Name,
			})
		}
	}

	if len(results) == 0 {
		_, err := s.session.ChannelMessageSend(msg.ChannelID, "search return 0 results")
		if err != nil {
			s.logger.WithError(err).Error("failed to send message")
			return err
		}
	}

	var contentSlc = make([]string, 0, len(results))
	for _, result := range results {
		contentSlc = append(contentSlc, fmt.Sprintf("%d: %s", result.id, result.name))
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

func (s *Service) handleInvGroupSearch(ctx context.Context, term string, strict bool) ([]*krinder.MySQLGroup, error) {

	var filters = make([]*krinder.Operator, 0)
	filters = append(filters, krinder.NewEqualOperator(store.GroupPublished, 1))

	switch strict {
	case true:
		filters = append(filters, krinder.NewEqualOperator(store.GroupName, term))
	case false:
		filters = append(filters, krinder.NewLikeOperator(store.GroupName, term))
	}

	names, err := s.universe.Groups(ctx, filters...)

	return names, errors.Wrap(err, "Failed to query universe for group")

}
