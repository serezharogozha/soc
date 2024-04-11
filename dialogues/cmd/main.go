package main

import (
	"context"
	"dialogues/pkg/config"
	"dialogues/pkg/infrastructure/datastore"
	"dialogues/pkg/infrastructure/log"
	"dialogues/pkg/repository"
	"dialogues/pkg/service"
	"dialogues/pkg/service/msgbroker"
	"dialogues/pkg/transport"
	"fmt"
	"os"
	"os/signal"
	"syscall"
)

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

	tarantoolClient := datastore.InitTarantool(cfg.Tarantool, initLogger)

	dialogueRepository := repository.BuildDialogueRepository(tarantoolClient)
	dialogueService := service.BuildDialogueService(dialogueRepository)

	broker := msgbroker.NewMsgBroker(
		"amqp://"+cfg.RabbitMq.User+":"+cfg.RabbitMq.Password+"@"+"rabbitmq"+":"+cfg.RabbitMq.Port+"/", &dialogueService)
	_ = broker.InitQueue("message_send")
	_ = broker.InitQueue("message_read")
	messageSentIncrementErrorQueue := broker.InitQueue("message_read_decrement_error")

	go broker.RunErrorIncrementConsumer(messageSentIncrementErrorQueue.Name)

	metrics := transport.InitMetrics(initLogger)

	server := transport.NewServer(
		dialogueService,
		broker,
		metrics,
	)

	if err := server.Start(); err != nil {
		fmt.Println("Error starting server: ", err)
		os.Exit(1)
	}
}
