package interfaces

import "github.com/FreeJ1nG/bikuntracker-backend/app/models"

type FavoriteRepository interface {
	SetFavoriteStop(user *models.User) (err error)
}
