package service

import (
	"awesomeProject10/pkg/domain"
	"context"
	"fmt"
)

type UserService struct {
	u domain.UserRepository
}

func BuildUserService(u domain.UserRepository) UserService {
	return UserService{u: u}
}

func (us UserService) CreateUser(ctx context.Context, user domain.User) (string, error) {
	userId, err := us.u.InsertUser(ctx, user)
	if err != nil {
		fmt.Println(err)
		return "", err
	}

	return userId, err
}

func (us UserService) GetUser(ctx context.Context, userId int) (*domain.User, error) {
	return us.u.GetUser(ctx, userId)
}
