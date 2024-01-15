package datastore

import (
	"context"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/rs/zerolog"
	"soc/pkg/config"
)

func InitDB(dbCfg config.DbConf, log *zerolog.Logger) *pgxpool.Pool {
	cfg, _ := pgxpool.ParseConfig("")
	cfg.ConnConfig.Host = dbCfg.Host
	cfg.ConnConfig.Port = uint16(dbCfg.Port)
	cfg.ConnConfig.User = dbCfg.User
	cfg.ConnConfig.Password = dbCfg.Password
	cfg.ConnConfig.Database = dbCfg.Dbname
	cfg.ConnConfig.PreferSimpleProtocol = true
	cfg.MaxConns = 20

	db, err := pgxpool.ConnectConfig(context.Background(), cfg)
	if err != nil {
		log.Error().Err(err).Msg("Unable to connect to database")
		panic(err)
	}

	return db
}
