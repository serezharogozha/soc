package domain

type MessageReadBroker struct {
	From        int    `json:"from"`
	To          int    `json:"to"`
	ReadCounter uint64 `json:"ReadCounter"`
}

type MessageSendBroker struct {
	From int `json:"from"`
	To   int `json:"to"`
}
