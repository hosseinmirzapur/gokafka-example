package models

type User struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type Notification struct {
	From    string `json:"from"`
	To      string `json:"to"`
	Message string `json:"message"`
}
