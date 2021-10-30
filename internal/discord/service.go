package discord

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
	"github.com/eveisesi/krinder/internal/esi"
	"github.com/eveisesi/krinder/internal/zkillboard"
	"github.com/sirupsen/logrus"
)

// exported for testing
type API interface {
}

type Service struct {
	logger *logrus.Logger

	session *discordgo.Session

	zkb *zkillboard.Service
	esi *esi.Service

	messages chan *discordgo.MessageCreate
}

func New(token string, logger *logrus.Logger, zkb *zkillboard.Service, esi *esi.Service) *Service {
	s := &Service{
		logger: logger,

		zkb: zkb,
		esi: esi,

		messages: make(chan *discordgo.MessageCreate, 5),
	}

	// s.commands = append(s.commands, NewCommand("search", s.searchResolver, s.searchExecutor))
	// s.commands = append(s.commands, NewCommand("ping", s.pingResolver, s.pingExecutor))
	// s.commands = append(s.commands, NewCommand("killright", s.killrightResolver, s.killrightExecutor))

	s.session = s.newDiscordSession(token)

	return s
}

func (s *Service) newDiscordSession(token string) *discordgo.Session {
	dgo, err := discordgo.New(fmt.Sprintf("Bot %s", token))
	if err != nil {
		panic(fmt.Sprintf("failed to initialize discord service: %s", err))
	}

	dgo.AddHandler(s.ready)
	dgo.AddHandler(s.handleMessageCreate)

	return dgo
}
