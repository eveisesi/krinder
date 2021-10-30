package main

import (
	"context"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/eveisesi/krinder/internal/discord"
	"github.com/eveisesi/krinder/internal/esi"
	mdb "github.com/eveisesi/krinder/internal/mongo"
	"github.com/eveisesi/krinder/internal/wars"
	"github.com/eveisesi/krinder/internal/zkillboard"
	"github.com/go-redis/redis/v8"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	cfg    *config
	logger *logrus.Logger
	wg     = new(sync.WaitGroup)
)

func init() {
	buildConfig()
	buildLogger()
}

func main() {
	// Initialize context with a timeout. This will prevent elongated issues with
	// dependency connections. We have 5 seconds to establish a connection or we die
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)

	// Builds a redis client
	redis := buildRedis(ctx)

	// Builds a mongo client
	mongo := buildMongo(ctx)

	// Call cancel func. We are done with that context and can release the resources
	cancel()

	warsRepo, err := mdb.NewWarRepository(mongo.Database("krinder"))
	if err != nil {
		logger.WithError(err).Error("failed to initialize wars Repository")
	}

	// Build out the services we want to use
	zkb := zkillboard.New(cfg.UserAgent)
	esi := esi.New(cfg.UserAgent, redis)

	done := make(chan bool, 1)

	wars := wars.NewService(logger, esi, warsRepo)

	wars.Initialize()

	// The discord service is the root service of this application.
	// It maintains a connection to the Discord Gateway and processes all commands
	// that users may issue via that gateway
	wg.Add(1)
	go discord.New(cfg.Discord.Token, logger, zkb, esi).Run(done, wg)

	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc

	for i := 0; i < 2; i++ {
		done <- true
	}

	wg.Wait()

}

func buildRedis(ctx context.Context) *redis.Client {
	redis := redis.NewClient(&redis.Options{
		Addr:     cfg.Redis.Host,
		Password: cfg.Redis.Pass,
	})

	_, err := redis.Ping(ctx).Result()
	if err != nil {
		logger.WithError(err).Fatal("failed to ping redis")
	}
	return redis
}

func buildMongo(ctx context.Context) *mongo.Client {

	clientOpts := options.Client()
	clientOpts.SetAppName("krinder-discord-bot")
	clientOpts.SetHosts([]string{cfg.Mongo.Host})
	clientOpts.SetAuth(options.Credential{
		AuthMechanism: "SCRAM-SHA-256",
		Username:      cfg.Mongo.User,
		Password:      cfg.Mongo.Pass,
	})

	client, err := mongo.Connect(ctx, clientOpts)
	if err != nil {
		logger.WithError(err).Error("failed to connect to mongo db")
	}

	err = client.Ping(ctx, nil)
	if err != nil {
		logger.WithError(err).Error("failed to ping mongo db")
	}

	return client
}
