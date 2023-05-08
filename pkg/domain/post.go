package domain

import "context"

type Post struct {
	Id     int    `json:"id"`
	Text   string `json:"text"`
	UserId int    `json:"user_id"`
}

type Posts []Post

type PostFeed struct {
	Posts `json:"posts"`
}

type PostRepository interface {
	CreatePost(ctx context.Context, post Post) error
	UpdatePost(ctx context.Context, post Post) error
	DeletePost(ctx context.Context, postId int) error
	GetPost(ctx context.Context, postId int) (*Post, error)
	GetFeed(ctx context.Context, userId int) (*PostFeed, error)
}
