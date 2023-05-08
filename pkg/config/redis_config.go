package config

import (
	"fmt"
	"os"
	"strconv"
)

type RedisConf struct {
	Host string
	Port int64
}

func buildRedisConf() (RedisConf, error) {
	dbHost := "redis"

	dbStrPort := os.Getenv("REDIS_PORT")

	if len(dbStrPort) == 0 {
		dbStrPort = "6379"
	}

	dbPort, err := strconv.ParseInt(dbStrPort, 10, 64)
	if err != nil {
		return RedisConf{}, fmt.Errorf("REDIS_PORT is not a number")
	}

	return RedisConf{
		Host: dbHost,
		Port: dbPort,
	}, nil
}
