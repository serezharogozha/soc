package config

type Config struct {
	Tarantool TarantoolConfig
	RabbitMq  RabbitConf
}

func Load() (Config, error) {
	tarantoolConf, err := buildTarantoolConfig()

	if err != nil {
		return Config{}, err
	}

	rabbitConf, err := buildRabbitConf()
	if err != nil {
		return Config{}, err
	}

	return Config{
		Tarantool: tarantoolConf,
		RabbitMq:  rabbitConf,
	}, nil

}
