package repository

import (
	"encoding/json"
	"fmt"
	"github.com/go-redis/redis"
	"soc/pkg/domain"
	"time"
)

type PostCacheRepository struct {
	redisClient *redis.Client
}

func BuildPostCacheRepository(redisClient *redis.Client) PostCacheRepository {
	return PostCacheRepository{redisClient: redisClient}
}

func (p PostCacheRepository) GetFeed(userId string) (*domain.PostFeed, error) {
	cachedFeed, err := p.Get("feed:" + userId)
	if err != nil {
		return nil, err
	}
	//TODO check errors
	postFeed := domain.PostFeed{}

	for _, cachedPost := range cachedFeed {
		post := new(domain.Post)
		err := json.Unmarshal([]byte(cachedPost), &post)
		if err != nil {

			fmt.Println(err)
			return nil, err
		}

		fmt.Println(post)
		postFeed.Posts = append(postFeed.Posts, *post)
	}

	return &postFeed, nil
}

func (p PostCacheRepository) AddFeed(userId string, postFeed *domain.PostFeed) error {
	for _, postDb := range postFeed.Posts {
		postJson, err := json.Marshal(postDb)
		err = p.Add("feed:"+userId, string(postJson))
		if err != nil {
			return err
		}
	}
	return nil
}

func (p PostCacheRepository) Add(key string, value string) error {
	err := p.redisClient.ZAdd(key, redis.Z{
		Score:  float64(time.Now().Unix()),
		Member: value,
	}).Err()
	if err != nil {
		return err
	}

	err = p.redisClient.ZRemRangeByRank(key, 0, -1001).Err()
	if err != nil {
		return err
	}

	return nil
}

func (p PostCacheRepository) Get(key string) ([]string, error) {
	values, err := p.redisClient.ZRevRange(key, 0, 999).Result()
	if err != nil {
		panic(err)
	}

	return values, nil
}
