package config

type Config struct {
	Tarantool TarantoolConfig
}

func Load() (Config, error) {
	tarantoolConf, err := buildTarantoolConfig()

	if err != nil {
		return Config{}, err
	}

	return Config{
		Tarantool: tarantoolConf,
	}, nil

}
