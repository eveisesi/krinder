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
	Mongo struct {
		Host string `envconfig:"MONGO_HOST" required:"true"`
		User string `envconfig:"MONGO_USER" required:"true"`
		Pass string `envconfig:"MONGO_PASS" required:"true"`
	}
	MySQL struct {
		Host string `required:"true"`
		User string `required:"true"`
		Pass string `required:"true"`
		DB   string `required:"true"`
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
