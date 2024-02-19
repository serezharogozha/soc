package config

type Config struct {
	DB        DbConf
	Redis     RedisConf
	RabbitMQ  RabbitConf
	Tarantool TarantoolConfig
}

func Load() (Config, error) {
	dbConf, err := buildDBConfig()

	if err != nil {
		return Config{}, err
	}

	redisConf, err := buildRedisConf()

	rabbitConf, err := buildRabbitConf()

	tarantoolConf, err := buildTarantoolConfig()

	if err != nil {
		return Config{}, err
	}

	return Config{
		DB:        dbConf,
		Redis:     redisConf,
		RabbitMQ:  rabbitConf,
		Tarantool: tarantoolConf,
	}, nil

}
