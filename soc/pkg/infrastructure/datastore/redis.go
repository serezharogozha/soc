package datastore

import (
	"github.com/go-redis/redis"
	"github.com/rs/zerolog"
	"soc/pkg/config"
	"strconv"
)

func InitRedis(redisCfg config.RedisConf, log *zerolog.Logger) *redis.Client {

	client := redis.NewClient(&redis.Options{
		Addr: redisCfg.Host + ":" + strconv.FormatInt(redisCfg.Port, 10),
	})

	_, err := client.Ping().Result()
	if err != nil {
		log.Error().Err(err).Msg("Unable to connect to redis")
		panic(err)
	}

	return client
}
