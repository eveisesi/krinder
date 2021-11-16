package main

import (
	"context"
	"database/sql"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/eveisesi/krinder/internal/discord"
	"github.com/eveisesi/krinder/internal/esi"
	"github.com/eveisesi/krinder/internal/store"
	"github.com/eveisesi/krinder/internal/universe"
	"github.com/eveisesi/krinder/internal/wars"
	"github.com/eveisesi/krinder/internal/zkillboard"
	"github.com/go-redis/redis/v8"
	mysqlDriver "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"github.com/robfig/cron/v3"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/mongo"
	mongoDriver "go.mongodb.org/mongo-driver/mongo"
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
	mongoConn := buildMongo(ctx)

	mysqlConn := buildMySQL()

	// Call cancel func. We are done with that context and can release the resources
	cancel()

	warsRepo, err := store.NewWarRepository(mongoConn.Database(cfg.Mongo.Database))
	if err != nil {
		logger.WithError(err).Fatal("failed to initialize wars repository")
	}

	universeRepo, err := store.NewUniverseRepository(mysqlConn, mongoConn.Database(cfg.Mongo.Database))
	if err != nil {
		logger.WithError(err).Fatal("failed to initialize universe repository")
	}

	// Build out the services we want to use
	zkb := zkillboard.New(cfg.UserAgent)
	esi := esi.New(cfg.UserAgent, redis)
	wars := wars.NewService(logger, esi, warsRepo)
	wars.Run()

	universe := universe.New(logger, redis, esi, universeRepo)
	universe.Run()

	cn := cron.New()
	_, err = cn.AddJob("@every 3h", wars)
	if err != nil {
		logger.WithError(err).Fatal("failed to add job to cron scheduler")
	}
	_, err = cn.AddJob("@every 24h", universe)
	if err != nil {
		logger.WithError(err).Fatal("failed to add job to cron scheduler")
	}
	done := make(chan bool, 1)
	wg.Add(1)
	go func(cn *cron.Cron, done chan bool, wg *sync.WaitGroup) {
		defer wg.Done()
		cn.Start()
		<-done
		entry := logger.WithField("service", "cron")
		entry.Info("hold channel received value, closing session")

		entry.Info("waiting for context to be marked done")
		<-cn.Stop().Done()
		entry.Info("context is done, terminating go routine")

	}(cn, done, wg)

	// The discord service is the root service of this application.
	// It maintains a connection to the Discord Gateway and processes all commands
	// that users may issue via that gateway
	wg.Add(1)
	go discord.New(cfg.Discord.Token, cfg.Environment, logger, zkb, esi, wars, universe).Run(done, wg)

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

	client, err := mongoDriver.Connect(ctx, clientOpts)
	if err != nil {
		logger.WithError(err).Error("failed to connect to mongo db")
	}

	err = client.Ping(ctx, nil)
	if err != nil {
		logger.WithError(err).Error("failed to ping mongo db")
	}

	return client
}

func buildMySQL() *sqlx.DB {

	m := cfg.MySQL

	config := mysqlDriver.Config{
		User:                 m.User,
		Passwd:               m.Pass,
		Net:                  "tcp",
		Addr:                 m.Host,
		DBName:               m.DB,
		Loc:                  time.UTC,
		Timeout:              time.Second,
		ReadTimeout:          time.Second,
		WriteTimeout:         time.Second,
		ParseTime:            true,
		AllowNativePasswords: true,
	}

	db, err := sql.Open("mysql", config.FormatDSN())
	if err != nil {
		log.Panicf("[MySQL Connect] Failed to connect to mysql server: %s", err)
	}

	db.SetConnMaxIdleTime(time.Second * 5)
	db.SetConnMaxLifetime(time.Second * 30)
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(15)

	err = db.Ping()
	if err != nil {
		log.Panicf("[MySQL Connect] Failed to ping mysql server: %s", err)
	}

	return sqlx.NewDb(db, "mysql")

}
