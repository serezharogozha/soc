package config

type Config struct {
	DB       DbConf
	Redis    RedisConf
	RabbitMQ RabbitConf
}

func Load() (Config, error) {
	dbConf, err := buildDBConfig()

	if err != nil {
		return Config{}, err
	}

	redisConf, err := buildRedisConf()

	rabbitConf, err := buildRabbitConf()

	if err != nil {
		return Config{}, err
	}

	return Config{
		DB:       dbConf,
		Redis:    redisConf,
		RabbitMQ: rabbitConf,
	}, nil

}
