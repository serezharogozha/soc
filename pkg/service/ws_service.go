package service

import (
	"github.com/go-redis/redis"
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

var Upgrader = websocket.Upgrader{
	ReadBufferSize:  readBufferSize,
	WriteBufferSize: writeBufferSize,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type WsService struct {
	Connections map[string]*websocket.Conn
	RedisClient *redis.Client
	mutex       sync.Mutex
}

func NewWsService(redisClient *redis.Client) *WsService {
	return &WsService{
		Connections: make(map[string]*websocket.Conn),
		RedisClient: redisClient,
	}
}

func (s *WsService) Send(conn *websocket.Conn, message string) {
	err := conn.WriteMessage(websocket.TextMessage, []byte(message))
	if err != nil {
		return
	}
}

func (s *WsService) Subscribe(conn *websocket.Conn, clientID string, userID string) {
	s.mutex.Lock()
	s.Connections[clientID] = conn
	s.mutex.Unlock()

	s.RedisClient.SAdd("subscriptions:"+userID, clientID)
}

func (s *WsService) Unsubscribe(clientID string, userID string) {
	s.mutex.Lock()
	delete(s.Connections, clientID)
	s.mutex.Unlock()

	s.RedisClient.SRem("subscriptions:"+userID, clientID)
}

func (s *WsService) Publish(userID string, message []byte) error {
	clientIDs, err := s.RedisClient.SMembers("subscriptions:" + userID).Result()
	if err != nil {
		return err
	}

	for _, clientID := range clientIDs {
		s.mutex.Lock()
		conn, ok := s.Connections[clientID]
		s.mutex.Unlock()
		if ok {
			s.Send(conn, string(message))
		}
	}

	return nil
}

func (s *WsService) ProcessMessage(conn *websocket.Conn, clientID string, userID string) {
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
			go s.Subscribe(conn, clientID, userID)
		case unsubscribe:
			go s.Unsubscribe(clientID, userID)
		default:
			go s.Send(conn, errActionUnrecognizable)
		}
	}
}
