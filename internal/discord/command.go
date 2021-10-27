package discord

import (
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/pkg/errors"
)

type commandResolverFunc func(msg string) bool
type commandExecutorFunc func(msg *discordgo.MessageCreate) error

type Command struct {
	name     string
	resolver commandResolverFunc
	executor commandExecutorFunc
}

func NewCommand(name string, resolver commandResolverFunc, executor commandExecutorFunc) *Command {
	return &Command{
		name:     name,
		resolver: resolver,
		executor: executor,
	}
}

func (s *Service) pingResolver(msg string) bool {
	return strings.HasPrefix(msg, "ping")
}

func (s *Service) pingExecutor(msg *discordgo.MessageCreate) error {

	_, err := s.session.ChannelMessageSend(msg.ChannelID, "pong!")
	if err != nil {
		return errors.Wrap(err, "failed to send channel message")
	}

	return nil

}
