package config

import (
	"fmt"
	"os"
	"strconv"
)

type TarantoolConfig struct {
	Host string
	Port int64
}

func buildTarantoolConfig() (TarantoolConfig, error) {
	tarantoolHost := os.Getenv("TARANTOOL_HOST")

	if len(tarantoolHost) == 0 {
		tarantoolHost = "tarantool"
	}

	tarantoolStrPort := os.Getenv("TARANTOOL_PORT")

	if len(tarantoolStrPort) == 0 {
		tarantoolStrPort = "3301"
	}

	tarantoolPort, err := strconv.ParseInt(tarantoolStrPort, 10, 64)
	if err != nil {
		return TarantoolConfig{}, fmt.Errorf("TARANTOOL_PORT is not a number")
	}

	return TarantoolConfig{
		Host: tarantoolHost,
		Port: tarantoolPort,
	}, nil
}
