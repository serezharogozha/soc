package repository

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v4/pgxpool"
)

type FriendRepository struct {
	db    *pgxpool.Pool
	dbRep *pgxpool.Pool
}

func BuildFriendRepository(db *pgxpool.Pool, dbRep *pgxpool.Pool) FriendRepository {
	return FriendRepository{db: db, dbRep: dbRep}
}

func (f FriendRepository) SetFriend(ctx context.Context, userId int, friendId int) error {
	const query = `INSERT INTO friends (user_id, friend_id) VALUES ($1, $2)`
	_, err := f.db.Exec(ctx, query, userId, friendId)
	if err != nil {
		fmt.Println(err)
		return err
	}

	return nil
}

func (f FriendRepository) DeleteFriend(ctx context.Context, userId int, friendId int) error {
	const query = `DELETE FROM friends WHERE user_id= $1 AND friend_id= $2`
	_, err := f.db.Exec(ctx, query, userId, friendId)
	if err != nil {
		fmt.Println(err)
		return err
	}

	return nil
}
