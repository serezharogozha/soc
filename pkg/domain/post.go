package domain

type Post struct {
	Id     int    `json:"id"`
	Text   string `json:"text"`
	UserId int    `json:"user_id"`
}

type PostWs struct {
	PostId       string `json:"postId"`
	PostText     string `json:"postText"`
	AuthorUserId string `json:"author_user_id"`
}

type Posts []Post

type PostFeed struct {
	Posts `json:"posts"`
}
