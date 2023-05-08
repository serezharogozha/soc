package queue

import (
	"awesomeProject10/pkg/config"
	"fmt"
	"github.com/streadway/amqp"
)

func CreateChanel(rabbitConf config.RabbitConf) *amqp.Channel {
	conn, err := amqp.Dial("amqp://" + rabbitConf.User + ":" + rabbitConf.Password + "@" + rabbitConf.Host + ":" + rabbitConf.Port + "/")
	if err != nil {
		fmt.Println("Failed Initializing Broker Connection")
		fmt.Println(err)
	}
	ch, err := conn.Channel()
	if err != nil {
		fmt.Println(err)
	}

	_, err = ch.QueueDeclare(
		"posts_create",
		false,
		false,
		false,
		false,
		nil,
	)

	return ch
}
