package domain

type DialogueMessage struct {
	UserID   int    `json:"user_id"`
	ToUserID int    `json:"to_user_id"`
	Text     string `json:"text"`
}
