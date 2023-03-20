package datastore

import (
	"awesomeProject10/pkg/config"
	"context"
	"github.com/jackc/pgx/v4/pgxpool"
)

func InitDB(dbCfg config.DbConf) *pgxpool.Pool {
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
		panic(err)
	}

	return db
}
