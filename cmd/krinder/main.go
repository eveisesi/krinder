package main

import (
	"context"

	"github.com/eveisesi/krinder/internal/discord"
	"github.com/eveisesi/krinder/internal/esi"
	"github.com/eveisesi/krinder/internal/zkillboard"
	"github.com/go-redis/redis/v8"
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
	redis := redis.NewClient(&redis.Options{
		Addr:     cfg.Redis.Host,
		Password: cfg.Redis.Pass,
	})

	_, err := redis.Ping(context.TODO()).Result()
	if err != nil {
		logger.WithError(err).Fatal("failed to ping redis")
	}
	zkb := zkillboard.New(cfg.UserAgent)
	esi := esi.New(cfg.UserAgent, redis)

	discord := discord.New(cfg.Discord.Token, logger, zkb, esi)
	discord.Run()

}
