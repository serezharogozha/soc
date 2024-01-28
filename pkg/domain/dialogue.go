package domain

type Dialogue struct {
	From int    `json:"from"`
	To   int    `json:"to"`
	Text string `json:"text"`
}

type TextDialogue struct {
	Text string `json:"text"`
}
