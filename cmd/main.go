package main

import (
	"awesomeProject10/pkg/config"
	"awesomeProject10/pkg/infrastructure/datastore"
	"awesomeProject10/pkg/repository"
	"awesomeProject10/pkg/service"
	"awesomeProject10/pkg/transport"
	"fmt"
	"os"
)

func main() {
	defer func() {
		if r := recover(); r != nil {
			_ = fmt.Sprintf("%s", r)
		}
	}()

	cfg, err := config.Load()
	if err != nil {
		fmt.Println("Error loading config: ", err)
		os.Exit(1)
	}

	dbPool := datastore.InitDB(cfg.DB)

	userRepository := repository.BuildUserRepository(dbPool)
	userService := service.BuildUserService(userRepository)

	server := transport.NewServer(dbPool, userService)

	if err := server.Start(); err != nil {
		fmt.Println("Error starting server: ", err)
		os.Exit(1)
	}

}
