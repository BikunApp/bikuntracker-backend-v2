package models

type User struct {
	Id        uint   `json:"id"`
	Name      string `json:"name"`
	Npm       string `json:"npm"`
	Email     string `json:"email"`
	Role      string `json:"role"`
	CreatedAt uint   `json:"created_at"`
	UpdatedAt uint   `json:"updated_at"`
}

func (u *User) IsAdmin() bool {
	return u.Role == "admin"
}
