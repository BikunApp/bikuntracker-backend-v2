package models

type Report struct {
	Id          uint   `json:"id"`
	UserId      uint   `json:"user_id"`
	Description string `json:"description"`
	Location    string `json:"location"`
	OccuredAt   uint   `json:"occured_at"`
	CreatedAt   uint   `json:"created_at"`
	UpdatedAt   uint   `json:"updated_at"`
}
