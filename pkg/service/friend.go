package service

import (
	"awesomeProject10/pkg/domain"
	"context"
)

type FriendService struct {
	f domain.FriendRepository
}

func BuildFriendService(f domain.FriendRepository) FriendService {
	return FriendService{f: f}
}

func (fs FriendService) SetFriend(ctx context.Context, userId int, friendId int) error {
	return fs.f.SetFriend(ctx, userId, friendId)
}

func (fs FriendService) DeleteFriend(ctx context.Context, userId int, friendId int) error {
	return fs.f.DeleteFriend(ctx, userId, friendId)
}
