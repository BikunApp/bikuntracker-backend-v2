package models

type User struct {
	Id        uint   `json:"id"`
	Name      string `json:"name"`
	Npm       string `json:"npm"`
	Email     string `json:"email"`
	CreatedAt uint   `json:"created_at"`
	UpdatedAt uint   `json:"updated_at"`
}
