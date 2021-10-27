package discord

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/eveisesi/krinder/internal/esi"
	"github.com/pkg/errors"
)

func (s *Service) searchResolver(msg string) bool {
	return strings.HasPrefix(msg, "search")
}

func (s *Service) searchExecutor(msg *discordgo.MessageCreate) error {

	// msg.Content == "search <name>"; stripped = "<name>"
	stripped := strings.TrimPrefix(msg.Content, "search ")
	category := "character"

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	ids, err := s.esi.Search(ctx, "character", stripped)
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

	_, err = s.session.ChannelMessageSend(msg.ChannelID, fmt.Sprintf("Search Return %d Results: ```%s```", len(contentSlc), strings.Join(contentSlc, "\n")))
	if err != nil {
		s.logger.WithError(err).Error("failed to send message")
		return err
	}

	return nil

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
