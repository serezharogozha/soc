package service

import (
	"awesomeProject10/pkg/domain"
	"context"
)

type PostService struct {
	p domain.PostRepository
}

func BuildPostService(p domain.PostRepository) PostService {
	return PostService{p: p}
}

func (ps PostService) CreatePost(ctx context.Context, post domain.Post) error {
	return ps.p.CreatePost(ctx, post)
}

func (ps PostService) UpdatePost(ctx context.Context, post domain.Post) error {
	return ps.p.UpdatePost(ctx, post)
}

func (ps PostService) DeletePost(ctx context.Context, postId int) error {
	return ps.p.DeletePost(ctx, postId)
}

func (ps PostService) GetPost(ctx context.Context, postId int) (*domain.Post, error) {
	return ps.p.GetPost(ctx, postId)
}

func (ps PostService) GetFeed(ctx context.Context, userId int) (*domain.PostFeed, error) {
	return ps.p.GetFeed(ctx, userId)
}
