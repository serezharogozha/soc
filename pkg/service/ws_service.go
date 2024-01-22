package service

import (
	"fmt"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

const (
	subscribe               = "subscribe"
	unsubscribe             = "unsubscribe"
	errInvalidMessage       = "Server: Invalid msg"
	errActionUnrecognizable = "Server: Action unrecognized"
	readBufferSize          = 1024
	writeBufferSize         = 1024
)

type WsMessage struct {
	Event string `json:"event"`
}

type Client struct {
	ID   string
	Conn *websocket.Conn
}

type SubscriptionMap struct {
	sync.RWMutex
	Subscribers map[string]map[string]*Client
}

type WsService struct {
	Subscriptions SubscriptionMap
}

type Message struct {
	Action  string `json:"action"`
	Topic   string `json:"topic"`
	Message string `json:"message"`
}

var Upgrader = websocket.Upgrader{
	ReadBufferSize:  readBufferSize,
	WriteBufferSize: writeBufferSize,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func (s *WsService) Send(conn *websocket.Conn, message string) {
	err := conn.WriteMessage(websocket.TextMessage, []byte(message))
	if err != nil {
		return
	}
}

func (s *WsService) ProcessMessage(conn *websocket.Conn, clientID string, userId string) {
	for {
		var msg WsMessage
		err := conn.ReadJSON(&msg)
		if err != nil {
			s.Send(conn, errInvalidMessage)
			conn.Close()
			break
		}

		switch msg.Event {
		case subscribe:
			fmt.Println("Subscribing user: ", userId)
			go s.Subscribe(conn, clientID, userId)
		case unsubscribe:
			go s.Unsubscribe(clientID, userId)
		default:
			go s.Send(conn, errActionUnrecognizable)
		}
	}
}

func (s *WsService) Publish(userId string, message []byte) error {
	fmt.Println("Publishing message to user: ", userId)
	s.Subscriptions.RLock()
	clients, ok := s.Subscriptions.Subscribers[userId]
	s.Subscriptions.RUnlock()

	if !ok {
		return nil
	}

	var wg sync.WaitGroup
	for _, client := range clients {
		wg.Add(1)
		go func(conn *websocket.Conn) {
			defer wg.Done()
			if conn != nil {
				s.Send(conn, string(message))
			}
		}(client.Conn)
	}

	wg.Wait()
	return nil
}

func (s *WsService) Subscribe(conn *websocket.Conn, clientID string, userId string) {
	client := &Client{ID: clientID, Conn: conn}

	s.Subscriptions.Lock()
	if s.Subscriptions.Subscribers[userId] == nil {
		s.Subscriptions.Subscribers[userId] = make(map[string]*Client)
	}
	s.Subscriptions.Subscribers[userId][clientID] = client
	s.Subscriptions.Unlock()
}

func (s *WsService) Unsubscribe(clientID string, userId string) {
	s.Subscriptions.Lock()
	if clients, ok := s.Subscriptions.Subscribers[userId]; ok {
		delete(clients, clientID)
	}
	s.Subscriptions.Unlock()
}

func NewWsService() *WsService {
	return &WsService{
		Subscriptions: SubscriptionMap{
			Subscribers: make(map[string]map[string]*Client),
		},
	}
}
