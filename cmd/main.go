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
	"soc/pkg/transport/ws"
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

	dbPool := datastore.InitDB(cfg.DB, initLogger)
	dbReplicaPool := datastore.InitDB(cfg.DBRep, initLogger)

	redisClient := datastore.InitRedis(cfg.Redis, initLogger)

	userRepository := repository.BuildUserRepository(dbPool, dbReplicaPool)
	userService := service.BuildUserService(userRepository)

	friendRepository := repository.BuildFriendRepository(dbPool, dbReplicaPool)
	friendService := service.BuildFriendService(friendRepository)

	postRepository := repository.BuildPostRepository(dbPool, dbReplicaPool)
	postCacheRepository := repository.BuildPostCacheRepository(redisClient)

	broker := msgbroker.NewMsgBroker("amqp://" + cfg.RabbitMQ.User + ":" + cfg.RabbitMQ.Password + "@" + cfg.RabbitMQ.Host + ":" + cfg.RabbitMQ.Port + "/")
	queueName := "posts"
	queue := broker.InitQueue(queueName)

	go broker.RunConsumer(queue.Name)

	wsHandler := ws.NewWsHandler()
	postService := service.BuildPostService(postRepository, postCacheRepository, wsHandler)

	server := transport.NewServer(
		dbPool,
		userService,
		friendService,
		postService,
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

type RecentCounter struct {
	milliseconds []int
	count        int
}

func Constructor() RecentCounter {
	counter := RecentCounter{
		milliseconds: make([]int, 0),
		count:        0,
	}

	return counter
}

func (this *RecentCounter) Ping(t int) int {
	this.milliseconds = append(this.milliseconds, t)

	for this.milliseconds[0] < t-3000 {
		this.milliseconds = this.milliseconds[1:]
	}

	this.count = len(this.milliseconds)

	return this.count
}
