package ws

import (
	"github.com/gorilla/websocket"
	"net/http"
	"time"
)

const (
	pongWait        = 60 * time.Second
	pingPeriod      = (pongWait * 9) / 10
	writeWait       = 10 * time.Second
	maxMessageSize  = 512
	readBufferSize  = 1024
	writeBufferSize = 1024
)

// TODO  thats handler not server so

var server = &WsHandler{
	Subscriptions: make(Subscription),
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  readBufferSize,
	WriteBufferSize: writeBufferSize,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}
