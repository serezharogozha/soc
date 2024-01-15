package service

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/rs/zerolog/log"
	"soc/pkg/domain"
	"soc/pkg/repository"
	"soc/pkg/transport/ws"
	"strconv"
)

type PostService struct {
	postRepository      repository.PostRepository
	postCacheRepository repository.PostCacheRepository
	wsHandler           *ws.WsHandler
}

func BuildPostService(postRepository repository.PostRepository, postCacheRepository repository.PostCacheRepository, wsHandler *ws.WsHandler) PostService {
	return PostService{
		postRepository:      postRepository,
		postCacheRepository: postCacheRepository,
		wsHandler:           wsHandler,
	}
}

func (ps PostService) GetFriendToPublish(post domain.Post) (domain.Friends, error) {
	friendsOfUser, err := ps.postRepository.GetFriendsOfUser(post.UserId)

	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	return friendsOfUser, nil
}

func (ps PostService) PublishPostToCache(post domain.Post, friendsOfUser domain.Friends) error {
	for _, friend := range friendsOfUser {
		UserIdStr := strconv.FormatInt(int64(friend.Id), 10)

		postJson, err := json.Marshal(post)
		if err != nil {
			fmt.Println(err)
			return err
		}

		err = ps.postCacheRepository.Add("feed:"+UserIdStr, string(postJson))
		if err != nil {
			fmt.Println("Failed to add post to cache")
			fmt.Println(err)
			return err
		}
		ps.PublishToWs(UserIdStr, string(postJson))
	}
	return nil
}

func (ps PostService) PublishToWs(UserIdStr string, postJson string) {
	topic := "feed:" + UserIdStr
	//TODO
	err := ps.wsHandler.Publish(topic, []byte(postJson))
	if err != nil {
		fmt.Println(err)
	}
}

func (ps PostService) CreatePost(ctx context.Context, post domain.Post) error {
	return ps.postRepository.CreatePost(ctx, post)
}

func (ps PostService) UpdatePost(ctx context.Context, post domain.Post) error {
	return ps.postRepository.UpdatePost(ctx, post)
}

func (ps PostService) DeletePost(ctx context.Context, postId int) error {
	return ps.postRepository.DeletePost(ctx, postId)
}

func (ps PostService) GetPost(ctx context.Context, postId int) (*domain.Post, error) {
	return ps.postRepository.GetPost(ctx, postId)
}

func (ps PostService) GetFeed(ctx context.Context, userId int) (*domain.PostFeed, error) {
	result := &domain.PostFeed{}

	UserIdStr := strconv.FormatInt(int64(userId), 10)
	cachedFeed, err := ps.postCacheRepository.GetFeed(UserIdStr)
	if err != nil {
		feedFromDb, err := ps.postRepository.GetFeed(ctx, userId)
		if err != nil {
			log.Log().Err(err).Msg("error getting feed from db")
		}

		result.Posts = append(result.Posts, feedFromDb.Posts...)
		err = ps.postCacheRepository.AddFeed(UserIdStr, cachedFeed)
	} else {
		result.Posts = append(result.Posts, cachedFeed.Posts...)
	}

	return result, nil
}
