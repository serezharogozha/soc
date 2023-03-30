package domain

import "context"

type User struct {
	Id         int    `json:"id"`
	FirstName  string `json:"first_name"`
	SecondName string `json:"second_name"`
	Birthdate  int    `json:"birthdate"`
	Biography  string `json:"biography"`
	City       string `json:"city"`
	Password   string `json:"password"`
}

type Login struct {
	Id       int    `json:"id"`
	Password string `json:"password"`
}

type Search struct {
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
}

type UserSafe struct {
	Id         int    `json:"id"`
	FirstName  string `json:"first_name"`
	SecondName string `json:"second_name"`
	Birthdate  int    `json:"birthdate"`
	City       string `json:"city"`
}

type UserRepository interface {
	InsertUser(ctx context.Context, user User) (string, error)
	GetUser(ctx context.Context, userId int) (*User, error)
	SearchUser(ctx context.Context, search Search) ([]UserSafe, error)
}
