package main

import (
	"context"
	"counter/pkg/config"
	"counter/pkg/infrastructure/datastore"
	"counter/pkg/infrastructure/log"
	"counter/pkg/repository"
	"counter/pkg/service"
	"counter/pkg/service/msgbroker"
	"counter/pkg/transport"
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

	counterRepository := repository.BuildCounterRepository(tarantoolClient)
	counterService := service.BuildCounterService(counterRepository)

	broker := msgbroker.NewMsgBroker(
		"amqp://"+cfg.RabbitMq.User+":"+cfg.RabbitMq.Password+"@"+"rabbitmq"+":"+cfg.RabbitMq.Port+"/", &counterService)
	sendQueue := broker.InitQueue("message_send")
	readQueue := broker.InitQueue("message_read")
	broker.InitQueue("message_send_increment_error")

	go broker.RunSendConsumer(sendQueue.Name)
	go broker.RunReadConsumer(readQueue.Name)

	server := transport.NewServer(
		counterService,
		broker,
	)

	if err := server.Start(); err != nil {
		fmt.Println("Error starting server: ", err)
		os.Exit(1)
	}
}
