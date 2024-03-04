package domain

type User struct {
	Id        int    `json:"id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Age       int    `json:"age"`
	Biography string `json:"biography"`
	City      string `json:"city"`
	Password  string `json:"password"`
}

type UserSafe struct {
	Id        int    `json:"id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Age       int    `json:"age"`
	City      string `json:"city"`
}

type Login struct {
	Id       int    `json:"id"`
	Password string `json:"password"`
}

type Search struct {
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
}
