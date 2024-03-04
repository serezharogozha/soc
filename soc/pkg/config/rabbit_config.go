package config

import (
	"fmt"
	"os"
)

type RabbitConf struct {
	User     string
	Password string
	Host     string
	Port     string
}

func buildRabbitConf() (RabbitConf, error) {
	rabbitHost := os.Getenv("RABBIT_HOST")

	if len(rabbitHost) == 0 {
		rabbitHost = "rabbit"
	}

	rabbitPort := os.Getenv("RABBIT_PORT")

	if len(rabbitPort) == 0 {
		rabbitPort = "5672"
	}

	rabbitUser := os.Getenv("RABBIT_USER")

	if len(rabbitUser) == 0 {
		return RabbitConf{}, fmt.Errorf("RABBIT_USER is not set")
	}

	rabbitPassword := os.Getenv("RABBIT_PASSWORD")

	if len(rabbitPassword) == 0 {
		return RabbitConf{}, fmt.Errorf("RABBIT_PASSWORD is not set")
	}

	return RabbitConf{
		Host:     rabbitHost,
		Port:     rabbitPort,
		User:     rabbitUser,
		Password: rabbitPassword,
	}, nil

}
