package discord

import (
	"context"
	"fmt"
	"sync"

	"github.com/bwmarrin/discordgo"
	"github.com/kballard/go-shellquote"
	"github.com/pkg/errors"
	"github.com/urfave/cli/v2"
)

var ErrMissingMsgIFace = errors.New("[ErrMissingMsgIFace] msg missing from metadata. contact application maintainer")
var ErrMsgIFaceTypeInvalid = errors.New("[ErrMsgIFaceTypeInvalid] invalid type for msg iface. contact application mainter")

func (s *Service) Run(done chan bool, wg *sync.WaitGroup) {

	defer wg.Done()

	err := s.session.Open()
	if err != nil {
		s.logger.WithError(err).Error("failed to open discord connection")
		return
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go s.handleMessageChannel(ctx)

	s.logger.Info("session initialize successfully, listening for messages")
	<-done
	s.logger.WithField("service", "discord").Info("hold channel received value, closing session")

	s.session.Close()

}

func (s *Service) ready(sess *discordgo.Session, event *discordgo.Ready) {
	err := sess.UpdateGameStatus(0, "Kill! Kill! Kill!")
	if err != nil {
		s.logger.WithError(err).Error("failed to set bot status")
	}
}

func (s *Service) handleMessageCreate(sess *discordgo.Session, msg *discordgo.MessageCreate) {

	// Ignore our own messages
	if msg.Author.ID == sess.State.User.ID {
		return
	}

	if s.environment == "production" {
		channel, err := s.session.Channel(msg.ChannelID)
		if err != nil {
			s.logger.WithError(err).Error("failed to fetch channel infomation")
			return
		}

		if channel.Type != discordgo.ChannelTypeDM {
			return
		}
	}

	s.messages <- msg

}

func (s *Service) handleMessageChannel(ctx context.Context) {

	for {
		select {
		case msg := <-s.messages:
			err := s.handleCommand(msg)
			if err != nil {
				_, err := s.session.ChannelMessageSend(msg.ChannelID, fmt.Sprintf("Your request encountered an error. Please try again in a few seconds, if the error continues, contact the Bot Maintainer\n%s", err))
				if err != nil {
					s.logger.WithError(err).Error("failed to send message to discord")
				}
			}
		case <-ctx.Done():
			break
		}
	}

}

func (s *Service) handleCommand(msg *discordgo.MessageCreate) error {

	defer buf.Reset()

	words, err := shellquote.Split(msg.Content)
	if err != nil {
		s.logger.WithError(err).Error("failed to parse inputted command")
		return nil
	}

	app := s.initializeCLI()

	if !s.shouldRunCLI(app, words) {
		return nil
	}

	words = append([]string{"krinder"}, words...)

	app.Metadata["msg"] = msg

	err = app.Run(words)
	if err != nil {
		return err
	}

	if buf.String() == "" {
		return nil
	}

	_, err = s.session.ChannelMessageSend(msg.ChannelID, buf.String())
	if err != nil {
		return errors.Wrap(err, "failed to send channel message")
	}

	return nil

}

func messageFromCLIContext(c *cli.Context) (*discordgo.MessageCreate, error) {
	msgIface, ok := c.App.Metadata["msg"]
	if !ok {
		return nil, ErrMissingMsgIFace
	}

	msg, ok := msgIface.(*discordgo.MessageCreate)
	if !ok {
		return nil, ErrMsgIFaceTypeInvalid
	}

	return msg, nil

}
