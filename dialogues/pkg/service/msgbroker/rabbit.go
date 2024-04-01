package msgbroker

import (
	"dialogues/pkg/domain"
	"dialogues/pkg/service"
	"encoding/json"
	"fmt"
	"github.com/streadway/amqp"
	"log"
	"sync"
)

// WaitGroupWrapper is a wrapper around sync.WaitGroup
type WaitGroupWrapper struct {
	sync.WaitGroup
}

// MsgBroker represents a message broker service
type MsgBroker struct {
	conn            *amqp.Connection
	ch              *amqp.Channel
	dialogueService *service.DialogueService
	wg              WaitGroupWrapper
}

// NewMsgBroker initializes a new MsgBroker instance
func NewMsgBroker(connString string, dialogueService *service.DialogueService) *MsgBroker {
	conn, err := amqp.Dial(connString)
	if err != nil {
		log.Fatalf("Failed to connect to RabbitMQ: %s", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		log.Fatalf("Failed to open a channel: %s", err)
	}

	return &MsgBroker{
		conn:            conn,
		ch:              ch,
		dialogueService: dialogueService,
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
func (mb *MsgBroker) RunErrorIncrementConsumer(queueName string) {
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
			fmt.Printf("Received a message: %s\n", d.Body)
			unreadMessage := domain.MessageReadBroker{}
			err := json.Unmarshal(d.Body, &unreadMessage)
			if err != nil {
				log.Println("")
			}

			mb.dialogueService.UndoDecrement(unreadMessage.From, unreadMessage.To, unreadMessage.ReadCounter)
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
