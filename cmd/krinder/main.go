package main

import (
	"github.com/eveisesi/krinder/internal/discord"
	"github.com/eveisesi/krinder/internal/esi"
	"github.com/eveisesi/krinder/internal/zkillboard"
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
	zkb := zkillboard.New(cfg.UserAgent)
	esi := esi.New(cfg.UserAgent)

	discord := discord.New(cfg.Discord.Token, logger, zkb, esi)
	discord.Run()

}
