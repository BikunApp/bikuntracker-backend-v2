package auth

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/FreeJ1nG/bikuntracker-backend/app/models"
	"github.com/FreeJ1nG/bikuntracker-backend/utils"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type util struct {
	config *utils.Config
}

func NewUtil(config *utils.Config) *util {
	return &util{
		config: config,
	}
}

func (u *util) GenerateTokenPair(user *models.User) (signedJwtToken string, signedRefreshToken string, err error) {
	now := time.Now()
	jwtExpiryInDays := u.config.JwtExpiryInDays
	token := jwt.NewWithClaims(
		jwt.SigningMethodHS256,
		models.JwtClaims{
			TokenType: "access",
			RegisteredClaims: jwt.RegisteredClaims{
				Issuer:    "bikun-tracker",
				IssuedAt:  jwt.NewNumericDate(now),
				ExpiresAt: jwt.NewNumericDate(now.Add(time.Duration(jwtExpiryInDays) * 24 * time.Hour)),
				Subject:   user.Npm,
			},
		},
	)

	signedJwtToken, err = token.SignedString([]byte(u.config.JwtSecretKey))
	if err != nil {
		err = fmt.Errorf("unable to sign access token: %w", err)
		return
	}

	refreshExpiryInDays := u.config.JwtRefreshExpiryInDays
	refreshToken := jwt.NewWithClaims(
		jwt.SigningMethodHS256,
		models.JwtClaims{
			TokenType: "refresh",
			RegisteredClaims: jwt.RegisteredClaims{
				Issuer:    "bikun-tracker",
				IssuedAt:  jwt.NewNumericDate(now),
				ExpiresAt: jwt.NewNumericDate(now.Add(time.Duration(refreshExpiryInDays) * 24 * time.Hour)),
				Subject:   user.Npm,
			},
		},
	)

	signedRefreshToken, err = refreshToken.SignedString([]byte(u.config.JwtSecretKey))
	if err != nil {
		err = fmt.Errorf("unable to sign refresh token: %w", err)
		return
	}

	return
}

func (u *util) HashPassword(password string) (passwordHash string, err error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return
	}
	passwordHash = string(hashedPassword)
	return
}

func (u *util) ExtractJwtToken(r *http.Request) (jwtToken string, err error) {
	authorization := r.Header.Get("Authorization")
	authSplit := strings.Split(authorization, " ")
	if len(authSplit) != 2 {
		err = fmt.Errorf("invalid authorization header format")
		return
	}

	prefix := authSplit[0]
	tokenString := authSplit[1]
	if prefix != "Bearer" {
		err = fmt.Errorf("jwt token not found on authorization header")
		return
	}

	jwtToken = tokenString
	return
}

func (u *util) ToJwtToken(tokenString string) (token *jwt.Token, err error) {
	token, err = jwt.Parse(tokenString, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %s", t.Method)
		}
		return []byte(u.config.JwtSecretKey), nil
	})

	if err != nil {
		err = fmt.Errorf("unable to parse token: %w", err)
		return
	}

	return
}
