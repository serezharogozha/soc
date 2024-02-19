package config

import (
	"fmt"
	"os"
	"strconv"
)

type DbConf struct {
	Host     string
	Port     int64
	User     string
	Password string
	Dbname   string
}

func buildDBConfig() (DbConf, error) {
	dbHost := os.Getenv("DB_HOST")

	if len(dbHost) == 0 {
		dbHost = "localhost"
	}

	dbStrPort := os.Getenv("DB_PORT")

	if len(dbStrPort) == 0 {
		dbStrPort = "5432"
	}

	dbPort, err := strconv.ParseInt(dbStrPort, 10, 64)
	if err != nil {
		return DbConf{}, fmt.Errorf("DB_PORT is not a number")
	}

	dbUser := os.Getenv("DB_USER")

	if len(dbUser) == 0 {
		return DbConf{}, fmt.Errorf("DB_USER is not set")
	}

	dbPassword := os.Getenv("DB_PASSWORD")

	dbName := os.Getenv("DB_NAME")

	if len(dbName) == 0 {
		return DbConf{}, fmt.Errorf("DB_NAME is not set")
	}

	fmt.Println(DbConf{
		Host:     dbHost,
		Port:     dbPort,
		User:     dbUser,
		Password: dbPassword,
		Dbname:   dbName,
	})

	return DbConf{
		Host:     dbHost,
		Port:     dbPort,
		User:     dbUser,
		Password: dbPassword,
		Dbname:   dbName,
	}, nil

}
