package datastore

import (
	"awesomeProject10/pkg/config"
	"fmt"
	"github.com/go-redis/redis"
	"strconv"
)

func InitRedis(redisCfg config.RedisConf) *redis.Client {

	client := redis.NewClient(&redis.Options{
		Addr: redisCfg.Host + ":" + strconv.FormatInt(redisCfg.Port, 10),
	})

	_, err := client.Ping().Result()
	if err != nil {
		fmt.Println("Error connecting to redis: ", err)
		panic(err)
	}

	return client
}
