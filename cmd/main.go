package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"soc/pkg/config"
	"soc/pkg/infrastructure/datastore"
	"soc/pkg/infrastructure/log"
	"soc/pkg/repository"
	"soc/pkg/service"
	"soc/pkg/service/msgbroker"
	"soc/pkg/transport"
	"syscall"
)

const queueName = "posts"

func main() {
	initLogger := log.InitLogger()

	defer func() {
		if r := recover(); r != nil {
			initLogger.Fatal().Msgf("%s", r)
		}
	}()

	cfg, err := config.Load()
	if err != nil {
		initLogger.Fatal().Msgf("Error loading config: %s", err)
	}

	sigsCh := make(chan os.Signal, 1)
	signal.Notify(sigsCh, syscall.SIGINT, syscall.SIGTERM)
	_, cancelCtx := context.WithCancel(context.Background())

	go func() {
		sig := <-sigsCh
		initLogger.Info().Msgf("Got signal: %s", sig)
		cancelCtx()
	}()

	dbPool := datastore.InitDB(cfg.DB, initLogger)

	redisClient := datastore.InitRedis(cfg.Redis, initLogger)

	tarantoolClient := datastore.InitTarantool(cfg.Tarantool, initLogger)

	userRepository := repository.BuildUserRepository(dbPool)
	userService := service.BuildUserService(userRepository)

	friendRepository := repository.BuildFriendRepository(dbPool)
	friendService := service.BuildFriendService(friendRepository)

	postRepository := repository.BuildPostRepository(dbPool)
	postCacheRepository := repository.BuildPostCacheRepository(redisClient)

	dialogueRepository := repository.BuildDialogueRepository(dbPool, tarantoolClient)
	dialogueService := service.BuildDialogueService(dialogueRepository)

	wsHandler := service.NewWsService(redisClient)
	if wsHandler == nil {
		initLogger.Fatal().Msg("Error initializing websocket handler")
	}

	broker := msgbroker.NewMsgBroker(
		"amqp://"+cfg.RabbitMQ.User+":"+cfg.RabbitMQ.Password+"@"+"rabbitmq"+":"+cfg.RabbitMQ.Port+"/",
		wsHandler)
	queue := broker.InitQueue(queueName)

	postService := service.BuildPostService(postRepository, postCacheRepository)

	go broker.RunConsumer(queue.Name, postService)

	server := transport.NewServer(
		dbPool,
		userService,
		friendService,
		postService,
		dialogueService,
		broker,
		wsHandler,
	)

	if err := server.Start(); err != nil {
		fmt.Println("Error starting server: ", err)
		os.Exit(1)
	}
}
