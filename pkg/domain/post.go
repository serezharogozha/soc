package domain

type Post struct {
	Id     int    `json:"id"`
	Text   string `json:"text"`
	UserId int    `json:"user_id"`
}

type Posts []Post

type PostFeed struct {
	Posts `json:"posts"`
}
