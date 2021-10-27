package discord

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/bwmarrin/discordgo"
)

func (s *Service) Run() {

	err := s.session.Open()
	if err != nil {
		s.logger.WithError(err).Error("failed to open discord connection")
		return
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go s.handleMessageChannel(ctx)

	s.logger.Info("session initialize successfully, listening for messages")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc

	s.session.Close()

}

func (s *Service) ready(sess *discordgo.Session, event *discordgo.Ready) {
	sess.UpdateGameStatus(0, "Kill! Kill! Kill!")
}

func (s *Service) handleMessageCreate(sess *discordgo.Session, msg *discordgo.MessageCreate) {

	// Ignore our own messages
	if msg.Author.ID == sess.State.User.ID {
		return
	}

	s.messages <- msg

}

func (s *Service) handleMessageChannel(ctx context.Context) {

	for {
		select {
		case msg := <-s.messages:
			s.handleCommand(msg)
		case <-ctx.Done():
			break
		}
	}

}

func (s *Service) handleCommand(msg *discordgo.MessageCreate) {

	for _, command := range s.commands {
		resolver := command.resolver(msg.Content)
		if resolver {
			err := command.executor(msg)
			if err != nil {

				// response to discord channel with error message
				fmt.Println(err)

			}

			break
		}
	}

}
