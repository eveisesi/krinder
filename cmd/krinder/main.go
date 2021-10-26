package main

import (
	"github.com/eveisesi/krinder/internal/discord"
	"github.com/sirupsen/logrus"
)

var (
	cfg    *config
	logger *logrus.Logger
)

func init() {
	buildConfig()
	buildLogger()
}

func main() {

	discord := discord.New(cfg.Discord.Token, logger)

	discord.Run()

}
