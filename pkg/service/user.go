package service

import (
	"context"
	"golang.org/x/crypto/bcrypt"
	"soc/pkg/domain"
	"soc/pkg/repository"
)

type UserService struct {
	u repository.UserRepository
}

func BuildUserService(u repository.UserRepository) UserService {
	return UserService{u: u}
}

func HashPassword(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashedPassword), nil
}

func CheckPassword(password string, hashedPassword string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
}

func (us UserService) CreateUser(ctx context.Context, user domain.User) (string, error) {
	password, err := HashPassword(user.Password)
	if err != nil {
		return "", err
	}

	user.Password = password
	userId, err := us.u.InsertUser(ctx, user)
	if err != nil {
		return "", err
	}

	return userId, err
}

func (us UserService) GetUser(ctx context.Context, userId int) (*domain.User, error) {
	return us.u.GetUser(ctx, userId)
}

func (us UserService) SearchUser(ctx context.Context, search domain.Search) ([]domain.UserSafe, error) {
	return us.u.SearchUser(ctx, search)
}
