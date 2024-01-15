package msgbroker

import (
	"encoding/json"
	"fmt"
	"github.com/streadway/amqp"
	"log"
	"soc/pkg/domain"
	"soc/pkg/service"
	"sync"
)

// WaitGroupWrapper is a wrapper around sync.WaitGroup
type WaitGroupWrapper struct {
	sync.WaitGroup
}

// MsgBroker represents a message broker service
type MsgBroker struct {
	conn *amqp.Connection
	ch   *amqp.Channel
	wg   WaitGroupWrapper
}

// NewMsgBroker initializes a new MsgBroker instance
func NewMsgBroker(connString string) *MsgBroker {
	fmt.Println(connString)
	conn, err := amqp.Dial(connString)
	if err != nil {
		log.Fatalf("Failed to connect to RabbitMQ: %s", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		log.Fatalf("Failed to open a channel: %s", err)
	}

	return &MsgBroker{
		conn: conn,
		ch:   ch,
	}
}

// InitQueue declares a queue and returns its details
func (mb *MsgBroker) InitQueue(queueName string) amqp.Queue {
	q, err := mb.ch.QueueDeclare(
		queueName,
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		log.Fatalf("Failed to declare a queue: %s", err)
	}

	return q
}

// RunProducer runs the message producer
//func (mb *MsgBroker) RunProducer(queueName string)

// TODO: MAYBE THAT
func (mb *MsgBroker) RunProducer(queueName string, inputData chan string) {

	mb.wg.Wrap(func() {
		for {
			message, ok := <-inputData
			if !ok {
				return
			}

			err := mb.ch.Publish(
				"",
				queueName,
				false,
				false,
				amqp.Publishing{
					ContentType: "text/plain",
					Body:        []byte(message),
				},
			)
			if err != nil {
				log.Printf("Failed to publish a message: %s", err)
			}

			fmt.Printf("Sent: %s\n", message)
		}
	})
}

// RunConsumer runs the message consumer
func (mb *MsgBroker) RunConsumer(queueName string) {
	msgs, err := mb.ch.Consume(
		queueName,
		"",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		log.Fatalf("Failed to register a consumer: %s", err)
	}

	mb.wg.Wrap(func() {
		for d := range msgs {
			fmt.Println("Received a message: " + string(d.Body))
			post := domain.Post{}
			err := json.Unmarshal(d.Body, &post)
			if err != nil {
				fmt.Println(err)
			}

			friends, err := service.PostService.GetFriendToPublish(post)
			if err != nil {
				fmt.Println(err)
			}

			err = service.postService.PublishPostToCache(post, friends)
			if err != nil {
				fmt.Println(err)
			}
		}
	})
}

// Close closes the MsgBroker connections
func (mb *MsgBroker) Close() {
	err := mb.conn.Close()
	if err != nil {
		return
	}
	mb.wg.Wait()
}

func (w *WaitGroupWrapper) Wrap(f func()) {
	w.Add(1)
	go func() {
		defer w.Done()
		f()
	}()
}
