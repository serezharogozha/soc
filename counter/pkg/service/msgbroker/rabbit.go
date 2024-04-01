package msgbroker

import (
	"counter/pkg/domain"
	"counter/pkg/service"
	"encoding/json"
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
	conn           *amqp.Connection
	ch             *amqp.Channel
	counterService *service.CounterService
	wg             WaitGroupWrapper
}

// NewMsgBroker initializes a new MsgBroker instance
func NewMsgBroker(connString string, counterService *service.CounterService) *MsgBroker {
	conn, err := amqp.Dial(connString)
	if err != nil {
		log.Fatalf("Failed to connect to RabbitMQ: %s", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		log.Fatalf("Failed to open a channel: %s", err)
	}

	return &MsgBroker{
		conn:           conn,
		ch:             ch,
		counterService: counterService,
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

// RunSendConsumer runs the send message consumer
func (mb *MsgBroker) RunSendConsumer(queueName string) {
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
			log.Printf("Received a message: %s", d.Body)
			sendMessage := domain.MessageSendBroker{}
			err := json.Unmarshal(d.Body, &sendMessage)
			if err != nil {
				return
			}

			err = mb.counterService.IncrementCounter(sendMessage.To, sendMessage.From)
			if err != nil {
				return
			}
		}
	})
}

// RunReadConsumer runs the read message consumer
func (mb *MsgBroker) RunReadConsumer(queueName string) {
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
			log.Printf("Received a message: %s", d.Body)
			readMessage := domain.MessageReadBroker{}
			err := json.Unmarshal(d.Body, &readMessage)

			if err != nil {
				err := mb.Publish("message_read_decrement_error", string(d.Body))
				if err != nil {
					return
				}
			}

			err = mb.counterService.DecrementCounter(readMessage.To, readMessage.From, readMessage.ReadCounter)
			if err != nil {
				err := mb.Publish("message_read_decrement_error", string(d.Body))
				if err != nil {
					return
				}
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
