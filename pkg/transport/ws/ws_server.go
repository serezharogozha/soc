package ws

import (
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

const (
	subscribe   = "subscribe"
	unsubscribe = "unsubscribe"
)

const (
	errInvalidMessage       = "Server: Invalid msg"
	errActionUnrecognizable = "Server: Action unrecognized"
)

type WsMessage struct {
	Event string          `json:"event"`
	Data  json.RawMessage `json:"data"`
}

var WsPoolMutex sync.Mutex
var wsConns []*websocket.Conn

type Subscription map[string]Client

type Client map[string]*websocket.Conn

type Message struct {
	Action  string `json:"action"`
	Topic   string `json:"topic"`
	Message string `json:"message"`
}

type WsHandler struct {
	Subscriptions Subscription
}

func (s WsHandler) HandleWS(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("failed upgrading connection"))
		return
	}
	defer conn.Close()
	clientID := uuid.New().String()

	s.ProcessMessage(conn, clientID)
}

func (s WsHandler) Send(conn *websocket.Conn, message string) {
	err := conn.WriteMessage(websocket.TextMessage, []byte(message))
	if err != nil {
		return
	}
}

func (s WsHandler) SendWithWait(conn *websocket.Conn, message string, wg *sync.WaitGroup) {
	err := conn.WriteMessage(websocket.TextMessage, []byte(message))
	if err != nil {
		return
	}
	wg.Done()
}

func (s WsHandler) RemoveClient(clientID string) {
	for _, client := range s.Subscriptions {
		delete(client, clientID)
	}
}

func (s WsHandler) ProcessMessage(conn *websocket.Conn, clientID string) {
	WsPoolMutex.Lock()
	wsConns = append(wsConns, conn)
	WsPoolMutex.Unlock()

	for {
		var msg WsMessage
		err := conn.ReadJSON(&msg)
		if err != nil {
			s.Send(conn, errInvalidMessage)
			conn.Close()
			break
		}

		fmt.Println("Processing message: " + string(msg.Data))

		switch msg.Event {
		case subscribe:
			go s.Subscribe(conn, clientID, string(msg.Data)) // todo topoic
		case unsubscribe:
			go s.Unsubscribe(clientID, string(msg.Data))
		default:
			go s.Send(conn, errActionUnrecognizable)
		}
	}
}

func (s WsHandler) Publish(topic string, message []byte) (err error) {
	fmt.Println("Publishing to topic: " + topic)
	if _, exist := s.Subscriptions[topic]; !exist {
		fmt.Println("No client subscribed to topic: " + topic)
		return
	}
	fmt.Println("Sending to clients")

	client := s.Subscriptions[topic]

	var wg sync.WaitGroup
	for _, conn := range client {
		wg.Add(1)
		fmt.Println("Sending to client")
		go s.SendWithWait(conn, string(message), &wg)
	}

	wg.Wait()
	return
}

func (s WsHandler) Subscribe(conn *websocket.Conn, clientID string, topic string) {
	if _, exist := s.Subscriptions[topic]; exist {
		client := s.Subscriptions[topic]

		if _, subbed := client[clientID]; subbed {
			return
		}

		client[clientID] = conn
		return
	}

	fmt.Println("Creating new subscription")
	newClient := make(Client)
	s.Subscriptions[topic] = newClient
	s.Subscriptions[topic][clientID] = conn
}

func (s WsHandler) Unsubscribe(clientID string, topic string) {
	if _, exist := s.Subscriptions[topic]; exist {
		client := s.Subscriptions[topic]

		delete(client, clientID)
	}
}

func NewWsHandler() *WsHandler {
	s := WsHandler{Subscriptions: Subscription{}}

	return &s
}
