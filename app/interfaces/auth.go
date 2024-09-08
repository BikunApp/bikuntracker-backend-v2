package interfaces

import (
	"net/http"

	"github.com/FreeJ1nG/bikuntracker-backend/app/models"
	"github.com/golang-jwt/jwt/v5"
)

type AuthUtil interface {
	GenerateTokenPair(user *models.User) (signedJwtToken string, signedRefreshToken string, err error)
	HashPassword(password string) (passwordHash string, err error)
	ExtractJwtToken(r *http.Request) (jwtToken string, err error)
	ToJwtToken(tokenString string) (token *jwt.Token, err error)
}

type AuthRepository interface {
	GetUser(npm string) (res models.User, err error)
	GetOrCreateUser(name, npm, email string) (res models.User, err error)
}

type AuthService interface {
	SSOLogin(ticket, serviceName string) (user models.User, err error)
}
