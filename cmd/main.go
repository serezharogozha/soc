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

	userRepository := repository.BuildUserRepository(dbPool)
	userService := service.BuildUserService(userRepository)

	friendRepository := repository.BuildFriendRepository(dbPool)
	friendService := service.BuildFriendService(friendRepository)

	postRepository := repository.BuildPostRepository(dbPool)
	postCacheRepository := repository.BuildPostCacheRepository(redisClient)

	dialogueRepository := repository.BuildDialogueRepository(dbPool)
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

/*func upload(dbPool *pgxpool.Pool) {
	file, err := os.Open("./cmd/people.csv")
	if err != nil {
		panic(err)
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			fmt.Println(err)
		}
	}(file)

	scanner := bufio.NewScanner(file)
	ctx, cancel := context.WithTimeout(context.Background(), 5000*time.Second)
	defer cancel()

	counter := 0

	for scanner.Scan() {
		fields := strings.Split(scanner.Text(), ",")
		firstLastName := strings.Split(fields[0], " ")
		firstName := firstLastName[1]
		lastName := firstLastName[0]

		age := fields[1]
		hometown := fields[2]

		_, err := dbPool.Exec(ctx, "INSERT INTO users (first_name, second_name, birthdate, city) VALUES ($1, $2, $3, $4)",
			firstName, lastName, age, hometown)
		if err != nil {
			fmt.Printf("Failed to insert user: %s\n", err)
		} else {
			counter++
			fmt.Printf("Inserted user %d\n", counter)
		}
	}
}*/
