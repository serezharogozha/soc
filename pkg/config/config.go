package config

type Config struct {
	DB    DbConf
	DBRep DbConf
}

func Load() (Config, error) {
	dbConf, err := buildDBConfig()
	dbReplicaConf, err := buildDBReplicaConfig()

	if err != nil {
		return Config{}, err
	}

	return Config{
		DB:    dbConf,
		DBRep: dbReplicaConf,
	}, nil

}
