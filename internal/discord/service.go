package discord

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
	"github.com/eveisesi/krinder/internal/esi"
	"github.com/eveisesi/krinder/internal/universe"
	"github.com/eveisesi/krinder/internal/wars"
	"github.com/eveisesi/krinder/internal/zkillboard"
	"github.com/sirupsen/logrus"
)

// exported for testing
type API interface {
}

type Service struct {
	environment string
	logger      *logrus.Logger

	session *discordgo.Session

	zkb *zkillboard.Service
	esi esi.API

	wars     *wars.Service
	universe *universe.Service

	messages chan *discordgo.MessageCreate
}

func New(token, environment string, logger *logrus.Logger, zkb *zkillboard.Service, esi esi.API, wars *wars.Service, universe *universe.Service) *Service {
	s := &Service{
		environment: environment,
		logger:      logger,

		zkb: zkb,
		esi: esi,

		wars:     wars,
		universe: universe,

		messages: make(chan *discordgo.MessageCreate, 5),
	}

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
