package datastore

import (
	"counter/pkg/config"
	"github.com/rs/zerolog"
	tarantool "github.com/tarantool/go-tarantool"
)

func InitTarantool(tarantoolCfg config.TarantoolConfig, log *zerolog.Logger) *tarantool.Connection {
	conn, err := tarantool.Connect("tarantool_counter:3302", tarantool.Opts{
		User: "guest",
	})
	if err != nil {
		log.Error().Err(err).Msg("Unable to connect to tarantool")
		panic(err)
	}

	return conn
}
