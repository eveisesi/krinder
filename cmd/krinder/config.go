package main

import (
	"fmt"

	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
)

type config struct {
	Discord struct {
		Token string `envconfig:"DISCORD_TOKEN" required:"true"`
	}
	Log struct {
		Level string `envconfig:"LOG_LEVEL" default:"info"`
	}
	Redis struct {
		Host string `envconfig:"REDIS_HOST" required:"true"`
		Pass string `envconfig:"REDIS_PASS" required:"true"`
	}
	UserAgent string `envconfig:"USER_AGENT" required:"true"`
}

func buildConfig() {
	_ = godotenv.Load(".config/.env")

	cfg = new(config)
	err := envconfig.Process("", cfg)
	if err != nil {
		panic(fmt.Sprintf("failed to config env: %s", err))
	}
}
