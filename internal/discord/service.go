package discord

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/bwmarrin/discordgo"
	"github.com/sirupsen/logrus"
)

// exported for testing
type API interface {
}

type Service struct {
	token  string
	logger *logrus.Logger
	// discord *discordgo.Session
}

func New(token string, logger *logrus.Logger) *Service {
	return &Service{
		logger: logger,
		token:  token,
	}
}

func newDiscordSession(token string) *discordgo.Session {
	dgo, err := discordgo.New(fmt.Sprintf("Bot %s", token))
	if err != nil {
		panic(fmt.Sprintf("failed to initialize discord service: %s", err))
	}
	return dgo
}

func (s *Service) Run() {

	sess := newDiscordSession(s.token)
	sess.AddHandler(s.ready)
	sess.AddHandler(s.handleMessageCreate)

	err := sess.Open()
	if err != nil {
		s.logger.WithError(err).Error("failed to open discord connection")
		return
	}

	s.logger.Info("session initialize successfully, listening for messages")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc

	sess.Close()

}

func (s *Service) ready(sess *discordgo.Session, event *discordgo.Ready) {
	sess.UpdateGameStatus(0, "Kill! Kill! Kill!")
}

func (s *Service) handleMessageCreate(sess *discordgo.Session, msg *discordgo.MessageCreate) {

	// Ignore our own messages
	if msg.Author.ID == sess.State.User.ID {
		return
	}

	// This is ping pong, ignore anything else
	if msg.Content != "ping" {
		return
	}

	channel, err := sess.Channel(msg.ChannelID)
	if err != nil {
		s.logger.WithError(err).Error("failed to lookup channel by id")
	}

	if channel.Type != discordgo.ChannelTypeDM {
		// we only deal with dm's, ignore all others
		return
	}

	_, err = sess.ChannelMessageSend(msg.ChannelID, "pong!")
	if err != nil {
		s.logger.WithError(err).Error("failed to send message")
	}
}
