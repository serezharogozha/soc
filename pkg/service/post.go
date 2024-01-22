package service

import (
	"context"
	"encoding/json"
	"github.com/rs/zerolog/log"
	"soc/pkg/domain"
	"soc/pkg/repository"
	"strconv"
)

type PostService struct {
	postRepository      repository.PostRepository
	postCacheRepository repository.PostCacheRepository
	wsHandler           *WsService
}

func BuildPostService(postRepository repository.PostRepository, postCacheRepository repository.PostCacheRepository) PostService {
	return PostService{
		postRepository:      postRepository,
		postCacheRepository: postCacheRepository,
	}
}

func (ps PostService) GetFriendToPublish(post domain.Post) (domain.Friends, error) {
	friendsOfUser, err := ps.postRepository.GetFriendsOfUser(post.UserId)

	if err != nil {
		return nil, err
	}

	return friendsOfUser, nil
}

func (ps PostService) PublishPostToCache(post domain.Post, friendsOfUser domain.Friends) error {
	for _, friend := range friendsOfUser {
		UserIdStr := strconv.FormatInt(int64(friend.Id), 10)

		postJson, err := json.Marshal(post)
		if err != nil {
			return err
		}

		err = ps.postCacheRepository.Add("feed:"+UserIdStr, string(postJson))
		if err != nil {
			return err
		}

	}
	return nil
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
