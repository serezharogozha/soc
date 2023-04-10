package main

import (
	"awesomeProject10/pkg/config"
	"awesomeProject10/pkg/infrastructure/datastore"
	"awesomeProject10/pkg/repository"
	"awesomeProject10/pkg/service"
	"awesomeProject10/pkg/transport"
	"bufio"
	"context"
	"fmt"
	"github.com/jackc/pgx/v4/pgxpool"
	"os"
	"strings"
	"time"
)

func main() {
	defer func() {
		if r := recover(); r != nil {
			_ = fmt.Sprintf("%s", r)
		}
	}()

	fmt.Println("config loading")
	cfg, err := config.Load()
	if err != nil {
		fmt.Println("Error loading config: ", err)
		os.Exit(1)
	}

	fmt.Println("init db")
	dbPool := datastore.InitDB(cfg.DB)
	fmt.Println("init replica db")
	dbReplicaPool := datastore.InitDB(cfg.DBRep)

	fmt.Println("build user repository")
	userRepository := repository.BuildUserRepository(dbPool, dbReplicaPool)
	userService := service.BuildUserService(userRepository)

	server := transport.NewServer(dbPool, userService)

	go upload(dbPool)

	if err := server.Start(); err != nil {
		fmt.Println("Error starting server: ", err)
		os.Exit(1)
	}

}
func upload(dbPool *pgxpool.Pool) {
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
}
