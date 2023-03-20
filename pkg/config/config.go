package config

type Config struct {
	DB DbConf
}

func Load() (Config, error) {
	dbConf, err := buildDBConfig()
	if err != nil {
		return Config{}, err
	}

	return Config{
		DB: dbConf,
	}, nil

}
