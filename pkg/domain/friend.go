package domain

import "context"

type Friend struct {
	Id int `json:"id"`
}

type Friends []Friend

type FriendRepository interface {
	SetFriend(ctx context.Context, userId int, friendId int) error
	DeleteFriend(ctx context.Context, userId int, friendId int) error
}
