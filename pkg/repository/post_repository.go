package repository

import (
	"awesomeProject10/pkg/domain"
	"awesomeProject10/pkg/service"
	"context"
	"encoding/json"
	"fmt"
	"github.com/go-redis/redis"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/streadway/amqp"
	"strconv"
)

type PostRepository struct {
	db    *pgxpool.Pool
	dbRep *pgxpool.Pool
	Cache *service.Last1000Cache
	redis *redis.Client
	ch    *amqp.Channel
}

func BuildPostRepository(db *pgxpool.Pool, dbRep *pgxpool.Pool, redis *redis.Client, cache *service.Last1000Cache, ch *amqp.Channel) PostRepository {
	return PostRepository{db: db, dbRep: dbRep, redis: redis, Cache: cache, ch: ch}
}

func (p PostRepository) CreatePost(ctx context.Context, post domain.Post) error {
	const query = `INSERT INTO posts (text, user_id) VALUES ($1, $2) RETURNING id`
	err := p.db.QueryRow(ctx, query, post.Text, post.UserId).Scan(&post.Id)
	if err != nil {
		fmt.Println(err)
		return err
	}

	body, err := json.Marshal(post)
	if err != nil {
		return err
	}

	err = p.ch.Publish(
		"",
		"posts",
		false,
		false,
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        body,
		},
	)
	fmt.Println("Successfully Published Message to Queue")

	if err != nil {
		fmt.Println(err)
	}

	return nil
}

func (p PostRepository) UpdatePost(ctx context.Context, post domain.Post) error {
	const query = `UPDATE posts SET text = $1 and user_id = $2 where id = $3`
	_, err := p.db.Exec(ctx, query, post.Text, post.UserId, post.Id)
	if err != nil {
		fmt.Println(err)
		return err
	}

	return nil
}

func (p PostRepository) DeletePost(ctx context.Context, postId int) error {
	const query = `DELETE FROM posts WHERE id =$ 1`
	_, err := p.db.Exec(ctx, query, postId)
	if err != nil {
		fmt.Println(err)
		return err
	}

	return nil
}

func (p PostRepository) GetPost(ctx context.Context, postId int) (*domain.Post, error) {
	const query = `SELECT * FROM posts WHERE id = $1`
	post := new(domain.Post)

	row := p.db.QueryRow(ctx, query, postId)
	err := row.Scan(&post.Id, &post.Text, &post.UserId)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}

		return nil, err
	}

	return post, nil
}

func (p PostRepository) GetFeed(ctx context.Context, userId int) (*domain.PostFeed, error) {
	userIdStr := strconv.FormatInt(int64(userId), 10)
	postFeed := domain.PostFeed{}

	cachedFeed, err := p.Cache.Get("feed:" + userIdStr)

	if err != nil {
		fmt.Println(err)
		const query = `SELECT * FROM posts LEFT JOIN friends ON posts.user_id = friends.friend_id WHERE friends.user_id = $1 ORDER BY posts.id DESC LIMIT 1000`
		dbFeed, err := p.db.Query(ctx, query, userId)

		if err != nil {
			fmt.Println(err)
			return nil, err
		}

		err = dbFeed.Scan(&postFeed.Posts)
		if err != nil {
			return nil, err
		}

		for _, postDb := range postFeed.Posts {
			postJson, err := json.Marshal(postDb)
			err = p.Cache.Add("feed:"+userIdStr, string(postJson))
			if err != nil {
				return nil, err
			}
		}
		return &postFeed, nil
	}

	for _, cachedPost := range cachedFeed {
		post := new(domain.Post)
		err := json.Unmarshal([]byte(cachedPost), &post)
		if err != nil {

			fmt.Println(err)
			return nil, err
		}

		fmt.Println(post)
		postFeed.Posts = append(postFeed.Posts, *post)
	}

	return &postFeed, nil
}

func (p PostRepository) GetFriendsOfUser(userId int) (domain.Friends, error) {
	friends := domain.Friends{}

	const query = `SELECT user_id FROM friends WHERE friend_id = $1`
	ctx := context.Context(context.Background())
	rows, err := p.db.Query(ctx, query, userId)
	if err != nil {
		return friends, err
	}

	for rows.Next() {
		friend := new(domain.Friend)
		err := rows.Scan(&friend.Id)
		fmt.Println(friend.Id)
		if err != nil {
			return friends, err
		}

		friends = append(friends, *friend)
	}

	return friends, nil
}
