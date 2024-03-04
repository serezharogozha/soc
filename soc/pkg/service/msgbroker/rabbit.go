package msgbroker

import (
	"encoding/json"
	"fmt"
	"github.com/streadway/amqp"
	"log"
	"soc/pkg/domain"
	"soc/pkg/service"
	"strconv"
	"sync"
)

// WaitGroupWrapper is a wrapper around sync.WaitGroup
type WaitGroupWrapper struct {
	sync.WaitGroup
}

// MsgBroker represents a message broker service
type MsgBroker struct {
	conn      *amqp.Connection
	ch        *amqp.Channel
	wg        WaitGroupWrapper
	wsHandler *service.WsService
}

// NewMsgBroker initializes a new MsgBroker instance
func NewMsgBroker(connString string, wsHandler *service.WsService) *MsgBroker {
	conn, err := amqp.Dial(connString)
	if err != nil {
		log.Fatalf("Failed to connect to RabbitMQ: %s", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		log.Fatalf("Failed to open a channel: %s", err)
	}

	return &MsgBroker{
		conn:      conn,
		ch:        ch,
		wsHandler: wsHandler,
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

func (mb *MsgBroker) Publish(queueName string, message string) error {
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
		return err
	}

	return nil
}

// RunConsumer runs the message consumer
func (mb *MsgBroker) RunConsumer(queueName string, service service.PostService) {
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
			post := domain.Post{}
			err := json.Unmarshal(d.Body, &post)
			if err != nil {
				fmt.Println(err)
			}

			friends, err := service.GetFriendToPublish(post)
			if err != nil {
				fmt.Println(err)
			}

			err = service.PublishPostToCache(post, friends)
			if err != nil {
				fmt.Println(err)
			}

			for _, friend := range friends {

				wsMessage := domain.PostWs{
					PostId:       strconv.Itoa(post.Id),
					PostText:     post.Text,
					AuthorUserId: strconv.Itoa(post.UserId),
				}

				wsMessageJson, err := json.Marshal(wsMessage)
				if err != nil {
					fmt.Println(err)
				}

				err = mb.wsHandler.Publish(strconv.Itoa(friend.Id), wsMessageJson)
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
