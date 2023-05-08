package service

import (
	"github.com/go-redis/redis"
	"time"
)

type Last1000Cache struct {
	client *redis.Client
}

func NewCache(client *redis.Client) *Last1000Cache {
	return &Last1000Cache{
		client: client,
	}
}

func (c *Last1000Cache) Add(key string, value string) error {
	err := c.client.ZAdd(key, redis.Z{
		Score:  float64(time.Now().Unix()),
		Member: value,
	}).Err()
	if err != nil {
		return err
	}

	err = c.client.ZRemRangeByRank(key, 0, -1001).Err()
	if err != nil {
		return err
	}

	return nil
}

func (c *Last1000Cache) Get(key string) ([]string, error) {
	values, err := c.client.ZRevRange(key, 0, 999).Result()
	if err != nil {
		panic(err)
	}

	return values, nil
}
