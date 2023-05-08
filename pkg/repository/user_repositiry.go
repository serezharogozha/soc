package repository

import (
	"awesomeProject10/pkg/domain"
	"context"
	"fmt"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"strconv"
)

type UserRepository struct {
	db    *pgxpool.Pool
	dbRep *pgxpool.Pool
}

func BuildUserRepository(db *pgxpool.Pool, dbRep *pgxpool.Pool) UserRepository {
	return UserRepository{db: db, dbRep: dbRep}
}

func (u UserRepository) GetUserById(ctx context.Context, userId int) (*domain.User, error) {
	const query = `SELECT * FROM users WHERE id=$1`
	user := new(domain.User)

	row := u.dbRep.QueryRow(ctx, query, userId)
	err := row.Scan(&user.Id, &user.FirstName, &user.SecondName, &user.Birthdate, &user.Biography, &user.City, &user.Password)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}

		return nil, err
	}

	return user, nil
}

func (u UserRepository) InsertUser(ctx context.Context, user domain.User) (string, error) {
	err := u.db.QueryRow(ctx, `
        INSERT INTO users (first_name, second_name, birthdate, biography, city, password)
        VALUES ($1, $2, $3, $4, $5, $6)
        RETURNING id
    `, user.FirstName, user.SecondName, user.Birthdate, user.Biography, user.City, user.Password).Scan(&user.Id)

	if err != nil {
		fmt.Println(err)
		return "", err
	}

	return strconv.Itoa(user.Id), nil
}

func (u UserRepository) GetUser(ctx context.Context, userId int) (*domain.User, error) {
	const query = `SELECT * FROM users WHERE id = $1`

	user := new(domain.User)

	row := u.dbRep.QueryRow(ctx, query, userId)
	err := row.Scan(&user.Id, &user.FirstName, &user.SecondName, &user.Birthdate, &user.Biography, &user.City, &user.Password)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}

		return nil, err
	}

	return user, nil
}

func (u UserRepository) SearchUser(ctx context.Context, search domain.Search) ([]domain.UserSafe, error) {
	const query = `SELECT id, first_name, second_name, birthdate, city FROM users WHERE first_name LIKE $1 and second_name LIKE $2 ORDER BY ID DESC`

	rows, _ := u.db.Query(ctx, query, search.FirstName, search.LastName)
	users := make([]domain.UserSafe, 0)

	for rows.Next() {
		var userSafe domain.UserSafe
		err := rows.Scan(&userSafe.Id, &userSafe.FirstName, &userSafe.SecondName, &userSafe.Birthdate, &userSafe.City)

		if err != nil {
			fmt.Println(err)
			if err == pgx.ErrNoRows {
				return nil, nil
			}
			return nil, err
		}
		users = append(users, userSafe)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return users, nil
}
