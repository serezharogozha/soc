package domain

type Dialogue struct {
	From int    `json:"from"`
	To   int    `json:"to"`
	Text string `json:"text"`
}

type TextDialogue struct {
	Text string `json:"text"`
}

type DialogueMessage struct {
	UserID   int    `json:"user_id"`
	ToUserID int    `json:"to_user_id"`
	Text     string `json:"text"`
}

type MessageReadBroker struct {
	From        int    `json:"from"`
	To          int    `json:"to"`
	ReadCounter uint64 `json:"ReadCounter"`
}

type MessageSendBroker struct {
	From int `json:"from"`
	To   int `json:"to"`
}
