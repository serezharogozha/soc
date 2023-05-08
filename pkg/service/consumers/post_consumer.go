package consumers

import (
	"awesomeProject10/pkg/config"
	"awesomeProject10/pkg/domain"
	"awesomeProject10/pkg/repository"
	"encoding/json"
	"fmt"
	"github.com/streadway/amqp"
	"strconv"
)

func Consume(postRepository repository.PostRepository, rabbitConf config.RabbitConf) {
	conn, err := amqp.Dial("amqp://" + rabbitConf.User + ":" + rabbitConf.Password + "@" + rabbitConf.Host + ":" + rabbitConf.Port + "/")
	if err != nil {
		fmt.Println("Failed Initializing Broker Connection")
		fmt.Println(err)
	}

	ch, err := conn.Channel()
	if err != nil {
		fmt.Println(err)
	}
	defer ch.Close()

	if err != nil {
		fmt.Println(err)
	}

	msgs, err := ch.Consume(
		"posts",
		"",
		true,
		false,
		false,
		false,
		nil,
	)

	forever := make(chan bool)
	go func() {
		for d := range msgs {
			post := domain.Post{}

			err := json.Unmarshal(d.Body, &post)

			err = addToCache(postRepository, post)
			if err != nil {
				fmt.Println(err)
			}
		}
	}()

	fmt.Println("Successfully Connected to our RabbitMQ Instance")
	fmt.Println(" [*] - Waiting for messages")
	<-forever
}

func addToCache(postRepository repository.PostRepository, post domain.Post) error {
	friendsOfUser, err := postRepository.GetFriendsOfUser(post.UserId)

	if err != nil {
		fmt.Println(err)
		return err
	}

	for _, friend := range friendsOfUser {
		UserIdStr := strconv.FormatInt(int64(friend.Id), 10)

		postJson, err := json.Marshal(post)
		if err != nil {
			fmt.Println(err)
			return err
		}

		err = postRepository.Cache.Add("feed:"+UserIdStr, string(postJson))
		if err != nil {
			fmt.Println(err)
			return err
		}
	}

	return nil
}
