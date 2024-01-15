package service

import (
	"context"
	"soc/pkg/repository"
)

type FriendService struct {
	friendRepository repository.FriendRepository
}

func BuildFriendService(friendRepository repository.FriendRepository) FriendService {
	return FriendService{
		friendRepository: friendRepository,
	}
}

func (fs FriendService) SetFriend(ctx context.Context, userId int, friendId int) error {
	return fs.friendRepository.SetFriend(ctx, userId, friendId)
}

func (fs FriendService) DeleteFriend(ctx context.Context, userId int, friendId int) error {
	return fs.friendRepository.DeleteFriend(ctx, userId, friendId)
}
